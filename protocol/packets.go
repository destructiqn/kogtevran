package protocol

import (
	"fmt"
	"strconv"

	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

const (
	ConnS2C = 1
	ConnC2S = 2
)

type WrappedPacket struct {
	Name string
	pk.Packet
}

type PacketMap map[int]WrappedPacket

var (
	PlayPacketsS2C = PacketMap{
		0x00: {Name: "Keep Alive"},
		0x01: {Name: "Join Game"},
		0x02: {Name: "Chat Message"},
		0x03: {Name: "Time Update"},
		0x04: {Name: "Entity Equipment"},
		0x05: {Name: "Spawn Position"},
		0x06: {Name: "Update Health"},
		0x07: {Name: "Respawn"},
		0x08: {Name: "Player Position And Look"},
		0x09: {Name: "Held Item Change"},
		0x0A: {Name: "Use Bed"},
		0x0B: {Name: "Animation"},
		0x0C: {Name: "Spawn Player"},
		0x0D: {Name: "Collect Item"},
		0x0E: {Name: "Spawn Object"},
		0x0F: {Name: "Spawn Mob"},
		0x10: {Name: "Spawn Painting"},
		0x11: {Name: "Spawn Experience Orb"},
		0x12: {Name: "Entity Velocity"},
		0x13: {Name: "Destroy Entities"},
		0x14: {Name: "Entity"},
		0x15: {Name: "Entity Relative Move"},
		0x16: {Name: "Entity Look"},
		0x17: {Name: "Entity Look And Relative Move"},
		0x18: {Name: "Entity Teleport"},
		0x19: {Name: "Entity Head Look"},
		0x1A: {Name: "Entity Status"},
		0x1B: {Name: "Attach Entity"},
		0x1C: {Name: "Entity Metadata"},
		0x1D: {Name: "Entity Effect"},
		0x1E: {Name: "Remove Entity Effect"},
		0x1F: {Name: "Set Experience"},
		0x20: {Name: "Entity Properties"},
		0x21: {Name: "Chunk Data"},
		0x22: {Name: "Multi Block Change"},
		0x23: {Name: "Block Change"},
		0x24: {Name: "Block Action"},
		0x25: {Name: "Block Break Animation"},
		0x26: {Name: "Map Chunk Bulk"},
		0x27: {Name: "Explosion"},
		0x28: {Name: "Effect"},
		0x29: {Name: "Sound Effect"},
		0x2A: {Name: "Particle"},
		0x2B: {Name: "Change Game State"},
		0x2C: {Name: "Spawn Global Entity"},
		0x2D: {Name: "Open Window"},
		0x2E: {Name: "Close Window"},
		0x2F: {Name: "Set Slot"},
		0x30: {Name: "Window Items"},
		0x31: {Name: "Window Property"},
		0x32: {Name: "Confirm Transaction"},
		0x33: {Name: "Update Sign"},
		0x34: {Name: "Map"},
		0x35: {Name: "Update Block Entity"},
		0x36: {Name: "Open Sign Editor"},
		0x37: {Name: "Statistics"},
		0x38: {Name: "Player List Item"},
		0x39: {Name: "Player Abilities"},
		0x3A: {Name: "Tab-Complete"},
		0x3B: {Name: "Scoreboard Objective"},
		0x3C: {Name: "Update Score"},
		0x3D: {Name: "Display Scoreboard"},
		0x3E: {Name: "Teams"},
		0x3F: {Name: "Plugin Message"},
		0x40: {Name: "Disconnect"},
		0x41: {Name: "Server Difficulty"},
		0x42: {Name: "Combat Event"},
		0x43: {Name: "Camera"},
		0x44: {Name: "World Border"},
		0x45: {Name: "Title"},
		0x46: {Name: "Set Compression"}, // Broken
		0x47: {Name: "Player List Header And Footer"},
		0x48: {Name: "Resource Pack Send"},
		0x49: {Name: "Update Entity NBT"},
	}

	PlayPacketsC2S = PacketMap{
		0x00: {Name: "Keep Alive"},
		0x01: {Name: "Chat Message"},
		0x02: {Name: "Use Entity"},
		0x03: {Name: "Player"},
		0x04: {Name: "Player Position"},
		0x05: {Name: "Player Look"},
		0x06: {Name: "Player Position And Look"},
		0x07: {Name: "Player Digging"},
		0x08: {Name: "Player Block Placement"},
		0x09: {Name: "Held Item Change"},
		0x0A: {Name: "Animation"},
		0x0B: {Name: "Entity Action"},
		0x0C: {Name: "Steer Vehicle"},
		0x0D: {Name: "Close Window"},
		0x0E: {Name: "Click Window"},
		0x0F: {Name: "Confirm Transaction"},
		0x10: {Name: "Creative Inventory Action"},
		0x11: {Name: "Enchant Item"},
		0x12: {Name: "Update Sign"},
		0x13: {Name: "Player Abilities"},
		0x14: {Name: "Tab-Complete"},
		0x15: {Name: "Client Settings"},
		0x16: {Name: "Client Status"},
		0x17: {Name: "Plugin Message"},
		0x18: {Name: "Spectate"},
		0x19: {Name: "Resource Pack Status"},
	}
)

func GetPacketMap(connType int) PacketMap {
	switch connType {
	case ConnC2S:
		return PlayPacketsC2S
	case ConnS2C:
		return PlayPacketsS2C
	}

	panic("unsupported direction")
}

func GetPacketDescription(packetID, connType int) (packet WrappedPacket, ok bool) {
	packet, ok = GetPacketMap(connType)[packetID]
	return
}

func GetPacketName(id int32, connType int) string {
	nameDisplay := "Unsupported Packet"
	packet, ok := GetPacketDescription(int(id), connType)
	if ok {
		nameDisplay = packet.Name
	}
	return nameDisplay
}

func FormatPacket(id int32, connType int) string {
	return fmt.Sprintf("%s %s", strconv.FormatInt(int64(id), 16), GetPacketName(id, connType))
}

func WrapPacket(packet pk.Packet, connType int) *WrappedPacket {
	return &WrappedPacket{
		Name:   GetPacketName(packet.ID, connType),
		Packet: packet,
	}
}
