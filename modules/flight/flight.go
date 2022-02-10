package flight

import (
	"time"

	"github.com/ruscalworld/vimeinterceptor/generic"
	"github.com/ruscalworld/vimeinterceptor/modules"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
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

func HandlePlayerAbilities(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleFlight) {
		go func(conn generic.Tunnel) {
			time.Sleep(100 * time.Millisecond)
			module, _ := conn.GetModuleHandler().GetModule(modules.ModuleFlight)
			err = module.(*Flight).Update()
		}(tunnel)
	}

	return packet.Marshal(), true, nil
}
