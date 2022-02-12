package protocol

import (
	"github.com/Tnze/go-mc/chat"
	pk "github.com/destructiqn/kogtevran/net/packet"
)

const (
	ClientboundLoginDisconnect = iota
	ClientboundEncryptionRequest
	ClientboundLoginSuccess
	ClientboundLoginSetCompression
)

type LoginDisconnect struct {
	Reason chat.Message
}

func (d *LoginDisconnect) Read(packet pk.Packet) error {
	return packet.Scan(&d.Reason)
}

func (d *LoginDisconnect) Marshal() pk.Packet {
	return pk.Marshal(ClientboundLoginDisconnect, d.Reason)
}

type EncryptionRequest struct {
	ServerID    pk.String
	PublicKey   pk.ByteArray
	VerifyToken pk.ByteArray
}

func (e *EncryptionRequest) Read(packet pk.Packet) error {
	return packet.Scan(&e.ServerID, &e.PublicKey, &e.VerifyToken)
}

func (e *EncryptionRequest) Marshal() pk.Packet {
	return pk.Marshal(ClientboundEncryptionRequest, e.ServerID, e.PublicKey, e.VerifyToken)
}

type LoginSuccess struct {
	UUID     pk.String
	Username pk.String
}

func (l *LoginSuccess) Read(packet pk.Packet) error {
	return packet.Scan(&l.UUID, &l.Username)
}

func (l *LoginSuccess) Marshal() pk.Packet {
	return pk.Marshal(ClientboundLoginSuccess, l.UUID, l.Username)
}

type SetCompression struct {
	Threshold pk.VarInt
}

func (s *SetCompression) Read(packet pk.Packet) error {
	return packet.Scan(&s.Threshold)
}

func (s *SetCompression) Marshal() pk.Packet {
	return pk.Marshal(ClientboundLoginSetCompression, s.Threshold)
}

const (
	ServerboundLoginStart = iota
	ServerboundEncryptionResponse
)

type LoginStart struct {
	Name pk.String
}

func (l *LoginStart) Read(packet pk.Packet) error {
	return packet.Scan(&l.Name)
}

func (l *LoginStart) Marshal() pk.Packet {
	return pk.Marshal(ServerboundLoginStart, l.Name)
}

type EncryptionResponse struct {
	SharedSecret pk.ByteArray
	VerifyToken  pk.ByteArray
}

func (e *EncryptionResponse) Read(packet pk.Packet) error {
	return packet.Scan(&e.SharedSecret, &e.VerifyToken)
}

func (e *EncryptionResponse) Marshal() pk.Packet {
	return pk.Marshal(ServerboundEncryptionResponse, e.SharedSecret, e.VerifyToken)
}
