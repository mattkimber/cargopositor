package utils

import (
	"bytes"
	"github.com/mattkimber/gandalf/geometry"
	"io/ioutil"
	"os"
)

func Make3DByteSlice(size geometry.Point) [][][]byte {
	result := make([][][]byte, size.X)

	for x := range result {
		result[x] = make([][]byte, size.Y)
		for y := range result[x] {
			result[x][y] = make([]byte, size.Z)
		}
	}

	return result
}

func CompareToFile(data []byte, filename string) (bool, error) {
	handle, err := os.Open(filename)
	if err != nil {
		return false, err
	}

	expected, err := ioutil.ReadAll(handle)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(data, expected) {
		return false, nil
	}

	return true, nil
}
