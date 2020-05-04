package magica

import (
	"bytes"
	"os"
	"testing"
	"utils"
)

func TestVoxelObject_Save(t *testing.T) {
	handle, err := os.Open("testdata/test_cube.vox")
	if err != nil {
		t.Errorf("Could not open input test data file: %v", err)
	}

	object, err := GetFromReader(handle)
	if err != nil {
		t.Errorf("Could not read object: %v", err)
	}

	if err := handle.Close(); err != nil {
		t.Errorf("Could not close input test data file: %v", err)
	}

	buf := bytes.Buffer{}
	if err := object.Save(&buf); err != nil {
		t.Errorf("Could not save object: %v", err)
	}

	result, err := utils.CompareToFile(buf.Bytes(), "testdata/test_cube_output.vox")
	if err != nil {
		t.Errorf("Could not read expected output file: %v", err)
	}

	if !result {
		t.Errorf("Output did not equal expected")
	}
}
