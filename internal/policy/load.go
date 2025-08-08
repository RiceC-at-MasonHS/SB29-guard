package policy // import "internal/policy" implements loading & schema validation utilities.

import (
	"errors"
	"reflect"
	"time"

	"gopkg.in/yaml.v3"
)

// StrictValidation controls whether JSON Schema validation is enforced during Load.
var StrictValidation = true

// Load parses YAML policy bytes into Policy struct (no schema validation yet)
func Load(b []byte) (*Policy, error) {
	// First unmarshal into generic map for schema validation
	var generic map[string]interface{}
	if err := yaml.Unmarshal(b, &generic); err != nil {
		return nil, err
	}
	if StrictValidation {
		normalizeDates(generic)
		if err := validateAgainstSchema(generic); err != nil {
			return nil, err
		}
	}
	var p Policy
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	if len(p.Records) == 0 {
		return nil, errors.New("no records")
	}
	return &p, nil
}

// normalizeDates walks the generic map and converts any time.Time values to YYYY-MM-DD strings
func normalizeDates(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, subv := range val {
			// recurse first
			normalizeDates(subv)
			// convert after recursion if time
			if rt := reflect.TypeOf(subv); rt != nil && rt.String() == "time.Time" {
				if tt, ok := subv.(time.Time); ok {
					val[k] = tt.Format("2006-01-02")
				}
			}
		}
	case []interface{}:
		for i, subv := range val {
			normalizeDates(subv)
			if rt := reflect.TypeOf(subv); rt != nil && rt.String() == "time.Time" {
				if tt, ok := subv.(time.Time); ok {
					val[i] = tt.Format("2006-01-02")
				}
			}
		}
	}
}
