package util

// Deduplicate removes duplicates from a slice while preserving order
func Deduplicate[T comparable](input []T) []T {
	if len(input) == 0 {
		return input
	}

	seen := make(map[T]struct{}, len(input))
	result := make([]T, 0, len(input))

	for _, v := range input {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}

// ToMap converts a slice of elements to a map, where the keys are determined by a keySelector function.
func ToMap[T any, K comparable](input []T, keySelector func(T) K) map[K]T {
	if len(input) == 0 || keySelector == nil {
		return nil
	}

	result := make(map[K]T, len(input)) // Initialize map with the size of input slice
	for _, v := range input {
		key := keySelector(v) // Generate the key using the keySelector function
		result[key] = v       // Add the element to the map
	}
	return result
}
