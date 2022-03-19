package nuker

import (
	"time"

	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
	"github.com/destructiqn/kogtevran/modules"
)

type Nuker struct {
	modules.SimpleTickingModule
	Radius int `option:"radius"`
}

func (n *Nuker) GetDescription() []string {
	return []string{
		"Все блоки вокруг тебя ломаются автоматически",
		"",
		"§nПараметры",
		"§7radius§f - радиус, в котором нужно ломать блоки",
		"§7interval§f - интервал между циклами ломания блоков",
	}
}

func (n *Nuker) GetIdentifier() string {
	return modules.ModuleNuker
}

func (n *Nuker) Tick() error {
	center := n.Tunnel.GetPlayerHandler().GetLocation()

	for x := int(center.X) - n.Radius; x <= int(center.X)+n.Radius; x++ {
		for y := int(center.Y) - n.Radius; y <= int(center.Y)+n.Radius; y++ {
			for z := int(center.Z) - n.Radius; z <= int(center.Z)+n.Radius; z++ {
				location := pk.Position{X: x, Y: y, Z: z}

				start := &protocol.PlayerDigging{
					Face:     1,
					Status:   0,
					Location: location,
				}

				finish := &protocol.PlayerDigging{
					Face:     1,
					Status:   2,
					Location: location,
				}

				err := n.Tunnel.WriteServer(start.Marshal())
				if err != nil {
					return err
				}

				time.Sleep(50 * time.Millisecond)

				err = n.Tunnel.WriteServer(finish.Marshal())
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
