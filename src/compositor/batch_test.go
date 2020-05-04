package compositor

import (
	"reflect"
	"testing"
)

func TestFromFile(t *testing.T) {
	batch, err := FromFile("testdata/batch_example.json")
	expected := Batch{
		Files:      []string{"example_input.vox"},
		Operations: []Operation{{Name: "empty", File: "", Type: "produce_empty", InputColourRamp: "20,30"}},
	}

	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("Expected %v, got %v", expected, batch)
	}
}
