package flight

import (
	"time"

	"github.com/destructiqn/kogtevran/generic"
	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
	"github.com/destructiqn/kogtevran/modules"
)

type Flight struct {
	modules.DefaultModule
	Speed float64 `option:"speed"`
}

func (f *Flight) GetDescription() []string {
	return []string{
		"Летай везде, где душе угодно",
		"Чтобы не кикало, нужно прижиматься к стенам, либо периодически отключать режим полёта",
		"",
		"§nПараметры",
		"§7speed§f - скорость полёта",
	}
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
		flags |= 0x04
	}

	if f.Tunnel.GetPlayerHandler().IsFlying() && f.Enabled {
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
