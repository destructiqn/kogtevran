package minecraft

import "math"

type Location struct {
	X, Y, Z    float64
	Yaw, Pitch byte
}

func (l *Location) Distance(another *Location) float64 {
	dx, dy, dz := another.X-l.X, another.Y-l.Y, another.Z-l.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}
