package dbutils

// GetLikeVal returns a string with % at the beginning and end
func GetLikeVal(v string) string {
	return "%" + v + "%"
}

// ConvertListToAny converts a list of any type to a list of any type
func ConvertListToAny[T any](list []T) []any {
	result := make([]any, len(list))
	for idx, v := range list {
		result[idx] = v
	}
	return result
}
