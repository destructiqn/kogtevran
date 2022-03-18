package proxy

import (
	"sync"

	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/minecraft"
	"github.com/destructiqn/kogtevran/protocol"
)

type EntityHandler struct {
	tunnel   *MinecraftTunnel
	Entities map[int]minecraft.Entity
	sync.Mutex
}

func NewEntityHandler(tunnel *MinecraftTunnel) *EntityHandler {
	return &EntityHandler{tunnel: tunnel, Entities: make(map[int]minecraft.Entity)}
}

func (h *EntityHandler) GetEntities() map[int]minecraft.Entity {
	return h.Entities
}

func (h *EntityHandler) GetEntity(entityID int) (minecraft.Entity, bool) {
	entity, ok := h.Entities[entityID]
	return entity, ok
}

func (h *EntityHandler) InitPlayer(entityID int, player *minecraft.Player) {
	if h.tunnel.GetPlayerHandler().GetEntityID() == int32(entityID) {
		return
	}

	h.Lock()
	h.Entities[entityID] = player
	h.Unlock()
}

func (h *EntityHandler) InitMob(entityID int, mob *minecraft.Mob) {
	h.Lock()
	h.Entities[entityID] = mob
	h.Unlock()
}

func (h *EntityHandler) EntityRelativeMove(entityID int, dx, dy, dz float64) {
	entity, ok := h.Entities[entityID]
	if !ok {
		return
	}

	entity.GetLocation().X += dx
	entity.GetLocation().Y += dy
	entity.GetLocation().Z += dz
}

func (h *EntityHandler) EntityTeleport(entityID int, x, y, z, yaw, pitch float64) {
	entity, ok := h.Entities[entityID]
	if !ok {
		return
	}

	entity.GetLocation().X, entity.GetLocation().Y, entity.GetLocation().Z = x, y, z
	entity.GetLocation().Yaw, entity.GetLocation().Pitch = yaw, pitch
}

func (h *EntityHandler) ResetEntities() {
	h.Lock()
	for id := range h.Entities {
		delete(h.Entities, id)
	}
	h.Unlock()
}

func (h *EntityHandler) DestroyEntities(entityIDs []int) {
	h.Lock()
	for _, id := range entityIDs {
		delete(h.Entities, id)
	}
	h.Unlock()
}

func HandleSpawnPlayer(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	spawnPlayer := packet.(*protocol.SpawnPlayer)

	player := &minecraft.Player{
		DefaultEntity: minecraft.DefaultEntity{Location: &minecraft.Location{
			X:     float64(spawnPlayer.X) / 32,
			Y:     float64(spawnPlayer.Y) / 32,
			Z:     float64(spawnPlayer.Z) / 32,
			Yaw:   float64(spawnPlayer.Yaw),
			Pitch: float64(spawnPlayer.Pitch),
		}},
	}

	tunnel.GetEntityHandler().InitPlayer(int(spawnPlayer.EntityID), player)
	return generic.PassPacket(), nil
}

func HandleSpawnMob(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	spawnMob := packet.(*protocol.SpawnMob)
	mob := &minecraft.Mob{
		DefaultEntity: minecraft.DefaultEntity{Location: &minecraft.Location{
			X:     float64(spawnMob.X) / 32,
			Y:     float64(spawnMob.Y) / 32,
			Z:     float64(spawnMob.Z) / 32,
			Yaw:   float64(spawnMob.Yaw),
			Pitch: float64(spawnMob.Pitch),
		}},
		Type: minecraft.MobType(spawnMob.Type),
	}

	tunnel.GetEntityHandler().InitMob(int(spawnMob.EntityID), mob)
	return generic.PassPacket(), nil
}

func HandleDestroyEntities(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	destroyEntities := packet.(*protocol.DestroyEntities)
	entityIDs := make([]int, 0)
	for _, entityID := range destroyEntities.EntityIDs {
		entityIDs = append(entityIDs, int(entityID))
	}

	tunnel.GetEntityHandler().DestroyEntities(entityIDs)
	return generic.PassPacket(), nil
}

func HandleEntityRelativeMove(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	entityRelativeMove := packet.(*protocol.EntityRelativeMove)
	dx, dy, dz := entityRelativeMove.DX, entityRelativeMove.DY, entityRelativeMove.DZ
	tunnel.GetEntityHandler().EntityRelativeMove(int(entityRelativeMove.EntityID), float64(dx)/32, float64(dy)/32, float64(dz)/32)
	return generic.PassPacket(), nil
}

func HandleEntityLookAndRelativeMove(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	entityLookAndRelativeMove := packet.(*protocol.EntityLookAndRelativeMove)
	dx, dy, dz := entityLookAndRelativeMove.DX, entityLookAndRelativeMove.DY, entityLookAndRelativeMove.DZ
	tunnel.GetEntityHandler().EntityRelativeMove(int(entityLookAndRelativeMove.EntityID), float64(dx)/32, float64(dy)/32, float64(dz)/32)
	return generic.PassPacket(), nil
}

func HandleEntityTeleport(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	entityTeleport := packet.(*protocol.EntityTeleport)
	x, y, z := entityTeleport.X, entityTeleport.Y, entityTeleport.Z
	yaw, pitch := entityTeleport.Yaw, entityTeleport.Pitch
	tunnel.GetEntityHandler().EntityTeleport(int(entityTeleport.EntityID), float64(x)/32, float64(y)/32, float64(z)/32, float64(yaw), float64(pitch))
	return generic.PassPacket(), nil
}
