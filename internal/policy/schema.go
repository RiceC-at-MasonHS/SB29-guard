// Package policy defines the policy data model, loading, validation, and schema helpers.
package policy

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

var (
	compiledOnce   sync.Once
	compiledSchema *jsonschema.Schema
	compileErr     error
)

//go:embed policy.schema.json
var embeddedPolicySchema []byte

func compileSchema() {
	compiledOnce.Do(func() {
		c := jsonschema.NewCompiler()
		// Provide a fake location so $id or relative refs work if added later
		if err := c.AddResource("file://policy.schema.json", bytes.NewReader(embeddedPolicySchema)); err != nil {
			compileErr = err
			return
		}
		sch, err := c.Compile("file://policy.schema.json")
		if err != nil {
			compileErr = err
			return
		}
		compiledSchema = sch
	})
}

// validateAgainstSchema validates the generic map representation of the policy
func validateAgainstSchema(doc map[string]interface{}) error {
	compileSchema()
	if compileErr != nil {
		return fmt.Errorf("schema compile error: %v", compileErr)
	}
	if compiledSchema == nil {
		return errors.New("schema not available")
	}
	if err := compiledSchema.Validate(doc); err != nil {
		return fmt.Errorf("schema validation failed: %s", flattenSchemaError(err))
	}
	return nil
}

// flattenSchemaError produces a concise single-line summary of nested schema errors
func flattenSchemaError(err error) string {
	if err == nil {
		return ""
	}
	// jsonschema/v5 returns *jsonschema.ValidationError with hierarchical context
	var parts []string
	queue := []error{err}
	for len(queue) > 0 {
		e := queue[0]
		queue = queue[1:]
		if ve, ok := e.(*jsonschema.ValidationError); ok {
			loc := strings.TrimPrefix(ve.InstanceLocation, "#/")
			if loc == "" {
				loc = "<root>"
			}
			parts = append(parts, fmt.Sprintf("%s: %s", path.Clean(loc), ve.Message))
			for _, c := range ve.Causes {
				queue = append(queue, c)
			}
		} else {
			parts = append(parts, e.Error())
		}
	}
	if len(parts) == 0 {
		return err.Error()
	}
	return strings.Join(parts, "; ")
}
