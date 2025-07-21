package healthcheck

import "fmt"

// convertAnyToInt utility function to convert an slice of any to an slice of integers.
// Moslty used for supported NIPS from the NIP 11 response.
func convertAnyToInt(input []any) ([]int, error) {
	var result []int

	for i, value := range input {
		switch v := value.(type) {
		case int:
			result = append(result, v)
		case float64:
			result = append(result, int(v))
		default:
			return nil, fmt.Errorf("element at index %d not supported - real type is %T", i, value)
		}
	}

	return result, nil
}

// Helper functions for nullable database fields.
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullBool(b bool) *bool {
	return &b
}
