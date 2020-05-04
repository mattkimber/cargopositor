package geometry

import (
	"reflect"
	"testing"
)

func TestBounds_GetSize(t *testing.T) {
	testCases := []struct {
		name     string
		bounds   Bounds
		expected Point
	}{
		{"simple input", Bounds{Point{X: 1, Y: 2, Z: 3}, Point{X: 3, Y: 3, Z: 10}}, Point{X: 2, Y: 1, Z: 7}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := &tc.bounds
			if result := b.GetSize(); !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("GetSize() = %v, want %v", result, tc.expected)
			}
		})
	}
}
