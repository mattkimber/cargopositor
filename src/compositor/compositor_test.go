package compositor

import (
	"bytes"
	"geometry"
	"magica"
	"testing"
	"utils"
)

func Test_getBounds(t *testing.T) {
	testCases := []struct {
		filename string
		expected geometry.Bounds
	}{
		{
			filename: "testdata/example_input.vox",
			expected: geometry.Bounds{
				Min: geometry.Point{X: 2, Y: 2, Z: 5},
				Max: geometry.Point{X: 17, Y: 7, Z: 9},
			},
		},
	}

	for _, tc := range testCases {
		object, err := magica.FromFile(tc.filename)
		if err != nil {
			t.Errorf("Could not read object: %v", err)
		}

		bounds := getBounds(&object)
		if bounds != tc.expected {
			t.Errorf("Object %s expected bounds %v, got %v", tc.filename, tc.expected, bounds)
		}
	}
}

func TestProduceEmpty(t *testing.T) {
	input, err := magica.FromFile("testdata/example_input.vox")
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	output := ProduceEmpty(input)

	buf := bytes.Buffer{}
	if err := output.Save(&buf); err != nil {
		t.Errorf("Could not save object: %v", err)
	}

	result, err := utils.CompareToFile(buf.Bytes(), "testdata/produce_empty.vox")
	if err != nil {
		t.Errorf("Could not read expected output file: %v", err)
	}

	if !result {
		t.Errorf("Output did not equal expected")
	}
}
