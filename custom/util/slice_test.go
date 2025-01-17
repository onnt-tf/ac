package util

import (
	"testing"
)

func TestDeduplicate(t *testing.T) {
	tests := []struct {
		name   string
		input  []int
		output []int
	}{
		{
			name:   "no duplicates",
			input:  []int{1, 2, 3, 4},
			output: []int{1, 2, 3, 4},
		},
		{
			name:   "with duplicates",
			input:  []int{1, 2, 2, 3, 3, 4},
			output: []int{1, 2, 3, 4},
		},
		{
			name:   "empty slice",
			input:  []int{},
			output: []int{},
		},
		{
			name:   "single element",
			input:  []int{1},
			output: []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Deduplicate(tt.input)
			for i, v := range result {
				if v != tt.output[i] {
					t.Errorf("expected %v at index %d, got %v", tt.output[i], i, v)
				}
			}
		})
	}
}

func TestToMap(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		keySelector func(string) string
		output      map[string]string
	}{
		{
			name:        "string to map",
			input:       []string{"apple", "banana", "cherry"},
			keySelector: func(s string) string { return string(s[0]) },
			output:      map[string]string{"a": "apple", "b": "banana", "c": "cherry"},
		},
		{
			name:        "empty slice",
			input:       []string{},
			keySelector: func(s string) string { return string(s[0]) },
			output:      map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToMap(tt.input, tt.keySelector)
			for k, v := range result {
				if v != tt.output[k] {
					t.Errorf("expected key %v to map to %v, got %v", k, tt.output[k], v)
				}
			}
		})
	}
}
