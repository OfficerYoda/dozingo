package types

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDurationParam_UnmarshalJSON_Valid(t *testing.T) {
	raw := []byte(`"24h"`)

	var d DurationParam
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Value != 24*time.Hour {
		t.Errorf("expected Value=%v, got %v", 24*time.Hour, d.Value)
	}
}

func TestDurationParam_UnmarshalJSON_ValidComplex(t *testing.T) {
	raw := []byte(`"1h30m"`)

	var d DurationParam
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Hour + 30*time.Minute
	if d.Value != want {
		t.Errorf("expected Value=%v, got %v", want, d.Value)
	}
}

func TestDurationParam_UnmarshalJSON_InvalidString(t *testing.T) {
	var d DurationParam
	err := json.Unmarshal([]byte(`"notaduration"`), &d)
	if err == nil {
		t.Fatalf("expected error for invalid duration string, got nil")
	}
}

func TestDurationParam_UnmarshalJSON_NonString(t *testing.T) {
	var d DurationParam
	err := json.Unmarshal([]byte(`12345`), &d)
	if err == nil {
		t.Fatalf("expected error when unmarshalling non-string into DurationParam")
	}
}

func TestDurationParam_MarshalJSON(t *testing.T) {
	d := DurationParam{Value: 24 * time.Hour}

	out, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `"24h0m0s"`
	if string(out) != want {
		t.Errorf("expected %s, got %s", want, string(out))
	}
}

func TestDurationParam_MarshalJSON_Zero(t *testing.T) {
	var d DurationParam

	out, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `"0s"`
	if string(out) != want {
		t.Errorf("expected %s for zero duration, got %s", want, string(out))
	}
}

func TestDurationParam_OnParamSet_Set(t *testing.T) {
	var d DurationParam

	// Simulate Huma's parameter parsing: write into the receiver, then
	// invoke OnParamSet to trigger the parse.
	d.Receiver().SetString("1h30m")
	d.OnParamSet(true, nil)

	want := time.Hour + 30*time.Minute
	if d.Value != want {
		t.Errorf("expected Value=%v after OnParamSet, got %v", want, d.Value)
	}
}

func TestDurationParam_OnParamSet_NotSet(t *testing.T) {
	var d DurationParam
	d.OnParamSet(false, nil)
	if d.Value != 0 {
		t.Errorf("expected Value=0 when isSet=false, got %v", d.Value)
	}
}

func TestDurationParam_OnParamSet_InvalidString(t *testing.T) {
	var d DurationParam

	// Bad input: OnParamSet swallows the parse error and leaves Value
	// as the zero duration. This is intentional behavior; document it.
	d.Receiver().SetString("notaduration")
	d.OnParamSet(true, nil)

	if d.Value != 0 {
		t.Errorf("expected Value=0 for invalid input, got %v", d.Value)
	}
}

func TestDurationParam_RoundTripInsideStruct(t *testing.T) {
	type body struct {
		Window DurationParam `json:"window"`
		Title  string        `json:"title"`
	}

	in := []byte(`{"window":"2h","title":"hi"}`)

	var b body
	if err := json.Unmarshal(in, &b); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if b.Window.Value != 2*time.Hour {
		t.Fatalf("Window not parsed correctly: %+v", b.Window)
	}
	if b.Title != "hi" {
		t.Errorf("expected title 'hi', got %q", b.Title)
	}

	out, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(out), `"window":"2h0m0s"`) {
		t.Errorf("marshalled output missing window: %s", string(out))
	}
}
