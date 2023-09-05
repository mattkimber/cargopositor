package compositor

import (
	"github.com/mattkimber/cargopositor/internal/utils"
	"github.com/mattkimber/gandalf/geometry"
	"github.com/mattkimber/gandalf/magica"
	"log"
	"math"
	"strings"
)

func getBounds(v *magica.VoxelObject, ignoreMask bool) geometry.Bounds {

	minP := geometry.Point{X: v.Size.X, Y: v.Size.Y, Z: v.Size.Z}
	maxP := geometry.Point{}

	if ignoreMask {
		return geometry.Bounds{Min: maxP, Max: geometry.Point{X: v.Size.X - 1, Y: v.Size.Y - 1, Z: v.Size.Z - 1}}
	}

	iterator := func(x, y, z int) {
		if v.Voxels[x][y][z] == 255 {
			if x < minP.X {
				minP.X = x
			}
			if y < minP.Y {
				minP.Y = y
			}
			if z < minP.Z {
				minP.Z = z
			}
			if x > maxP.X {
				maxP.X = x
			}
			if y > maxP.Y {
				maxP.Y = y
			}
			if z > maxP.Z {
				maxP.Z = z
			}
		}
	}

	v.Iterate(iterator)

	return geometry.Bounds{Min: minP, Max: maxP}
}

// ProduceEmpty returns the base object without any cargo
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

// Identity returns the base object without any changes at all
func Identity(v magica.VoxelObject) (r magica.VoxelObject) {
	r = v.Copy()
	return r
}

// RotateAndTile (and tile) the base object
func RotateAndTile(v magica.VoxelObject, angle float64, xOffset, yOffset int, scale geometry.PointF, boundingVolume BoundingVolume) (r magica.VoxelObject) {
	radians := (angle * math.Pi) / 180

	// If no bounding volume was supplied default to (0,0,0)-(max, max, max)
	if (boundingVolume.Max == geometry.Point{}) {
		boundingVolume.Max = geometry.Point{X: v.Size.X, Y: v.Size.Y, Z: v.Size.Z}
	}

	// If no scale is supplied default to (1,1,1)
	if (scale == geometry.PointF{}) {
		scale = geometry.PointF{X: 1, Y: 1, Z: 1}
	}

	r = v.Copy()

	bvx := boundingVolume.Max.X - boundingVolume.Min.X
	bvy := boundingVolume.Max.Y - boundingVolume.Min.Y
	bvz := boundingVolume.Max.Z - boundingVolume.Min.Z

	// Clear the object
	iterator := func(x, y, z int) {
		r.Voxels[x][y][z] = 0
	}

	r.Iterate(iterator)

	// RotateAndTile the output
	iterator = func(x, y, z int) {
		sx := ((bvx + xOffset + int((float64(x)*math.Cos(radians)-float64(y)*math.Sin(radians))*scale.X)) % bvx) + boundingVolume.Min.X
		sy := ((bvy + yOffset + int((float64(x)*math.Sin(radians)+float64(y)*math.Cos(radians))*scale.Y)) % bvy) + boundingVolume.Min.Y
		sz := ((z + bvz) % bvz) + boundingVolume.Min.Z

		if r.Voxels[x][y][z] == 0 && sx >= 0 && sy >= 0 && sx < v.Size.X && sy < v.Size.Y {
			r.Voxels[x][y][z] = v.Voxels[sx][sy][sz]
		}
	}

	r.Iterate(iterator)

	return r
}

// Stairstep the base object (for every m steps in x, move n steps in z)
func Stairstep(v magica.VoxelObject, m float64, n int) (r magica.VoxelObject) {
	r = v.Copy()

	// Clear the object
	iterator := func(x, y, z int) {
		r.Voxels[x][y][z] = 0
	}

	r.Iterate(iterator)

	// Stairstep the output
	iterator = func(x, y, z int) {
		step := z + int((float64(x)/m)*float64(n))
		begin := step

		if x > 0 {
			prevStep := z + int((float64(x-1)/m)*float64(n))
			if prevStep < step {
				begin -= n
			}
		}

		for s := begin; s < step+n; s++ {
			if s >= 0 && s < v.Size.Z {
				if r.Voxels[x][y][s] == 0 {
					r.Voxels[x][y][s] = v.Voxels[x][y][z]
				}
			}
		}
	}

	v.Iterate(iterator)

	return
}

// AddScaled scales a cargo object to the cargo area
func AddScaled(dst magica.VoxelObject, src magica.VoxelObject, inputRamps, outputRamps []string, scaleLogic geometry.PointF, overwrite bool, ignoreMask bool, maskOriginal bool, maskNew bool) (r magica.VoxelObject) {
	r = dst.Copy()

	// If there is an input/output ramp, we always use the first one when scaling
	if len(inputRamps) > 0 && len(outputRamps) > 0 {
		src = Recolour(src, inputRamps[0], outputRamps[0])
	}

	dstBounds := getBounds(&r, ignoreMask)
	srcBounds := geometry.Bounds{Min: geometry.Point{}, Max: geometry.Point{X: src.Size.X, Y: src.Size.Y, Z: src.Size.Z}}
	srcSize, dstSize := srcBounds.GetSize(), dstBounds.GetSize()

	scale := geometry.PointF{
		X: ((float64(srcSize.X) / float64(dstSize.X+1)) * (1 - scaleLogic.X)) + scaleLogic.X,
		Y: (float64(srcSize.Y)/float64(dstSize.Y+1))*(1-scaleLogic.Y) + scaleLogic.Y,
		Z: (float64(srcSize.Z)/float64(dstSize.Z+1))*(1-scaleLogic.Z) + scaleLogic.Z,
	}

	iterator := func(x, y, z int) {
		if (ignoreMask && r.Voxels[x][y][z] == 0) || r.Voxels[x][y][z] == 255 || overwrite {
			minX := byte(math.Floor(float64(x-dstBounds.Min.X) * scale.X))
			minY := byte(math.Floor(float64(y-dstBounds.Min.Y) * scale.Y))
			minZ := byte(math.Floor(float64(z-dstBounds.Min.Z) * scale.Z))

			maxX := byte(math.Ceil(float64((x+1)-dstBounds.Min.X) * scale.X))
			maxY := byte(math.Ceil(float64((y+1)-dstBounds.Min.Y) * scale.Y))
			maxZ := byte(math.Ceil(float64((z+1)-dstBounds.Min.Z) * scale.Z))

			values := map[byte]int{}
			maxIndexCount, modalIndex := 0, byte(0)

			for i := minX; i < maxX; i++ {
				for j := minY; j < maxY; j++ {
					for k := minZ; k < maxZ; k++ {

						if i < byte(srcBounds.Max.X) && j < byte(srcBounds.Max.Y) && k < byte(srcBounds.Max.Z) {
							c := src.Voxels[i][j][k]
							if c != 0 {
								values[c]++
							}
						}
					}
				}
			}

			for k, v := range values {
				if v > maxIndexCount {
					maxIndexCount = v
					modalIndex = k
				}
			}

			if maskNew && modalIndex != 0 {
				modalIndex = 255
			}

			if !overwrite || modalIndex != 0 {
				r.Voxels[x][y][z] = modalIndex
			}
		} else if maskOriginal && r.Voxels[x][y][z] != 0 {
			r.Voxels[x][y][z] = 255
		}
	}

	r.Iterate(iterator)

	return r
}

// AddRepeated repeats a cargo object across the cargo area up to n times
func AddRepeated(v magica.VoxelObject, originalSrc magica.VoxelObject, n int, inputRamps, outputRamps []string, overwrite bool, ignoreMask bool, ignoreTruncation bool, maskOriginal bool, maskNew bool, flipX bool) (r magica.VoxelObject) {
	r = v.Copy()

	dstBounds := getBounds(&r, ignoreMask)
	srcBounds := geometry.Bounds{Min: geometry.Point{}, Max: geometry.Point{X: originalSrc.Size.X, Y: originalSrc.Size.Y, Z: originalSrc.Size.Z}}
	srcSize, dstSize := srcBounds.GetSize(), dstBounds.GetSize()

	lastItem := -1
	ramps := len(inputRamps)

	// Create all the necessary recolour objects
	srcObjects := make([]magica.VoxelObject, ramps)
	if ramps > 0 && len(inputRamps) == len(outputRamps) {
		for idx := range inputRamps {
			srcObjects[idx] = Recolour(originalSrc, inputRamps[idx], outputRamps[idx])
		}
	} else {
		srcObjects = append(srcObjects, originalSrc)
		ramps = 1
	}

	items := (dstSize.Y + 1) / srcSize.Y
	cols := (dstSize.X + 1) / srcSize.X
	rows := (dstSize.Z + 1) / srcSize.Z

	yOffset := ((dstSize.Y + 1) - (items * srcSize.Y)) / 2
	xOffset := ((dstSize.X) - (cols * srcSize.X)) / 2

	if ignoreTruncation {
		yOffset = 0
		xOffset = 0
	}

	var src magica.VoxelObject

	iterator := func(x, y, z int) {
		if (ignoreMask && r.Voxels[x][y][z] == 0) || r.Voxels[x][y][z] == 255 || overwrite {
			item := ((y - yOffset) - dstBounds.Min.Y) / srcSize.Y
			col := (dstBounds.Max.X - (x + (xOffset / 2))) / (srcSize.X + xOffset)
			row := (z - dstBounds.Min.Z) / srcSize.Z

			if item+(col*items)+(row*cols*rows) != lastItem {
				// Pick the recolour ramp for this item
				src = srcObjects[(item+(col*items)+(row*cols*rows))%ramps]
			}

			lastItem = item + (col * items) + (row * cols * rows)

			sx := srcSize.X - 1 - (((dstBounds.Max.X) - (x + (xOffset / 2))) % (srcSize.X + xOffset))

			if flipX {
				sx = (x - dstBounds.Min.X) % srcSize.X
			}
			sy := (y - (yOffset + dstBounds.Min.Y)) % srcSize.Y
			sz := (z - dstBounds.Min.Z) % srcSize.Z

			if (n == 0 || overwrite || item+(col*items)+(row*cols*rows) < n) && ((n == 0 && ignoreTruncation) || (item < items && col < cols && row < rows)) && (y-dstBounds.Min.Y) >= yOffset {
				if sx < 0 || sx >= srcSize.X {
					r.Voxels[x][y][z] = 0
				} else {
					if !overwrite || src.Voxels[sx][sy][sz] != 0 {
						if maskNew {
							if src.Voxels[sx][sy][sz] == 0 {
								r.Voxels[x][y][z] = 0
							} else {
								r.Voxels[x][y][z] = 255
							}
						} else {
							r.Voxels[x][y][z] = src.Voxels[sx][sy][sz]
						}
					}
				}
			} else if !overwrite {
				r.Voxels[x][y][z] = 0
			}
		} else if r.Voxels[x][y][z] != 0 && maskOriginal {
			r.Voxels[x][y][z] = 255
		}
	}

	r.Iterate(iterator)
	return r
}

// Remove one voxel object from another (or clip against a colour)
func Remove(v magica.VoxelObject, src magica.VoxelObject, index uint8) (r magica.VoxelObject) {
	r = v.Copy()

	iterator := func(x, y, z int) {
		if x < src.Size.X && y < src.Size.Y && z < src.Size.Z && src.Voxels[x][y][z] != index {
			r.Voxels[x][y][z] = 0
		}
	}

	r.Iterate(iterator)

	return r
}

type Ramp struct {
	InputLength      float64
	OutputLength     float64
	StartIndex       int
	EndIndex         int
	OutputStartIndex int
}

// Recolour according to input/output ramps
func Recolour(v magica.VoxelObject, inputRamp, outputRamp string) (r magica.VoxelObject) {
	r = v.Copy()

	if inputRamp == "" || outputRamp == "" {
		return r
	}

	// Deal with the old GoRender format
	if !strings.ContainsRune(inputRamp, '-') && !strings.ContainsRune(outputRamp, '-') {
		inputRamp = strings.Replace(inputRamp, ",", "-", -1)
		outputRamp = strings.Replace(outputRamp, ",", "-", -1)
	}

	inputRamps := strings.Split(inputRamp, ",")
	outputRamps := strings.Split(outputRamp, ",")

	if len(inputRamps) != len(outputRamps) {
		log.Print("WARNING: Invalid colour remap specification (ramp lengths don't match) - object not recoloured")
		return r
	}

	ramps := make([]Ramp, len(inputRamps))

	for idx := range inputRamps {
		inputs, outputs := utils.SplitAndParseToInt(inputRamps[idx]), utils.SplitAndParseToInt(outputRamps[idx])

		if len(inputs) < 2 || len(outputs) < 2 {
			log.Printf("WARNING: Invalid colour remap specification %s/%s (invalid ramp length) - object not recoloured", inputRamps[idx], outputRamps[idx])
			return r
		}

		ramps[idx] = Ramp{
			InputLength:      float64(inputs[1] - inputs[0]),
			OutputLength:     float64(outputs[1] - outputs[0]),
			StartIndex:       inputs[0],
			EndIndex:         inputs[1],
			OutputStartIndex: outputs[0],
		}

	}

	iterator := func(x, y, z int) {
		c := r.Voxels[x][y][z]
		for _, rmp := range ramps {
			if c >= byte(rmp.StartIndex) && c <= byte(rmp.EndIndex) {
				output := rmp.OutputStartIndex + int(math.Round((float64(int(c)-rmp.StartIndex)/rmp.InputLength)*rmp.OutputLength))
				r.Voxels[x][y][z] = byte(output)

				// Only apply the first ramp we find (don't repeatedly map colours)
				break
			}
		}
	}

	r.Iterate(iterator)

	return r
}

// RotateY Rotates an object around its Y axis
func RotateY(v magica.VoxelObject, angle float64) (r magica.VoxelObject) {
	sin, cos := math.Sin(degToRad(angle)), math.Cos(degToRad(angle))

	orgMidpointX := float64(v.Size.X) / 2
	orgMidpointZ := float64(v.Size.Z) / 2

	xVector := (orgMidpointX * math.Abs(cos)) + (orgMidpointZ * math.Abs(sin))
	zVector := (orgMidpointX * math.Abs(sin)) + (orgMidpointZ * math.Abs(cos))

	sizeX, sizeZ := int(math.Ceil(xVector*2)), int(math.Ceil(zVector*2))

	r = magica.VoxelObject{
		Voxels:      nil,
		PaletteData: v.PaletteData,
		Size:        geometry.Point{X: sizeX, Y: v.Size.Y, Z: sizeZ},
	}

	// Create the voxel array
	r.Voxels = make([][][]byte, r.Size.X)
	for x := 0; x < r.Size.X; x++ {
		r.Voxels[x] = make([][]byte, r.Size.Y)
		for y := 0; y < r.Size.Y; y++ {
			r.Voxels[x][y] = make([]byte, r.Size.Z)
		}
	}

	vMidpointX := float64(v.Size.X) / 2
	vMidpointZ := float64(v.Size.Z) / 2

	iterator := func(x, y, z int) {

		fdx := float64(x) - (float64(r.Size.X) / 2)
		fdz := float64(z) - (float64(r.Size.Z) / 2)

		fdx, fdz = (fdx*cos)+(fdz*sin), (fdx*-sin)+(fdz*cos)

		dx := int(math.Ceil(fdx + vMidpointX))
		dz := int(math.Ceil(fdz + vMidpointZ))

		if dx >= 0 && dz >= 0 && dx < v.Size.X && dz < v.Size.Z {
			r.Voxels[x][y][z] = v.Voxels[dx][y][dz]
		}
	}

	r.Iterate(iterator)

	return r
}

// RotateZ rotates an object around its Z axis, from the bottom
func RotateZ(v magica.VoxelObject, angle float64) (r magica.VoxelObject) {
	sin, cos := math.Sin(degToRad(angle)), math.Cos(degToRad(angle))

	orgMidpointY := float64(v.Size.Y) / 2
	orgMidpointZ := float64(v.Size.Z) / 2

	zVector := (orgMidpointY * math.Abs(sin)) + (orgMidpointZ * math.Abs(cos))
	yVector := (orgMidpointY * math.Abs(cos)) + (orgMidpointZ * math.Abs(sin))

	sizeZ, sizeY := int(math.Ceil(zVector*2)), int(math.Ceil(yVector*2))

	r = magica.VoxelObject{
		Voxels:      nil,
		PaletteData: v.PaletteData,
		Size:        geometry.Point{X: v.Size.X, Y: sizeY, Z: sizeZ},
	}

	// Create the voxel array
	r.Voxels = make([][][]byte, r.Size.X)
	for x := 0; x < r.Size.X; x++ {
		r.Voxels[x] = make([][]byte, r.Size.Y)
		for y := 0; y < r.Size.Y; y++ {
			r.Voxels[x][y] = make([]byte, r.Size.Z)
		}
	}

	vMidpointY := float64(v.Size.Y) / 2

	iterator := func(x, y, z int) {

		fdy := float64(y) - (float64(r.Size.Y) / 2)
		fdz := float64(z)

		fdy, fdz = (fdy*cos)+(fdz*sin), (fdy*-sin)+(fdz*cos)

		dy := int(math.Ceil(fdy + vMidpointY))
		dz := int(math.Ceil(fdz))

		if dy >= 0 && dz >= 0 && dy < v.Size.Y && dz < v.Size.Z {
			r.Voxels[x][y][z] = v.Voxels[x][dy][dz]
		}
	}

	r.Iterate(iterator)

	return r
}

func degToRad(angle float64) float64 {
	return (angle / 180.0) * math.Pi
}
