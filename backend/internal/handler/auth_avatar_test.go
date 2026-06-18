package handler

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

// ===== Auto-generated avatar on register =====

func TestRegister_GeneratesAvatar_HappyPath(t *testing.T) {
	setupTest(t)

	preUploads := fakeUploader.uploadCount()

	w := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "autoavatar",
		"password": "mypassword123",
		"email":    "autoavatar@example.com",
	})
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	// Generator was asked to produce an image. The seed is a random UUID
	// minted per registration (not the username), so we only assert it's
	// well-formed rather than equal to a fixed value.
	if c := fakeAvatarGen.seedCount(); c == 0 {
		t.Fatalf("expected avatar generator to be called at least once, got %d", c)
	}
	seed, ok := fakeAvatarGen.lastSeed()
	if !ok {
		t.Fatal("expected at least one recorded seed")
	}
	uuidRegex := regexp.MustCompile(`^[0-9a-f-]{36}$`)
	if !uuidRegex.MatchString(seed) {
		t.Errorf("expected generator seed to be a UUID, got %q", seed)
	}

	// Exactly one new upload (the generated avatar) was performed.
	if added := fakeUploader.uploadCount() - preUploads; added != 1 {
		t.Fatalf("expected 1 upload during register, got %d", added)
	}
	last, ok := fakeUploader.lastUpload()
	if !ok {
		t.Fatal("expected a recorded upload")
	}
	if last.Extension != ".svg" {
		t.Errorf("expected uploaded extension .svg, got %q", last.Extension)
	}
	if last.ContentType != "image/svg+xml" {
		t.Errorf("expected upload content type image/svg+xml, got %q", last.ContentType)
	}
	// The fake generator embeds the seed in the body so we can prove the
	// bytes flowed end-to-end. Recorded seed must match the bytes.
	if !strings.Contains(string(last.Bytes), `data-seed="`+seed+`"`) {
		t.Errorf("expected uploaded bytes to carry seed %q, got %q", seed, last.Bytes)
	}

	// Response avatar_url must point at the same key the uploader saw.
	urlStr, ok := resp["avatar_url"].(string)
	if !ok {
		t.Fatalf("expected avatar_url string on register, got %T (%v)", resp["avatar_url"], resp["avatar_url"])
	}
	expected := "http://profile-pictures.garage.test/" + last.Key
	if urlStr != expected {
		t.Errorf("expected register avatar_url %q, got %q", expected, urlStr)
	}

	// DB row carries the same key.
	userID := resp["user_id"].(string)
	if got := loadUserByID(t, userID).AvatarKey; got != last.Key {
		t.Errorf("expected user.avatar_key %q (matching uploader), got %q", last.Key, got)
	}
}

func TestRegister_GeneratorFails_StillSucceedsWithDefaultKey(t *testing.T) {
	setupTest(t)

	// Force the next Generate call to fail. Register must swallow the
	// error (best-effort) and still create the user.
	fakeAvatarGen.failNext = errors.New("dicebear unreachable")

	preUploads := fakeUploader.uploadCount()

	w := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "genfail",
		"password": "mypassword123",
		"email":    "genfail@example.com",
	})
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	// User exists.
	userID, ok := resp["user_id"].(string)
	if !ok || userID == "" {
		t.Fatalf("expected user_id on register, got %v", resp["user_id"])
	}
	if loadUserByID(t, userID).Username != "genfail" {
		t.Error("expected user row to be created despite avatar failure")
	}

	// No upload was attempted because Generate failed before Upload.
	if added := fakeUploader.uploadCount() - preUploads; added != 0 {
		t.Errorf("expected 0 uploads when generator fails, got %d", added)
	}

	// avatar_key was never overwritten, so it's the migration default.
	if got := loadUserByID(t, userID).AvatarKey; got != "default.svg" {
		t.Errorf("expected avatar_key to remain at migration default 'default.svg', got %q", got)
	}

	// avatar_url still resolves (URL builder runs against 'default.svg').
	if resp["avatar_url"] == nil {
		t.Error("expected avatar_url to fall back to default.svg URL, got nil")
	} else if urlStr, _ := resp["avatar_url"].(string); !strings.HasSuffix(urlStr, "/default.svg") {
		t.Errorf("expected avatar_url to point at /default.svg, got %q", urlStr)
	}
}

func TestRegister_UploaderFails_StillSucceedsWithDefaultKey(t *testing.T) {
	setupTest(t)

	// Generator works, but the uploader explodes. Register must still
	// succeed; the user row keeps the migration default avatar_key.
	fakeUploader.failNext = errors.New("garage explode")

	w := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "uploadfail",
		"password": "mypassword123",
		"email":    "uploadfail@example.com",
	})
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	userID := resp["user_id"].(string)
	if got := loadUserByID(t, userID).AvatarKey; got != "default.svg" {
		t.Errorf("expected avatar_key to remain 'default.svg' after upload failure, got %q", got)
	}

	// Generator was called (proving the failure was on the upload side).
	if fakeAvatarGen.seedCount() == 0 {
		t.Error("expected generator to be invoked before the upload")
	}

	// avatar_url still resolves to the default-asset URL.
	if urlStr, _ := resp["avatar_url"].(string); !strings.HasSuffix(urlStr, "/default.svg") {
		t.Errorf("expected avatar_url to fall back to /default.svg, got %q", urlStr)
	}
}

func TestRegister_GeneratorErrorDoesNotLeak(t *testing.T) {
	setupTest(t)

	fakeAvatarGen.failNext = errors.New("dicebear-internal-detail-shouldnt-leak")

	w := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "leaktest",
		"password": "mypassword123",
		"email":    "leaktest@example.com",
	})
	assertStatus(t, w, http.StatusOK)

	if strings.Contains(w.Body.String(), "dicebear-internal-detail-shouldnt-leak") {
		t.Errorf("response body must not leak internal generator error, got: %s", w.Body.String())
	}
}
