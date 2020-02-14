package interop

import (
	"reflect"
	"strings"
	"testing"
)

var headerDecodeCases = []struct {
	name string
	json string
	want Header
}{
	// TODO(@MattWindsor91): add more test cases
	{"sbrlx",
		`{
		"name": "SBRlx",
		"locations": null,
		"init": { "x": 0, "y": 0 },
		"postcondition": "exists (0:a == 0 /\\ 1:a == 0)"
	 }`, Header{
			Name:      "SBRlx",
			Locations: nil,
			Init: map[string]int{
				"x": 0,
				"y": 0,
			},
			Postcondition: `exists (0:a == 0 /\ 1:a == 0)`,
		},
	},
}

// TestReadHeader tests that we can read headers properly from JSON.
func TestReadHeader(t *testing.T) {
	for _, c := range headerDecodeCases {
		t.Run(c.name, func(t *testing.T) {
			rd := strings.NewReader(c.json)
			if got, err := ReadHeader(rd); err != nil {
				t.Errorf("decode failed with error (%s): %q", err, c.json)
			} else if !reflect.DeepEqual(*got, c.want) {
				t.Errorf("decode got=%v; want=%v; input: %q", got, c.want, c.json)
			}
		})
	}
}