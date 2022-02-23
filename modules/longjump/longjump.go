package longjump

import (
	"math"

	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/modules"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
)

type LongJump struct {
	modules.SimpleModule
	Power int `option:"power"`
}

func (l *LongJump) GetIdentifier() string {
	return modules.ModuleLongJump
}

func (l *LongJump) Boost(yaw float64) (x int16, y int16, z int16) {
	x = int16(-math.Sin(float64(yaw)*(math.Pi/180.0)) * float64(l.Power*5000))
	y = int16(5000)
	z = int16(math.Cos(float64(yaw)*(math.Pi/180.0)) * float64(l.Power*5000))
	return
}

func HandlePlayerPosition(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	playerPosition := packet.(*protocol.PlayerPosition)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleLongJump) {
		err = HandleJumpCandidate(bool(playerPosition.OnGround), float64(playerPosition.Y), tunnel)
		if err != nil {
			return
		}
	}

	return generic.PassPacket(), nil
}

func HandleServerPlayerPositionAndLook(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	playerPosition := packet.(*protocol.ServerPlayerPositionAndLook)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleLongJump) {
		err = HandleJumpCandidate(bool(playerPosition.OnGround), float64(playerPosition.Y), tunnel)
		if err != nil {
			return
		}
	}

	return generic.PassPacket(), nil
}

func HandleJumpCandidate(onGround bool, y float64, tunnel generic.Tunnel) error {
	location := tunnel.GetPlayerHandler().GetLocation()
	if !onGround && tunnel.GetPlayerHandler().IsOnGround() && y > location.Y {
		module, ok := tunnel.GetModuleHandler().GetModule(modules.ModuleLongJump)
		if ok {
			longJump := module.(*LongJump)
			x, y, z := longJump.Boost(location.Yaw)
			velocityPacket := protocol.EntityVelocity{
				EntityID: pk.VarInt(tunnel.GetPlayerHandler().GetEntityID()),
				VX:       pk.Short(x),
				VY:       pk.Short(y),
				VZ:       pk.Short(z),
			}

			return tunnel.WriteClient(velocityPacket.Marshal())
		}
	}

	return nil
}