package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUUIDParam_UnmarshalJSON_Valid(t *testing.T) {
	const id = "da78a8e3-506f-4e79-a50c-75c2b95156cc"
	raw := []byte(`"` + id + `"`)

	var u UUIDParam
	if err := json.Unmarshal(raw, &u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !u.Value.Valid {
		t.Fatalf("expected Value to be Valid")
	}
	if got := u.Value.String(); got != id {
		t.Errorf("expected UUID %q, got %q", id, got)
	}
}

func TestUUIDParam_UnmarshalJSON_InvalidString(t *testing.T) {
	var u UUIDParam
	err := json.Unmarshal([]byte(`"not-a-uuid"`), &u)
	if err == nil {
		t.Fatalf("expected error for invalid uuid string, got nil")
	}
}

func TestUUIDParam_UnmarshalJSON_NonString(t *testing.T) {
	var u UUIDParam
	err := json.Unmarshal([]byte(`12345`), &u)
	if err == nil {
		t.Fatalf("expected error when unmarshalling non-string into UUIDParam")
	}
}

func TestUUIDParam_MarshalJSON_Valid(t *testing.T) {
	const id = "19c5bdf1-e8a7-4d7d-bc49-f2583887907a"
	var u UUIDParam
	if err := u.Value.Scan(id); err != nil {
		t.Fatalf("failed to scan uuid: %v", err)
	}

	out, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `"` + id + `"`
	if string(out) != want {
		t.Errorf("expected %s, got %s", want, string(out))
	}
}

func TestUUIDParam_MarshalJSON_Zero(t *testing.T) {
	var u UUIDParam
	out, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != `""` {
		t.Errorf("expected empty string for zero UUID, got %s", string(out))
	}
}

func TestUUIDParam_RoundTripInsideStruct(t *testing.T) {
	type body struct {
		AuthorID UUIDParam `json:"author_id"`
		Title    string    `json:"title"`
	}

	const id = "d4908384-0571-4ae6-a2d8-192c572fec2b"
	in := []byte(`{"author_id":"` + id + `","title":"hi"}`)

	var b body
	if err := json.Unmarshal(in, &b); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !b.AuthorID.Value.Valid || b.AuthorID.Value.String() != id {
		t.Fatalf("AuthorID not parsed correctly: %+v", b.AuthorID)
	}
	if b.Title != "hi" {
		t.Errorf("expected title 'hi', got %q", b.Title)
	}

	out, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(out), `"author_id":"`+id+`"`) {
		t.Errorf("marshalled output missing author_id: %s", string(out))
	}
}
