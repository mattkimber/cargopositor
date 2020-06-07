package compositor

import (
	"bytes"
	"geometry"
	"magica"
	"os"
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

		bounds := getBounds(&object, false)
		if bounds != tc.expected {
			t.Errorf("Object %s expected bounds %v, got %v", tc.filename, tc.expected, bounds)
		}
	}
}

func TestAddScaledWithSize(t *testing.T) {
	src, err := magica.FromFile("testdata/2by2.vox")
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	fn := func(v magica.VoxelObject) magica.VoxelObject {
		return AddScaled(v, src, "2,16", "72,79", geometry.PointF{X: 1.0, Z: 1.0}, false)
	}
	testOperation(t, fn, "testdata/not_scaled.vox")
}

func TestAddScaled(t *testing.T) {
	src, err := magica.FromFile("testdata/example_cargo.vox")
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	fn := func(v magica.VoxelObject) magica.VoxelObject {
		return AddScaled(v, src, "2,16", "72,79", geometry.PointF{}, false)
	}
	testOperation(t, fn, "testdata/scaled.vox")
}

func TestAddRepeated(t *testing.T) {
	testAddRepeatedInner(t, 2, "testdata/example_small.vox", "testdata/repeated_small.vox", "testdata/example_input.vox", false, false)
	testAddRepeatedInner(t, 6, "testdata/example_tiny.vox", "testdata/repeated_tiny.vox", "testdata/example_input.vox", false, false)
	testAddRepeatedInner(t, 0, "testdata/example_tiny.vox", "testdata/repeated_tiny_no_limit.vox", "testdata/example_input.vox", false, false)
	testAddRepeatedInner(t, 1, "testdata/example_centred.vox", "testdata/repeated_tiny_centred.vox", "testdata/example_input.vox", false, false)
}

func TestIgnoreMask(t *testing.T) {
	testAddRepeatedInner(t, 0, "testdata/no_mask_b.vox", "testdata/no_mask.vox", "testdata/no_mask_a.vox", true, true)
	testAddRepeatedInner(t, 0, "testdata/no_mask_b.vox", "testdata/no_mask_tiny_1.vox", "testdata/example_tiny.vox", true, true)
	testAddRepeatedInner(t, 0, "testdata/example_tiny.vox", "testdata/no_mask_tiny_2.vox", "testdata/no_mask_a.vox", true, true)
}

func testAddRepeatedInner(t *testing.T, n int, input, expected string, inputFilename string, ignoreMask bool, ignoreTruncate bool) {
	src, err := magica.FromFile(input)
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	fn := func(v magica.VoxelObject) magica.VoxelObject {
		return AddRepeated(v, src, n, "2,16", "72,79", ignoreMask, ignoreTruncate, false)
	}
	testOperationWithInputFilename(t, fn, expected, inputFilename)
}

func TestRecolour(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject { return Recolour(v, "2,16", "72,79") }
	testOperation(t, fn, "testdata/recolour.vox")
}

func TestProduceEmpty(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject { return ProduceEmpty(v) }
	testOperation(t, fn, "testdata/produce_empty.vox")
}

func TestStairstep(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject { return Stairstep(v, 4, 1) }
	testOperationWithInputFilename(t, fn, "testdata/stairstep_output.vox", "testdata/stairstep.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject { return Stairstep(v, 2, 1) }
	testOperationWithInputFilename(t, fn, "testdata/stairstep_output_2.vox", "testdata/stairstep.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject { return Stairstep(v, 1, 3) }
	testOperationWithInputFilename(t, fn, "testdata/stairstep_output_3.vox", "testdata/stairstep.vox")
}

func TestRotate(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject {
		return Rotate(v, 45, -10, 0, geometry.PointF{X: 1.0, Y: 1.0}, BoundingVolume{})
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_45.vox", "testdata/rotate_input.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return Rotate(v, -30, 5, 0, geometry.PointF{X: 1.0, Y: 1.0}, BoundingVolume{})
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_30.vox", "testdata/rotate_input.vox")

}

func testOperation(t *testing.T, op func(v magica.VoxelObject) magica.VoxelObject, filename string) {
	testOperationWithInputFilename(t, op, filename, "testdata/example_input.vox")
}

func testOperationWithInputFilename(t *testing.T, op func(v magica.VoxelObject) magica.VoxelObject, filename, inputFilename string) {
	input, err := magica.FromFile(inputFilename)
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	output := op(input)

	buf := bytes.Buffer{}
	if err := output.Save(&buf); err != nil {
		t.Errorf("Could not save object: %v", err)
	}

	// Create test output if it doesn't exist. Check created output
	// in MagicaVoxel to ensure it is correct.
	if _, err := os.Stat(filename); err != nil {
		of, _ := os.Create(filename)
		output.Save(of)
	}

	result, err := utils.CompareToFile(buf.Bytes(), filename)
	if err != nil {
		t.Errorf("Could not read expected output file: %v", err)
	}

	if !result {
		t.Errorf("Output did not equal expected")
	}
}

func TestRemove(t *testing.T) {
	src, err := magica.FromFile("testdata/example_small.vox")
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	fn := func(v magica.VoxelObject) magica.VoxelObject {
		return Remove(v, src, 0)
	}
	testOperation(t, fn, "testdata/remove.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return Remove(v, src, 255)
	}
	testOperation(t, fn, "testdata/clip.vox")
}
