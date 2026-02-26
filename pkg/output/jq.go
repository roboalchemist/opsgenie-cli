package output

import (
	"fmt"

	"github.com/itchyny/gojq"
)

// applyJQ evaluates a jq expression against data and returns the result.
// The result is:
//   - A single value if the expression produces one output
//   - A []interface{} if the expression produces multiple outputs
//   - nil if the expression produces no output
func applyJQ(data interface{}, expr string) (interface{}, error) {
	// Normalize data to plain JSON types
	normalized, err := toJSONValue(data)
	if err != nil {
		return nil, fmt.Errorf("normalize for jq: %w", err)
	}

	// Parse the jq expression
	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("parse jq expression %q: %w", expr, err)
	}

	// Compile for better error messages
	code, err := gojq.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("compile jq expression %q: %w", expr, err)
	}

	// Collect all results
	iter := code.Run(normalized)
	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return nil, fmt.Errorf("jq evaluation: %w", err)
		}
		results = append(results, v)
	}

	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		return results[0], nil
	default:
		return results, nil
	}
}
