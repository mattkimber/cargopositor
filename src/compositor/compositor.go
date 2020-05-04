package compositor

import (
	"geometry"
	"magica"
)

func getBounds(v *magica.VoxelObject) geometry.Bounds {
	min := geometry.Point{X: 255, Y: 255, Z: 255}
	max := geometry.Point{}

	iterator := func(x, y, z int) {
		if v.Voxels[x][y][z] == 255 {
			if x < min.X {
				min.X = x
			}
			if y < min.Y {
				min.Y = y
			}
			if z < min.Z {
				min.Z = z
			}
			if x > max.X {
				max.X = x
			}
			if y > max.Y {
				max.Y = y
			}
			if z > max.Z {
				max.Z = z
			}
		}
	}

	v.Iterate(iterator)

	return geometry.Bounds{Min: min, Max: max}
}

// Return the base object without any cargo
// (remove special voxels)
func ProduceEmpty(v magica.VoxelObject) magica.VoxelObject {
	iterator := func(x, y, z int) {
		if v.Voxels[x][y][z] == 255 {
			v.Voxels[x][y][z] = 0
		}
	}

	v.Iterate(iterator)

	return v
}