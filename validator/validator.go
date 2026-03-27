package validator

import (
	"encoding/json"
	"fmt"
	"go_project/config"
	"go_project/logger"
)

type Violation struct {
	Field    string
	Issue    string
	Expected string
	Got      string
}

func ValidateJSON(body []byte, schema map[string]interface{}, direction string, contract *config.Contract) []Violation {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		v := Violation{
			Field:    "body",
			Issue:    "invalid JSON",
			Expected: "valid JSON",
			Got:      string(body),
		}
		logger.LogViolation(contract.Endpoint, contract.Method, direction, "body", "invalid JSON", "valid JSON", string(body))
		return []Violation{v}
	}

	violations := validateObject(data, schema, "")

	for _, v := range violations {
		logger.LogViolation(contract.Endpoint, contract.Method, direction, v.Field, v.Issue, v.Expected, v.Got)
	}

	if len(violations) == 0 {
		logger.LogOK(contract.Endpoint, contract.Method, direction)
	}

	return violations
}

// validateObject checks a data map against a schema map recursively.
// prefix is used to build dotted field paths for nested violations (e.g. "address.street").
func validateObject(data map[string]interface{}, schema map[string]interface{}, prefix string) []Violation {
	var violations []Violation

	// Check every field declared in the schema
	for field, expectedSchema := range schema {
		fullField := field
		if prefix != "" {
			fullField = prefix + "." + field
		}

		val, exists := data[field]
		if !exists {
			violations = append(violations, Violation{
				Field:    fullField,
				Issue:    "missing field",
				Expected: describeSchema(expectedSchema),
				Got:      "null",
			})
			continue
		}

		violations = append(violations, validateValue(val, expectedSchema, fullField)...)
	}

	// Check for extra fields not declared in the schema
	for field := range data {
		fullField := field
		if prefix != "" {
			fullField = prefix + "." + field
		}
		if _, exists := schema[field]; !exists {
			violations = append(violations, Violation{
				Field:    fullField,
				Issue:    "unexpected field",
				Expected: "not present",
				Got:      getType(data[field]),
			})
		}
	}

	return violations
}

// validateValue checks a single value against its schema definition.
// The schema can be a string (primitive type), a map (nested object), or a slice (array with item type).
func validateValue(val interface{}, schema interface{}, fullField string) []Violation {
	var violations []Violation

	switch s := schema.(type) {

	case string:
		// Primitive type check: schema is just "string", "number", "boolean", etc.
		actualType := getType(val)
		if actualType != s {
			violations = append(violations, Violation{
				Field:    fullField,
				Issue:    "wrong type",
				Expected: s,
				Got:      actualType,
			})
		}

	case map[string]interface{}:
		// Nested object: schema is itself a schema map
		nested, ok := val.(map[string]interface{})
		if !ok {
			violations = append(violations, Violation{
				Field:    fullField,
				Issue:    "wrong type",
				Expected: "object",
				Got:      getType(val),
			})
		} else {
			violations = append(violations, validateObject(nested, s, fullField)...)
		}

	case []interface{}:
		// Array: schema is a slice whose first element is the expected item type
		arr, ok := val.([]interface{})
		if !ok {
			violations = append(violations, Violation{
				Field:    fullField,
				Issue:    "wrong type",
				Expected: "array",
				Got:      getType(val),
			})
		} else if len(s) > 0 {
			for i, item := range arr {
				itemField := fmt.Sprintf("%s[%d]", fullField, i)
				violations = append(violations, validateValue(item, s[0], itemField)...)
			}
		}
	}

	return violations
}

// describeSchema returns a human-readable description of a schema for use in violation messages.
func describeSchema(schema interface{}) string {
	switch schema.(type) {
	case string:
		return schema.(string)
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	default:
		return "unknown"
	}
}

func getType(val interface{}) string {
	switch val.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}
