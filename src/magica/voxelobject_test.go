package magica

import (
	"geometry"
	"testing"
)

func TestVoxelObject_GetPoints(t *testing.T) {
	v := VoxelObject{
		Size:   geometry.Point{X: 1, Y: 1, Z: 2},
		Voxels: [][][]byte{{{254, 255}}},
	}

	points := v.GetPoints()

	if len(points) != 2 {
		t.Errorf("Expected 2 points, got %d", len(points))
	}

	if points[0].Colour != 254 {
		t.Errorf("Expected colour 254, got %d", points[0].Colour)
	}

	if points[1].Colour != 255 {
		t.Errorf("Expected colour 255, got %d", points[1].Colour)
	}

	p := geometry.Point{X: 0, Y: 0, Z: 1}
	if p != points[1].Point {
		t.Errorf("Expected location [0,0,1], got %v", points[1].Point)
	}
}
