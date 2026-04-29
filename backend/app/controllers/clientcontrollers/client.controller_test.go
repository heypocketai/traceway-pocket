package clientcontrollers

import (
	"encoding/json"
	"testing"
)

func TestIsEmptyRaw(t *testing.T) {
	cases := []struct {
		name string
		in   json.RawMessage
		want bool
	}{
		{"nil", nil, true},
		{"empty bytes", json.RawMessage(""), true},
		{"whitespace null", json.RawMessage(" null "), true},
		{"plain null", json.RawMessage("null"), true},
		{"empty array", json.RawMessage("[]"), true},
		{"empty object", json.RawMessage("{}"), true},
		{"non-empty array", json.RawMessage("[1]"), false},
		{"non-empty object", json.RawMessage(`{"a":1}`), false},
		{"string value", json.RawMessage(`"x"`), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isEmptyRaw(c.in); got != c.want {
				t.Fatalf("isEmptyRaw(%q) = %v, want %v", string(c.in), got, c.want)
			}
		})
	}
}
