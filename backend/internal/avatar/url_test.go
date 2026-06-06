package avatar

import (
	"testing"
)

func TestNewURLBuilder_ValidHTTP(t *testing.T) {
	b, err := NewURLBuilder("http://garage.test", "profile-pictures")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestNewURLBuilder_ValidHTTPS(t *testing.T) {
	b, err := NewURLBuilder("https://garage.example.com", "pics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := b.URL("a.png")
	if got == nil {
		t.Fatalf("expected non-nil URL for non-empty key")
	}
	if *got != "https://pics.garage.example.com/a.png" {
		t.Errorf("expected https URL, got %q", *got)
	}
}

func TestNewURLBuilder_InvalidURL(t *testing.T) {
	// url.Parse rejects control bytes in URLs.
	_, err := NewURLBuilder("http://exa\x00mple.com", "pics")
	if err == nil {
		t.Fatal("expected error for malformed URL")
	}
}

func TestURLBuilder_URL_BasicHTTP(t *testing.T) {
	b, err := NewURLBuilder("http://garage.test", "profile-pictures")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	got := b.URL("abcd.svg")
	if got == nil {
		t.Fatal("expected non-nil URL")
	}
	const want = "http://profile-pictures.garage.test/abcd.svg"
	if *got != want {
		t.Errorf("expected %q, got %q", want, *got)
	}
}

func TestURLBuilder_URL_HostWithPort(t *testing.T) {
	b, err := NewURLBuilder("http://garage.test:3900", "pics")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	got := b.URL("key.png")
	if got == nil {
		t.Fatal("expected non-nil URL")
	}
	const want = "http://pics.garage.test:3900/key.png"
	if *got != want {
		t.Errorf("expected %q, got %q", want, *got)
	}
}

func TestURLBuilder_URL_PreservesUUIDStyleKey(t *testing.T) {
	b, _ := NewURLBuilder("http://garage.test", "profile-pictures")
	key := "11111111-2222-3333-4444-555555555555.png"
	got := b.URL(key)
	if got == nil {
		t.Fatal("expected non-nil URL")
	}
	if want := "http://profile-pictures.garage.test/" + key; *got != want {
		t.Errorf("expected %q, got %q", want, *got)
	}
}

func TestURLBuilder_URL_EmptyKey_ReturnsNil(t *testing.T) {
	b, _ := NewURLBuilder("http://garage.test", "pics")
	if got := b.URL(""); got != nil {
		t.Errorf("expected nil for empty key, got %q", *got)
	}
}

func TestURLBuilder_URL_WhitespaceKey_ReturnsNil(t *testing.T) {
	b, _ := NewURLBuilder("http://garage.test", "pics")
	if got := b.URL("   "); got != nil {
		t.Errorf("expected nil for whitespace-only key, got %q", *got)
	}
	if got := b.URL("\t\n"); got != nil {
		t.Errorf("expected nil for whitespace-only key, got %q", *got)
	}
}

func TestURLBuilder_URL_NilReceiver_ReturnsNil(t *testing.T) {
	// The production code defensively guards against a nil *URLBuilder
	// receiver. This test pins that contract so callers can rely on it.
	var b *URLBuilder
	if got := b.URL("anything.png"); got != nil {
		t.Errorf("expected nil URL from nil builder, got %q", *got)
	}
}
