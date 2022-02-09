package protocol

import (
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

const ServerboundHandshake = iota

type Handshake struct {
	ProtocolVersion pk.VarInt
	ServerAddress   pk.String
	ServerPort      pk.UnsignedShort
	NextState       pk.VarInt
}

func (h *Handshake) Marshal() pk.Packet {
	return pk.Marshal(ServerboundHandshake, h.ProtocolVersion, h.ServerAddress, h.ServerPort, h.NextState)
}

func (h *Handshake) Read(packet pk.Packet) error {
	return packet.Scan(&h.ProtocolVersion, &h.ServerAddress, &h.ServerPort, &h.NextState)
}
