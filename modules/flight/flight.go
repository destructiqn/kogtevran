package flight

import (
	"time"

	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/modules"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
)

type Flight struct {
	modules.DefaultModule
	Speed float64 `option:"speed"`
}

func (f *Flight) GetIdentifier() string {
	return modules.ModuleFlight
}

func (f *Flight) Toggle() (bool, error) {
	f.Enabled = !f.Enabled
	err := f.Update()
	if err != nil {
		return f.Enabled, err
	}
	return f.Enabled, nil
}

func (f *Flight) Update() error {
	flags := 0
	if f.Enabled {
		flags = 0x04
	}

	if f.Tunnel.GetPlayerHandler().IsFlying() {
		flags |= 0x02
	}

	return f.Tunnel.WriteClient(pk.Marshal(protocol.ClientboundPlayerAbilities, pk.Byte(flags), pk.Float(0.05*f.Speed), pk.Float(0.1)))
}

func HandlePlayerAbilities(_ protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleFlight) {
		go func(conn generic.Tunnel) {
			time.Sleep(100 * time.Millisecond)
			module, _ := conn.GetModuleHandler().GetModule(modules.ModuleFlight)
			err = module.(*Flight).Update()
		}(tunnel)
	}

	return generic.PassPacket(), nil
}
