package compositor

import (
	"geometry"
	"magica"
	"math"
	"utils"
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
func ProduceEmpty(v magica.VoxelObject) (r magica.VoxelObject) {
	r = v.Copy()

	iterator := func(x, y, z int) {
		if r.Voxels[x][y][z] == 255 {
			r.Voxels[x][y][z] = 0
		}
	}

	r.Iterate(iterator)

	return r
}

// Scale a cargo object to the cargo area
func AddScaled(dst magica.VoxelObject, src magica.VoxelObject, inputRamp, outputRamp string) (r magica.VoxelObject) {
	r = dst.Copy()

	src = Recolour(src, inputRamp, outputRamp)
	dstBounds := getBounds(&r)
	srcBounds := geometry.Bounds{Min: geometry.Point{}, Max: geometry.Point{X: src.Size.X, Y: src.Size.Y, Z: src.Size.Z}}
	srcSize, dstSize := srcBounds.GetSize(), dstBounds.GetSize()

	iterator := func(x, y, z int) {
		if r.Voxels[x][y][z] == 255 {
			minX := byte(math.Floor(float64(x-dstBounds.Min.X) * (float64(srcSize.X) / float64(dstSize.X+1))))
			minY := byte(math.Floor(float64(y-dstBounds.Min.Y) * (float64(srcSize.Y) / float64(dstSize.Y+1))))
			minZ := byte(math.Floor(float64(z-dstBounds.Min.Z) * (float64(srcSize.Z) / float64(dstSize.Z+1))))

			maxX := byte(math.Ceil(float64((x+1)-dstBounds.Min.X) * (float64(srcSize.X) / float64(dstSize.X+1))))
			maxY := byte(math.Ceil(float64((y+1)-dstBounds.Min.Y) * (float64(srcSize.Y) / float64(dstSize.Y+1))))
			maxZ := byte(math.Ceil(float64((z+1)-dstBounds.Min.Z) * (float64(srcSize.Z) / float64(dstSize.Z+1))))

			values := map[byte]int{}
			max, modalIndex := 0, byte(0)

			for i := minX; i < maxX; i++ {
				for j := minY; j < maxY; j++ {
					for k := minZ; k < maxZ; k++ {

						c := src.Voxels[i][j][k]
						if c != 0 {
							values[c]++
						}
					}
				}
			}

			for k, v := range values {
				if v > max {
					max = v
					modalIndex = k
				}
			}

			r.Voxels[x][y][z] = modalIndex
		}
	}

	r.Iterate(iterator)

	return r
}

// Repeat a cargo object across the cargo area up to n times
func AddRepeated(v magica.VoxelObject, src magica.VoxelObject, n int, inputRamp, outputRamp string) (r magica.VoxelObject) {
	r = v.Copy()

	src = Recolour(src, inputRamp, outputRamp)
	dstBounds := getBounds(&r)
	srcBounds := geometry.Bounds{Min: geometry.Point{}, Max: geometry.Point{X: src.Size.X, Y: src.Size.Y, Z: src.Size.Z}}
	srcSize, dstSize := srcBounds.GetSize(), dstBounds.GetSize()

	items := (dstSize.Y + 1) / srcSize.Y
	cols := (dstSize.X + 1) / srcSize.X
	rows := (dstSize.Z + 1) / srcSize.Z

	iterator := func(x, y, z int) {
		if r.Voxels[x][y][z] == 255 {
			item := (y - dstBounds.Min.Y) / srcSize.Y
			col := (dstBounds.Max.X - x) / srcSize.X
			row := (z - dstBounds.Min.Z) / srcSize.Z

			sx := (dstBounds.Max.X - x) % srcSize.X
			sy := (y - dstBounds.Min.Y) % srcSize.Y
			sz := (z - dstBounds.Min.Z) % srcSize.Z

			if (n == 0 || item+(col*items)+(row*cols*rows) < n) && item < items && col < cols && row < rows {
				r.Voxels[x][y][z] = src.Voxels[sx][sy][sz]
			} else {
				r.Voxels[x][y][z] = 0
			}
		}
	}

	r.Iterate(iterator)
	return r
}

// Recolour according to input/output ramps
func Recolour(v magica.VoxelObject, inputRamp, outputRamp string) (r magica.VoxelObject) {
	r = v.Copy()

	if inputRamp == "" || outputRamp == "" {
		return r
	}

	inputs, outputs := utils.SplitAndParseToInt(inputRamp), utils.SplitAndParseToInt(outputRamp)

	if len(inputs) < 2 || len(outputs) < 2 {
		return r
	}

	inputRampLen, outputRampLen := float64(inputs[1]-inputs[0]), float64(outputs[1]-outputs[0])

	iterator := func(x, y, z int) {
		c := r.Voxels[x][y][z]
		if c >= byte(inputs[0]) && c <= byte(inputs[1]) {
			output := outputs[0] + int(math.Round((float64(int(c)-inputs[0])/inputRampLen)*outputRampLen))
			r.Voxels[x][y][z] = byte(output)
		}
	}

	r.Iterate(iterator)

	return r
}
