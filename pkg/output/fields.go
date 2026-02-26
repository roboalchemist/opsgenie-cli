package output

import (
	"encoding/json"
	"fmt"
)

// filterFields filters the given data to only include the specified fields.
// It handles:
//   - map[string]interface{}: returns a new map with only the requested keys
//   - []interface{}: applies field filtering to each element
//   - Other types: marshals to JSON, then filters
func filterFields(data interface{}, fields []string) (interface{}, error) {
	if len(fields) == 0 {
		return data, nil
	}

	// Normalize to a JSON-decoded structure so we can filter consistently
	normalized, err := toJSONValue(data)
	if err != nil {
		return nil, fmt.Errorf("normalize for field filter: %w", err)
	}

	return applyFieldFilter(normalized, fields), nil
}

// applyFieldFilter recursively applies field filtering to a JSON value.
func applyFieldFilter(data interface{}, fields []string) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(fields))
		for _, f := range fields {
			if val, ok := v[f]; ok {
				result[f] = val
			}
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = applyFieldFilter(item, fields)
		}
		return result

	default:
		return data
	}
}

// toJSONValue round-trips data through JSON encoding to get a
// map[string]interface{} / []interface{} / primitive structure.
func toJSONValue(data interface{}) (interface{}, error) {
	// If it's already a plain JSON-compatible value, return as-is
	switch data.(type) {
	case map[string]interface{}, []interface{}, string, float64, bool, nil:
		return data, nil
	}

	// Marshal then unmarshal to normalize structs/custom types
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	var out interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return out, nil
}
