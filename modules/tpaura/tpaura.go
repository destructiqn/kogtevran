package tpaura

import (
	"math"
	"math/rand"
	"time"

	"github.com/destructiqn/kogtevran/minecraft"
	"github.com/destructiqn/kogtevran/modules"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
)

type TPAura struct {
	modules.SimpleTickingModule
	SearchRadius   float64 `option:"searchRadius"`
	TeleportRadius float64 `option:"teleportRadius"`
	SendClient     bool    `option:"sendClient"`
	CurrentTarget  minecraft.Entity
}

func (t *TPAura) GetDescription() []string {
	return []string{
		"Автоматически телепортирует тебя к игрокам",
		"",
		"§nПараметры",
		"§7searchRadius§f - радиус поиска игроков-целей",
		"§7teleportRadius§f - расстояние, на котором ты будешь находиться после каждого телепорта",
		"§7sendClient§f - синхронизирует телепортации с клиентом",
		"§7interval§f - интервал между телепортациями",
	}
}

func (t *TPAura) Tick() error {
	playerHandler := t.Tunnel.GetPlayerHandler()
	playerLocation := playerHandler.GetLocation()
	if t.CurrentTarget == nil || t.CurrentTarget.GetLocation().Distance(playerLocation) > t.SearchRadius {
		t.CurrentTarget = t.PickEntity()
	}

	if t.CurrentTarget == nil {
		return nil
	}

	location := t.PickLocation(t.CurrentTarget.GetLocation())
	moduleHandler := t.Tunnel.GetModuleHandler()
	onGround := moduleHandler.IsModuleEnabled(modules.ModuleNoFall) || playerHandler.IsOnGround()

	if t.SendClient {
		err := t.Tunnel.WriteClient((&protocol.EntityTeleport{
			EntityID: pk.VarInt(playerHandler.GetEntityID()),
			X:        pk.Int(location.X),
			Y:        pk.Int(location.Y),
			Z:        pk.Int(location.Z),
			Yaw:      pk.Angle(location.Yaw),
			Pitch:    pk.Angle(location.Pitch),
			OnGround: pk.Boolean(onGround),
		}).Marshal())

		if err != nil {
			return err
		}
	}

	return t.Tunnel.WriteServer((&protocol.PlayerPosition{
		X:        pk.Double(location.X),
		Y:        pk.Double(location.Y),
		Z:        pk.Double(location.Z),
		OnGround: pk.Boolean(onGround),
	}).Marshal())
}

func (t *TPAura) PickLocation(center *minecraft.Location) *minecraft.Location {
	rand.Seed(time.Now().UnixNano())
	angle := float64(rand.Intn(360) - 180)
	x, z := math.Cos(angle*(math.Pi/180)), math.Sin(angle*(math.Pi/180))

	return &minecraft.Location{
		X:     center.X + x,
		Y:     center.Y,
		Z:     center.Z + z,
		Yaw:   center.Yaw,
		Pitch: center.Pitch,
	}
}

func (t *TPAura) PickEntity() minecraft.Entity {
	location := t.Tunnel.GetPlayerHandler().GetLocation()
	for _, entity := range t.Tunnel.GetEntityHandler().GetEntities() {
		if entity.GetLocation().Distance(location) > t.SearchRadius {
			continue
		}

		return entity
	}

	return nil
}

func (t *TPAura) GetIdentifier() string {
	return modules.ModuleTPAura
}
