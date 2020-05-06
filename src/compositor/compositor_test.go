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

func TestAddScaled(t *testing.T) {
	src, err := magica.FromFile("testdata/example_cargo.vox")
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	fn := func(v magica.VoxelObject) magica.VoxelObject { return AddScaled(v, src, "2,16", "72,79") }
	testOperation(t, fn, "testdata/scaled.vox")
}

func TestAddRepeated(t *testing.T) {
	testAddRepeatedInner(t, 2, "testdata/example_small.vox", "testdata/repeated_small.vox")
	testAddRepeatedInner(t, 6, "testdata/example_tiny.vox", "testdata/repeated_tiny.vox")
	testAddRepeatedInner(t, 0, "testdata/example_tiny.vox", "testdata/repeated_tiny_no_limit.vox")
	testAddRepeatedInner(t, 1, "testdata/example_centred.vox", "testdata/repeated_tiny_centred.vox")
}

func testAddRepeatedInner(t *testing.T, n int, input, expected string) {
	src, err := magica.FromFile(input)
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	fn := func(v magica.VoxelObject) magica.VoxelObject { return AddRepeated(v, src, n, "2,16", "72,79") }
	testOperation(t, fn, expected)
}

func TestRecolour(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject { return Recolour(v, "2,16", "72,79") }
	testOperation(t, fn, "testdata/recolour.vox")
}

func TestProduceEmpty(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject { return ProduceEmpty(v) }
	testOperation(t, fn, "testdata/produce_empty.vox")
}

func testOperation(t *testing.T, op func(v magica.VoxelObject) magica.VoxelObject, filename string) {
	input, err := magica.FromFile("testdata/example_input.vox")
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	output := op(input)

	buf := bytes.Buffer{}
	if err := output.Save(&buf); err != nil {
		t.Errorf("Could not save object: %v", err)
	}

	result, err := utils.CompareToFile(buf.Bytes(), filename)
	if err != nil {
		t.Errorf("Could not read expected output file: %v", err)
	}

	if !result {
		t.Errorf("Output did not equal expected")
	}
}
