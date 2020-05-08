package geometry

type Point struct {
	X, Y, Z int
}

type PointWithColour struct {
	Point  Point
	Colour byte
}

type PointF struct {
	X, Y, Z float64
}

type Bounds struct {
	Min, Max Point
}

func (b *Bounds) GetSize() Point {
	return Point{X: b.Max.X - b.Min.X, Y: b.Max.Y - b.Min.Y, Z: b.Max.Z - b.Min.Z}
}
