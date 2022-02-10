package proxy

import (
	"sync"

	"github.com/ruscalworld/vimeinterceptor/generic"
	"github.com/ruscalworld/vimeinterceptor/minecraft"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

type EntityHandler struct {
	Entities map[int]minecraft.Entity
	sync.Mutex
}

func NewEntityHandler() *EntityHandler {
	return &EntityHandler{Entities: make(map[int]minecraft.Entity)}
}

func (h *EntityHandler) GetEntities() map[int]minecraft.Entity {
	return h.Entities
}

func (h *EntityHandler) GetEntity(entityID int) (minecraft.Entity, bool) {
	entity, ok := h.Entities[entityID]
	return entity, ok
}

func (h *EntityHandler) InitPlayer(entityID int, player *minecraft.Player) {
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

func (h *EntityHandler) EntityTeleport(entityID int, x, y, z float64, yaw, pitch byte) {
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

func HandleSpawnPlayer(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	spawnPlayer := packet.(*protocol.SpawnPlayer)

	player := &minecraft.Player{
		DefaultEntity: minecraft.DefaultEntity{Location: &minecraft.Location{
			X:     float64(spawnPlayer.X) / 32,
			Y:     float64(spawnPlayer.Y) / 32,
			Z:     float64(spawnPlayer.Z) / 32,
			Yaw:   byte(spawnPlayer.Yaw),
			Pitch: byte(spawnPlayer.Pitch),
		}},
	}

	tunnel.GetEntityHandler().InitPlayer(int(spawnPlayer.EntityID), player)
	return pk.Packet{}, true, nil
}

func HandleSpawnMob(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	spawnMob := packet.(*protocol.SpawnMob)
	mob := &minecraft.Mob{
		DefaultEntity: minecraft.DefaultEntity{Location: &minecraft.Location{
			X:     float64(spawnMob.X) / 32,
			Y:     float64(spawnMob.Y) / 32,
			Z:     float64(spawnMob.Z) / 32,
			Yaw:   byte(spawnMob.Yaw),
			Pitch: byte(spawnMob.Pitch),
		}},
		Type: minecraft.MobType(spawnMob.Type),
	}

	tunnel.GetEntityHandler().InitMob(int(spawnMob.EntityID), mob)
	return pk.Packet{}, true, nil
}

func HandleDestroyEntities(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	destroyEntities := packet.(*protocol.DestroyEntities)
	entityIDs := make([]int, 0)
	for _, entityID := range destroyEntities.EntityIDs {
		entityIDs = append(entityIDs, int(entityID))
	}

	tunnel.GetEntityHandler().DestroyEntities(entityIDs)
	return packet.Marshal(), true, nil
}

func HandleEntityRelativeMove(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	entityRelativeMove := packet.(*protocol.EntityRelativeMove)
	dx, dy, dz := entityRelativeMove.DX, entityRelativeMove.DY, entityRelativeMove.DZ
	tunnel.GetEntityHandler().EntityRelativeMove(int(entityRelativeMove.EntityID), float64(dx)/32, float64(dy)/32, float64(dz)/32)
	return packet.Marshal(), true, nil
}

func HandleEntityLookAndRelativeMove(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	entityLookAndRelativeMove := packet.(*protocol.EntityLookAndRelativeMove)
	dx, dy, dz := entityLookAndRelativeMove.DX, entityLookAndRelativeMove.DY, entityLookAndRelativeMove.DZ
	tunnel.GetEntityHandler().EntityRelativeMove(int(entityLookAndRelativeMove.EntityID), float64(dx)/32, float64(dy)/32, float64(dz)/32)
	return packet.Marshal(), true, nil
}

func HandleEntityTeleport(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	entityTeleport := packet.(*protocol.EntityTeleport)
	x, y, z := entityTeleport.X, entityTeleport.Y, entityTeleport.Z
	yaw, pitch := entityTeleport.Yaw, entityTeleport.Pitch
	tunnel.GetEntityHandler().EntityTeleport(int(entityTeleport.EntityID), float64(x)/32, float64(y)/32, float64(z)/32, byte(yaw), byte(pitch))
	return packet.Marshal(), true, nil
}
