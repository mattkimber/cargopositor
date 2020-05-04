package compositor

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"magica"
	"os"
	"strings"
)

type Batch struct {
	Files      []string    `json:"files"`
	Operations []Operation `json:"operations"`
}

type Operation struct {
	Name string `json:"name"`
	Type string `json:"type"`
	File string `json:"file"`
}

func FromJson(handle io.Reader) (b Batch, err error) {
	data, err := ioutil.ReadAll(handle)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &b)
	return
}

func FromFile(filename string) (b Batch, err error) {
	handle, err := os.Open(filename)
	if err != nil {
		return
	}

	b, err = FromJson(handle)
	if err != nil {
		handle.Close()
		return
	}

	err = handle.Close()
	return
}

func getOutputFileName(filename, suffix string) string {
	lastExtension := strings.LastIndex(filename, ".")
	if lastExtension != -1 {
		return filename[:lastExtension] + suffix + ".vox"
	}

	return filename + suffix + ".vox"
}

func saveFile(v *magica.VoxelObject, filename string) error {
	handle, err := os.Create(filename)
	if err != nil {
		return err
	}

	err = v.Save(handle)
	if err != nil {
		handle.Close()
		return err
	}

	err = handle.Close()
	return err
}

func (b *Batch) Run() error {
	for _, f := range b.Files {
		input, err := magica.FromFile(f)
		if err != nil {
			return err
		}

		for _, op := range b.Operations {
			switch op.Type {
			case "produce_empty":
				output := ProduceEmpty(input)
				if err := saveFile(&output, getOutputFileName(f, op.Name)); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unkown operation %s", op.Type)
			}
		}
	}

	return nil
}
