package handler

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

/// ===== Multipart helpers =====

// buildMultipartAvatar builds a multipart form body with one file part for
// the avatar upload endpoint. If fieldName is empty, no `avatar` part is
// emitted; instead an unrelated text field is added so the body is still a
// valid multipart payload (used to exercise the missing-field branch).
//
// Returns the Content-Type header value (with boundary) and the raw body
// bytes ready to feed to httptest.NewRequest.
func buildMultipartAvatar(t *testing.T, fieldName, filename, contentType string, body []byte) (string, []byte) {
	t.Helper()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	if fieldName == "" {
		// Missing-field case: emit some other valid part so the multipart
		// body is well-formed but does not contain an `avatar` file.
		if err := mw.WriteField("not_avatar", "ignored"); err != nil {
			t.Fatalf("write filler field: %v", err)
		}
	} else {
		h := make(map[string][]string)
		h["Content-Disposition"] = []string{
			`form-data; name="` + fieldName + `"; filename="` + filename + `"`,
		}
		if contentType != "" {
			h["Content-Type"] = []string{contentType}
		}
		part, err := mw.CreatePart(h)
		if err != nil {
			t.Fatalf("create form part: %v", err)
		}
		if _, err := part.Write(body); err != nil {
			t.Fatalf("write part body: %v", err)
		}
	}

	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return mw.FormDataContentType(), buf.Bytes()
}

// doMultipartRequest mirrors doRequestWithCookies but sends a pre-built
// multipart body instead of a JSON one. The caller owns Content-Type
// (returned by buildMultipartAvatar) so the boundary lines up.
func doMultipartRequest(method, path, contentType string, body []byte, cookies []*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

/// ===== PUT /users/me/avatar =====

// avatarKeyURLRegex matches the avatar URL the test URLBuilder emits:
// http://profile-pictures.garage.test/<uuid>.png. The UUID is the random
// key the service generates; the suffix is the file extension preserved
// from the upload.
var avatarKeyURLRegex = regexp.MustCompile(`^http://profile-pictures\.garage\.test/[0-9a-f-]{36}\.png$`)

func TestUploadAvatar_NoCookie_401(t *testing.T) {
	setupTest(t)

	ct, body := buildMultipartAvatar(t, "avatar", "hello.png", "image/png", []byte("fake-png-bytes"))
	w := doMultipartRequest(http.MethodPut, "/api/users/me/avatar", ct, body, nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestUploadAvatar_AnonymousSessionOnly_401(t *testing.T) {
	setupTest(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)

	ct, body := buildMultipartAvatar(t, "avatar", "hello.png", "image/png", []byte("fake-png-bytes"))
	w := doMultipartRequest(http.MethodPut, "/api/users/me/avatar", ct, body, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestUploadAvatar_MissingAvatarField_422(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "missingfield", "mypassword123", nil)
	userID := (*resp)["user_id"].(string)

	// Register itself uploads one avatar (best-effort auto-generation),
	// so count uploads from this point forward.
	preCount := fakeUploader.uploadCount()

	// fieldName="" -> body is valid multipart but has no `avatar` part.
	ct, body := buildMultipartAvatar(t, "", "", "", nil)
	w := doMultipartRequest(http.MethodPut, "/api/users/me/avatar", ct, body, cookiesFor(userID))
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Errorf("expected 422/400 for missing avatar field, got %d (body: %s)", w.Code, w.Body.String())
	}

	// No additional upload should have been attempted.
	if got := fakeUploader.uploadCount() - preCount; got != 0 {
		t.Errorf("expected 0 additional uploads when avatar field is missing, got %d", got)
	}
}

func TestUploadAvatar_Success_PNG(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "pnguser", "mypassword123", stringPtr("png@example.com"))
	userID := (*resp)["user_id"].(string)

	// Register already auto-generated an avatar; record the count so we
	// can assert this upload was the second.
	preCount := fakeUploader.uploadCount()

	pngBytes := []byte("fake-png-bytes")
	ct, body := buildMultipartAvatar(t, "avatar", "hello.png", "image/png", pngBytes)
	w := doMultipartRequest(http.MethodPut, "/api/users/me/avatar", ct, body, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var got map[string]any
	decodeJSON(t, w, &got)

	assertJSONField(t, got, "user_id", userID)
	assertJSONField(t, got, "username", "pnguser")
	assertJSONField(t, got, "email", "png@example.com")

	urlStr, ok := got["avatar_url"].(string)
	if !ok {
		t.Fatalf("expected avatar_url string, got %T (%v)", got["avatar_url"], got["avatar_url"])
	}
	if !avatarKeyURLRegex.MatchString(urlStr) {
		t.Errorf("avatar_url %q does not match %s", urlStr, avatarKeyURLRegex)
	}

	// One additional upload (on top of the register-time one).
	if added := fakeUploader.uploadCount() - preCount; added != 1 {
		t.Fatalf("expected 1 additional upload from PUT /me/avatar, got %d", added)
	}
	last, ok := fakeUploader.lastUpload()
	if !ok {
		t.Fatal("expected a recorded upload")
	}
	if last.ContentType != "image/png" {
		t.Errorf("expected content type image/png, got %q", last.ContentType)
	}
	if last.Extension != ".png" {
		t.Errorf("expected extension .png, got %q", last.Extension)
	}
	if !strings.HasSuffix(last.Key, ".png") {
		t.Errorf("expected key to end in .png, got %q", last.Key)
	}
	if !bytes.Equal(last.Bytes, pngBytes) {
		t.Errorf("expected uploaded bytes %q, got %q", pngBytes, last.Bytes)
	}

	// DB row holds the same key the uploader recorded.
	user := loadUserByID(t, userID)
	if user.AvatarKey != last.Key {
		t.Errorf("expected DB AvatarKey %q, got %q", last.Key, user.AvatarKey)
	}

	// Response avatar_url is the URL builder's view of that key.
	expectedURL := "http://profile-pictures.garage.test/" + last.Key
	if urlStr != expectedURL {
		t.Errorf("expected avatar_url %q, got %q", expectedURL, urlStr)
	}
}

func TestUploadAvatar_Success_SVG_PreservesExtension(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "svguser", "mypassword123", stringPtr("svg@example.com"))
	userID := (*resp)["user_id"].(string)

	svgBytes := []byte("<svg></svg>")
	ct, body := buildMultipartAvatar(t, "avatar", "me.svg", "image/svg+xml", svgBytes)
	w := doMultipartRequest(http.MethodPut, "/api/users/me/avatar", ct, body, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	last, ok := fakeUploader.lastUpload()
	if !ok {
		t.Fatal("expected a recorded upload")
	}
	if last.Extension != ".svg" {
		t.Errorf("expected extension .svg, got %q", last.Extension)
	}
	if !strings.HasSuffix(last.Key, ".svg") {
		t.Errorf("expected key to end in .svg, got %q", last.Key)
	}
	if last.ContentType != "image/svg+xml" {
		t.Errorf("expected content type image/svg+xml, got %q", last.ContentType)
	}
}

func TestUploadAvatar_Success_OverwritesPreviousKey(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "overwriteuser", "mypassword123", nil)
	userID := (*resp)["user_id"].(string)

	// Register itself uploads one avatar, so count from this point on.
	preCount := fakeUploader.uploadCount()

	// First explicit upload.
	ct1, body1 := buildMultipartAvatar(t, "avatar", "first.png", "image/png", []byte("first-bytes"))
	w1 := doMultipartRequest(http.MethodPut, "/api/users/me/avatar", ct1, body1, cookiesFor(userID))
	assertStatus(t, w1, http.StatusOK)
	var got1 map[string]any
	decodeJSON(t, w1, &got1)
	url1, _ := got1["avatar_url"].(string)
	if url1 == "" {
		t.Fatalf("expected avatar_url on first upload, got %v", got1["avatar_url"])
	}

	// Second upload with different bytes -> service mints a fresh UUID
	// key, so the public URL must change.
	ct2, body2 := buildMultipartAvatar(t, "avatar", "second.png", "image/png", []byte("second-bytes"))
	w2 := doMultipartRequest(http.MethodPut, "/api/users/me/avatar", ct2, body2, cookiesFor(userID))
	assertStatus(t, w2, http.StatusOK)
	var got2 map[string]any
	decodeJSON(t, w2, &got2)
	url2, _ := got2["avatar_url"].(string)
	if url2 == "" {
		t.Fatalf("expected avatar_url on second upload, got %v", got2["avatar_url"])
	}

	if url1 == url2 {
		t.Errorf("expected second upload to change avatar_url; both were %q", url1)
	}

	if added := fakeUploader.uploadCount() - preCount; added != 2 {
		t.Fatalf("expected 2 additional uploads, got %d", added)
	}

	last, _ := fakeUploader.lastUpload()
	user := loadUserByID(t, userID)
	if user.AvatarKey != last.Key {
		t.Errorf("expected DB AvatarKey %q (second upload), got %q", last.Key, user.AvatarKey)
	}
}

func TestUploadAvatar_UploaderFails_500(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "uploaderfail", "mypassword123", nil)
	userID := (*resp)["user_id"].(string)

	// Register already auto-generated an avatar; capture the resulting
	// avatar_key so we can assert the failed manual upload doesn't touch
	// it.
	preKey := loadUserByID(t, userID).AvatarKey
	if preKey == "" {
		t.Fatal("expected register to have set an avatar_key (best-effort auto-generation)")
	}

	// Arm the fake to return an error on the next Upload call. The
	// service must surface this as a 5xx without leaking the underlying
	// error message to the client.
	fakeUploader.failNext = errors.New("garage down")

	ct, body := buildMultipartAvatar(t, "avatar", "hello.png", "image/png", []byte("fake-png-bytes"))
	w := doMultipartRequest(http.MethodPut, "/api/users/me/avatar", ct, body, cookiesFor(userID))

	if w.Code < 500 {
		t.Errorf("expected server error (>=500), got %d (body: %s)", w.Code, w.Body.String())
	}

	// Internal error message must not leak.
	if strings.Contains(w.Body.String(), "garage down") {
		t.Errorf("response body must not leak internal error, got: %s", w.Body.String())
	}

	// DB row unchanged: failed upload must not overwrite the user's
	// avatar_key (it should still be whatever register set).
	if got := loadUserByID(t, userID).AvatarKey; got != preKey {
		t.Errorf("expected AvatarKey to remain %q after failed upload, got %q", preKey, got)
	}
}

/// ===== avatar_url surfacing on /users/me =====

func TestMe_AvatarURL_AutoGeneratedAfterRegister(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "noavatarme", "mypassword123", stringPtr("noav@example.com"))
	userID := (*resp)["user_id"].(string)

	w := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var got map[string]any
	decodeJSON(t, w, &got)

	// Register's best-effort avatar generator picks up the username as
	// seed and uploads a UUID-keyed image; /me must reflect that fresh
	// key, not the column's "default.svg" placeholder.
	urlStr, ok := got["avatar_url"].(string)
	if !ok {
		t.Fatalf("expected avatar_url string from /me, got %T (%v)", got["avatar_url"], got["avatar_url"])
	}
	regKey := loadUserByID(t, userID).AvatarKey
	if regKey == "" {
		t.Fatal("expected user.AvatarKey to be set by register's best-effort avatar")
	}
	if regKey == "default.svg" {
		t.Errorf("expected register to overwrite the migration default 'default.svg', got %q", regKey)
	}
	expected := "http://profile-pictures.garage.test/" + regKey
	if urlStr != expected {
		t.Errorf("expected avatar_url %q, got %q", expected, urlStr)
	}
}

func TestMe_AvatarURL_PresentAfterUpload(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "havatarme", "mypassword123", stringPtr("hav@example.com"))
	userID := (*resp)["user_id"].(string)

	ct, body := buildMultipartAvatar(t, "avatar", "hello.png", "image/png", []byte("fake-png-bytes"))
	upload := doMultipartRequest(http.MethodPut, "/api/users/me/avatar", ct, body, cookiesFor(userID))
	assertStatus(t, upload, http.StatusOK)

	var uploadResp map[string]any
	decodeJSON(t, upload, &uploadResp)
	uploadedURL, _ := uploadResp["avatar_url"].(string)
	if uploadedURL == "" {
		t.Fatalf("expected non-empty avatar_url after upload, got %v", uploadResp["avatar_url"])
	}

	w := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var got map[string]any
	decodeJSON(t, w, &got)
	gotURL, _ := got["avatar_url"].(string)
	if gotURL != uploadedURL {
		t.Errorf("expected /me avatar_url %q to match upload response %q", gotURL, uploadedURL)
	}
}
