package compositor

import (
	"bytes"
	"github.com/mattkimber/cargopositor/internal/utils"
	"github.com/mattkimber/gandalf/geometry"
	"github.com/mattkimber/gandalf/magica"
	"os"
	"testing"
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
		return AddScaled(v, src, []string{"2,16"}, []string{"72,79"}, geometry.PointF{X: 1.0, Z: 1.0}, false, false, false, false)
	}
	testOperation(t, fn, "testdata/not_scaled.vox")
}

func TestAddScaled(t *testing.T) {
	src, err := magica.FromFile("testdata/example_cargo.vox")
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	fn := func(v magica.VoxelObject) magica.VoxelObject {
		return AddScaled(v, src, []string{"2-16"}, []string{"72,79"}, geometry.PointF{}, false, false, false, false)
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
		return AddRepeated(v, src, n, []string{"2-16", "254-255"}, []string{"72-79", "1-7"}, false, ignoreMask, ignoreTruncate, false, false, false)
	}
	testOperationWithInputFilename(t, fn, expected, inputFilename)
}

func TestRecolour(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject { return Recolour(v, "2,16", "72,79") }
	testOperation(t, fn, "testdata/recolour.vox")
}

func TestProduceEmpty(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject { return ProduceEmpty(v, nil, nil) }
	testOperation(t, fn, "testdata/produce_empty.vox")
}

func TestIdentity(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject { return Identity(v) }
	testOperation(t, fn, "testdata/identity.vox")
}

func TestLayers(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject { return Identity(v) }
	testOperationWithInputFilenameAndLayers(t, fn, "testdata/layer_0_output.vox", "testdata/example_input_layers.vox", []int{0})
	testOperationWithInputFilenameAndLayers(t, fn, "testdata/layer_1_output.vox", "testdata/example_input_layers.vox", []int{1})
	testOperationWithInputFilenameAndLayers(t, fn, "testdata/layer_2_output.vox", "testdata/example_input_layers.vox", []int{2})
	testOperationWithInputFilenameAndLayers(t, fn, "testdata/layer_3_output.vox", "testdata/example_input_layers.vox", []int{3})
	testOperationWithInputFilenameAndLayers(t, fn, "testdata/layer_4_output.vox", "testdata/example_input_layers.vox", []int{4})

	testOperationWithInputFilenameAndLayers(t, fn, "testdata/layer_012_output.vox", "testdata/example_input_layers.vox", []int{0, 1, 2})
	testOperationWithInputFilenameAndLayers(t, fn, "testdata/layer_023_output.vox", "testdata/example_input_layers.vox", []int{0, 2, 3})
}

func TestStairstep(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject { return Stairstep(v, 4, 1) }
	testOperationWithInputFilename(t, fn, "testdata/stairstep_output.vox", "testdata/stairstep.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject { return Stairstep(v, 2, 1) }
	testOperationWithInputFilename(t, fn, "testdata/stairstep_output_2.vox", "testdata/stairstep.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject { return Stairstep(v, 1, 3) }
	testOperationWithInputFilename(t, fn, "testdata/stairstep_output_3.vox", "testdata/stairstep.vox")
}

func TestRotateAndTile(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject {
		return RotateAndTile(v, 45, -10, 0, geometry.PointF{X: 1.0, Y: 1.0}, BoundingVolume{})
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_45.vox", "testdata/rotate_input.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return RotateAndTile(v, -30, 5, 0, geometry.PointF{X: 1.0, Y: 1.0}, BoundingVolume{})
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_30.vox", "testdata/rotate_input.vox")

}

func TestRotateY(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject {
		return RotateY(v, 0)
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_y_output_0.vox", "testdata/rotate_y_input.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return RotateY(v, 5)
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_y_output_5.vox", "testdata/rotate_y_input.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return RotateY(v, 45)
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_y_output_45.vox", "testdata/rotate_y_input.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return RotateY(v, -30)
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_y_output_30.vox", "testdata/rotate_y_input.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return RotateY(v, 90)
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_y_output_90.vox", "testdata/rotate_y_input.vox")

}

func TestRotateZ(t *testing.T) {
	fn := func(v magica.VoxelObject) magica.VoxelObject {
		return RotateZ(v, 0)
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_z_output_0.vox", "testdata/rotate_z_input.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return RotateZ(v, 5)
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_z_output_5.vox", "testdata/rotate_z_input.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return RotateZ(v, 45)
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_z_output_45.vox", "testdata/rotate_z_input.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return RotateZ(v, -30)
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_z_output_30.vox", "testdata/rotate_z_input.vox")

	fn = func(v magica.VoxelObject) magica.VoxelObject {
		return RotateZ(v, 90)
	}
	testOperationWithInputFilename(t, fn, "testdata/rotate_z_output_90.vox", "testdata/rotate_z_input.vox")

}

func testOperation(t *testing.T, op func(v magica.VoxelObject) magica.VoxelObject, filename string) {
	testOperationWithInputFilename(t, op, filename, "testdata/example_input.vox")
}

func testOperationWithInputFilename(t *testing.T, op func(v magica.VoxelObject) magica.VoxelObject, filename, inputFilename string) {
	testOperationWithInputFilenameAndLayers(t, op, filename, inputFilename, []int{})
}

func testOperationWithInputFilenameAndLayers(t *testing.T, op func(v magica.VoxelObject) magica.VoxelObject, filename, inputFilename string, layers []int) {
	input, err := magica.FromFileWithLayers(inputFilename, layers)
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
		t.Errorf("Output %s did not equal expected", filename)
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
