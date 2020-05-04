package magica

import (
	"encoding/binary"
	"fmt"
	"geometry"
	"io"
	"io/ioutil"
	"os"
	"utils"
)

const magic = "VOX "

func isHeaderValid(handle io.Reader) bool {
	result, err := getChunkHeader(handle)
	return err == nil && result == magic
}

func getChunkHeader(handle io.Reader) (string, error) {
	limitedReader := io.LimitReader(handle, 4)
	result, err := ioutil.ReadAll(limitedReader)
	return string(result), err
}

func getSizeFromChunk(handle io.Reader) (geometry.Point, error) {
	data, err := getChunkData(handle, 12)

	if err != nil {
		return geometry.Point{}, err
	}

	return geometry.Point{
		X: int(binary.LittleEndian.Uint32(data[0:4])),
		Y: int(binary.LittleEndian.Uint32(data[4:8])),
		Z: int(binary.LittleEndian.Uint32(data[8:12])),
	}, nil
}

func getPointDataFromChunk(handle io.Reader) ([]geometry.PointWithColour, error) {
	data, err := getChunkData(handle, 4)

	if err != nil {
		return getNilValueForPointDataFromChunk(), err
	}

	result := make([]geometry.PointWithColour, len(data)/4)

	for i := 0; i < len(data); i += 4 {
		point := geometry.PointWithColour{
			Point: geometry.Point{X: int(data[i]), Y: int(data[i+1]), Z: int(data[i+2])}, Colour: data[i+3],
		}

		result[i/4] = point
	}

	return result, nil
}

func getPaletteDataFromChunk(handle io.Reader) (data []byte, err error) {
	data, err = getChunkData(handle, 0)
	return
}

func getVoxelObjectFromPointData(size geometry.Point, data []geometry.PointWithColour) VoxelData {
	result := utils.Make3DByteSlice(size)

	for _, p := range data {
		if p.Point.X < size.X && p.Point.Y < size.Y && p.Point.Z < size.Z && p.Colour != 0 {
			result[p.Point.X][p.Point.Y][p.Point.Z] = p.Colour
		}
	}

	return result
}

func skipUnhandledChunk(handle io.Reader) {
	_, _ = getChunkData(handle, 0)
}

func getChunkData(handle io.Reader, minSize int64) ([]byte, error) {
	parsedSize := getSize(handle)

	// Still need to read to the end even if the size
	// is invalid
	limitedReader := io.LimitReader(handle, parsedSize)
	data, err := ioutil.ReadAll(limitedReader)

	if parsedSize < minSize || parsedSize%4 != 0 {
		return nil, fmt.Errorf("invalid chunk size for xyzi")
	}

	if int64(len(data)) < parsedSize {
		return nil, fmt.Errorf("chunk size declared %d but was %d", parsedSize, len(data))
	}

	return data, err
}

func getSize(handle io.Reader) int64 {
	limitedReader := io.LimitReader(handle, 8)
	size, err := ioutil.ReadAll(limitedReader)

	if err != nil {
		return 0
	}

	parsedSize := int64(binary.LittleEndian.Uint32(size[0:4]))
	return parsedSize
}

func getNilValueForPointDataFromChunk() []geometry.PointWithColour {
	return []geometry.PointWithColour{}
}

func GetMagicaVoxelObject(handle io.Reader) (VoxelObject, error) {
	if !isHeaderValid(handle) {
		return VoxelObject{}, fmt.Errorf("header not valid")
	}
	getChunkHeader(handle)

	size := geometry.Point{}
	pointData := make([]geometry.PointWithColour, 0)
	var paletteData []byte

	for {
		chunkType, err := getChunkHeader(handle)

		if err != nil {
			return VoxelObject{}, fmt.Errorf("error reading chunk header: %v", err)
		}

		if chunkType == "" {
			break
		}

		switch chunkType {
		case "SIZE":
			// We only expect one SIZE chunk, but use the last value
			size, err = getSizeFromChunk(handle)
			if err != nil {
				return VoxelObject{}, fmt.Errorf("error reading size chunk: %v", err)
			}
		case "XYZI":
			data, err := getPointDataFromChunk(handle)
			if err != nil {
				return VoxelObject{}, fmt.Errorf("error reading size chunk: %v", err)
			}

			pointData = append(pointData, data...)
		case "RGBA":
			paletteData, err = getPaletteDataFromChunk(handle)
			if err != nil {
				return VoxelObject{}, fmt.Errorf("Error reading palette chunk: %v", err)
			}
		default:
			skipUnhandledChunk(handle)
		}
	}

	if size.X == 0 || size.Y == 0 || size.Z == 0 {
		return VoxelObject{}, fmt.Errorf("invalid size %v", size)
	}

	object := VoxelObject{}
	object.Voxels = getVoxelObjectFromPointData(size, pointData)
	object.PaletteData = paletteData
	object.Size = size
	return object, nil
}

func GetFromReader(handle io.Reader) (v VoxelObject, err error) {
	v, err = GetMagicaVoxelObject(handle)
	return
}

func FromFile(filename string) (v VoxelObject, err error) {
	handle, err := os.Open(filename)
	if err != nil {
		return VoxelObject{}, err
	}

	v, err = GetFromReader(handle)
	if err != nil {
		return v, err
	}

	if err := handle.Close(); err != nil {
		return v, err
	}

	return v, nil
}
