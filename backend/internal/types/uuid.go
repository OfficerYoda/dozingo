package types

import (
	"encoding/json"
	"reflect"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDParam is a custom param Wrapper for pgtype.UUID. This indriectly allows
// for the direct use of pgtype.UUID and skips the extra parsing from string step
type UUIDParam struct {
	Value pgtype.UUID
	raw   string
}

func (u UUIDParam) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:   "string",
		Format: "uuid",
	}
}

// Receiver / OnParamSet are Huma's hooks for path & query parameter parsing.

func (u *UUIDParam) Receiver() reflect.Value {
	return reflect.ValueOf(&u.raw).Elem()
}

func (u *UUIDParam) OnParamSet(isSet bool, parsed any) {
	if !isSet {
		return
	}
	_ = u.Value.Scan(u.raw)
}

// UnmarshalJSON allows UUIDParam to be used in JSON request bodies.
func (u *UUIDParam) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	u.raw = s
	return u.Value.Scan(s)
}

// MarshalJSON renders the UUID as a JSON string.
func (u UUIDParam) MarshalJSON() ([]byte, error) {
	if !u.Value.Valid {
		return []byte(`""`), nil
	}
	return json.Marshal(u.Value.String())
}
