package magica

import (
	"geometry"
	"utils"
)

type VoxelData [][][]byte

type VoxelObject struct {
	Voxels      [][][]byte
	PaletteData []byte
	Size        geometry.Point
}

func (v *VoxelObject) GetPoints() (result []geometry.PointWithColour) {
	ct := 0

	v.Iterate(func(x, y, z int) {
		if v.Voxels[x][y][z] != 0 {
			ct++
		}
	})

	result = make([]geometry.PointWithColour, ct)
	ct = 0

	v.Iterate(func(x, y, z int) {
		if v.Voxels[x][y][z] != 0 {
			result[ct] = geometry.PointWithColour{
				Point:  geometry.Point{X: x, Y: y, Z: z},
				Colour: v.Voxels[x][y][z],
			}
			ct++
		}
	})

	return result
}

func (v *VoxelObject) Copy() (result VoxelObject) {
	result = VoxelObject{}

	result.Size = v.Size
	// We don't do anything with the palette data, so a shallow copy is okay
	result.PaletteData = v.PaletteData

	result.Voxels = utils.Make3DByteSlice(v.Size)
	v.Iterate(func(x, y, z int) { result.Voxels[x][y][z] = v.Voxels[x][y][z] })

	return
}

func (v *VoxelObject) Iterate(iterator func(int, int, int)) {
	for x := 0; x < v.Size.X; x++ {
		for y := 0; y < v.Size.Y; y++ {
			for z := 0; z < v.Size.Z; z++ {
				iterator(x, y, z)
			}
		}
	}
}
