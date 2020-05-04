package compositor

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

type Batch struct {
	Files      []string    `json:"files"`
	Operations []Operation `json:"operations"`
}

type Operation struct {
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
