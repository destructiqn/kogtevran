package minecraft

import (
	"math"

	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
)

type Location struct {
	X, Y, Z    float64
	Yaw, Pitch float64
}

func (l *Location) Distance(another *Location) float64 {
	dx, dy, dz := another.X-l.X, another.Y-l.Y, another.Z-l.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func (l *Location) ToPosition() pk.Position {
	return pk.Position{X: int(l.X), Y: int(l.Y), Z: int(l.Z)}
}
