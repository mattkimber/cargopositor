package utils

import (
	"reflect"
	"testing"
)

func TestSplitAndParseToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{"valid data", "1-2-3", []int{1, 2, 3}},
		{"invalid data", "1-no-this-", nil},
		{"single item", "1", []int{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := SplitAndParseToInt(tt.input); !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SplitAndParseToInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}
