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
	got := b.URL("a.png", "")
	if got != "https://pics.garage.example.com/a.png" {
		t.Errorf("expected https URL, got %q", got)
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
	got := b.URL("abcd.svg", "")
	const want = "http://profile-pictures.garage.test/abcd.svg"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestURLBuilder_URL_HostWithPort(t *testing.T) {
	b, err := NewURLBuilder("http://garage.test:3900", "pics")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	got := b.URL("key.png", "")
	const want = "http://pics.garage.test:3900/key.png"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestURLBuilder_URL_PreservesUUIDStyleKey(t *testing.T) {
	b, _ := NewURLBuilder("http://garage.test", "profile-pictures")
	key := "11111111-2222-3333-4444-555555555555.png"
	got := b.URL(key, "")
	if want := "http://profile-pictures.garage.test/" + key; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestURLBuilder_URL_EmptyKey_ReturnsFallback(t *testing.T) {
	b, _ := NewURLBuilder("http://garage.test", "pics")
	const fallback = "http://profile-pictures.garage.test/default.svg"
	if got := b.URL("", fallback); got != fallback {
		t.Errorf("expected fallback for empty key, got %q", got)
	}
}

func TestURLBuilder_URL_WhitespaceKey_ReturnsFallback(t *testing.T) {
	b, _ := NewURLBuilder("http://garage.test", "pics")
	const fallback = "http://profile-pictures.garage.test/default.svg"
	if got := b.URL("   ", fallback); got != fallback {
		t.Errorf("expected fallback for whitespace-only key, got %q", got)
	}
	if got := b.URL("\t\n", fallback); got != fallback {
		t.Errorf("expected fallback for whitespace-only key, got %q", got)
	}
}

func TestURLBuilder_URL_NilReceiver_ReturnsFallback(t *testing.T) {
	// The production code defensively guards against a nil *URLBuilder
	// receiver. This test pins that contract so callers can rely on it.
	var b *URLBuilder
	const fallback = "http://profile-pictures.garage.test/default.svg"
	if got := b.URL("anything.png", fallback); got != fallback {
		t.Errorf("expected fallback URL from nil builder, got %q", got)
	}
}
