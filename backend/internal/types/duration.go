package types

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// DurationParam wraps time.Duration for use as an API parameter, allowing direct
// duration values without string parsing overhead.
// Note: due to huma reasons the json "default" tag does not work with this type
type DurationParam struct {
	Value time.Duration
	raw   string
}

func (d DurationParam) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:        "time.Duration",
		Pattern:     `^-?(\d+(?:\.\d+)?(?:ns|us|ms|s|m|h))+$`,
		Description: "Duration e.g. 24h, 1h30m, 1.5h. No days!",
	}
}

// Receiver / OnParamSet are Huma's hooks for path & query parameter parsing.

func (d *DurationParam) Receiver() reflect.Value {
	return reflect.ValueOf(&d.raw).Elem()
}

func (d *DurationParam) OnParamSet(isSet bool, _ any) {
	if !isSet {
		return
	}
	d.Value, _ = time.ParseDuration(d.raw)
}

// UnmarshalJSON allows DurationParam to be used in JSON request bodies.
func (d *DurationParam) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	d.raw = s
	d.Value, err = time.ParseDuration(d.raw)

	return err
}

// MarshalJSON renders the UUID as a JSON string.
func (d DurationParam) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Value.String())
}
