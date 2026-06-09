// Package types contains shared parameter and JSON wrapper types used by the API layer.
package types

import (
	"bytes"
	"encoding/json"

	"github.com/danielgtaylor/huma/v2"
)

// NullableString is a tri-state JSON string field for PATCH-style payloads:
//
//   - The key is omitted from the JSON object: Set is false, Value is nil.
//   - The key is present with a string value: Set is true, Value is &"...".
//   - The key is present with JSON null: Set is true, Value is nil.
//
// Use it on optional PATCH body fields where you need to distinguish
// "leave alone" from "clear".
type NullableString struct {
	Set   bool
	Value *string
}

// Schema declares this as a nullable string in the OpenAPI schema.
func (NullableString) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:     "string",
		Nullable: true,
	}
}

// UnmarshalJSON populates Set and Value from the raw bytes. Called only when
// the key is present in the input object; absent keys keep the zero value
// (Set=false), which is exactly the "leave alone" state.
func (n *NullableString) UnmarshalJSON(data []byte) error {
	n.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		n.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = &s

	return nil
}

// MarshalJSON renders Value (or null) when the field is set; an unset field
// would normally be elided via `omitempty` on the parent struct.
func (n NullableString) MarshalJSON() ([]byte, error) {
	if !n.Set || n.Value == nil {
		return []byte("null"), nil
	}

	return json.Marshal(*n.Value)
}
