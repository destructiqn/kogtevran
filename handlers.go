package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Tnze/go-mc/chat"
	"github.com/ruscalworld/vimeinterceptor/minecraft"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

type RawPacketHandler func(packet *Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error)
type PacketHandler func(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error)
type ProtocolStateHandler map[int32]RawPacketHandler
type ProtocolStateHandlerPool map[ConnectionState]ProtocolStateHandler

const CompressionThreshold = 1024

var (
	ServerboundHandlers = ProtocolStateHandlerPool{
		ConnStateHandshake: ProtocolStateHandler{
			protocol.ServerboundHandshake: WrapPacketHandlers(&protocol.Handshake{}, HandleHandshake),
		},

		ConnStateLogin: ProtocolStateHandler{
			protocol.ServerboundEncryptionResponse: WrapPacketHandlers(&protocol.EncryptionResponse{}, HandleEncryptionResponse),
		},

		ConnStatePlay: ProtocolStateHandler{
			protocol.ServerboundChatMessage:           WrapPacketHandlers(&protocol.ChatMessage{}, HandleChatMessage),
			protocol.ServerboundPlayer:                WrapPacketHandlers(&protocol.Player{}, HandlePlayer),
			protocol.ServerboundPlayerPosition:        WrapPacketHandlers(&protocol.PlayerPosition{}, HandlePlayerPosition),
			protocol.ServerboundPlayerPositionAndLook: WrapPacketHandlers(&protocol.ServerPlayerPositionAndLook{}, HandleServerPlayerPositionAndLook),
			protocol.ServerboundPlayerAbilities:       WrapPacketHandlers(&protocol.ServerPlayerAbilities{}, HandleServerPlayerAbilities),
		},
	}

	ClientboundHandlers = ProtocolStateHandlerPool{
		ConnStateLogin: ProtocolStateHandler{
			protocol.ClientboundEncryptionRequest:   WrapPacketHandlers(&protocol.EncryptionRequest{}, HandleEncryptionRequest),
			protocol.ClientboundLoginSuccess:        WrapPacketHandlers(&protocol.LoginSuccess{}, HandleLoginSuccess),
			protocol.ClientboundLoginSetCompression: WrapPacketHandlers(&protocol.SetCompression{}, HandleSetCompression),
		},

		ConnStatePlay: ProtocolStateHandler{
			protocol.ClientboundJoinGame:                  WrapPacketHandlers(&protocol.JoinGame{}, HandleJoinGame),
			protocol.ClientboundPlayerPositionAndLook:     WrapPacketHandlers(&protocol.PlayerPositionAndLook{}, HandlePlayerPositionAndLook),
			protocol.ClientboundSpawnPlayer:               WrapPacketHandlers(&protocol.SpawnPlayer{}, HandleSpawnPlayer),
			protocol.ClientboundSpawnMob:                  WrapPacketHandlers(&protocol.SpawnMob{}, HandleSpawnMob),
			protocol.ClientboundEntityVelocity:            WrapPacketHandlers(&protocol.EntityVelocity{}, HandleEntityVelocity),
			protocol.ClientboundDestroyEntities:           WrapPacketHandlers(&protocol.DestroyEntities{}, HandleDestroyEntities),
			protocol.ClientboundEntityRelativeMove:        WrapPacketHandlers(&protocol.EntityRelativeMove{}, HandleEntityRelativeMove),
			protocol.ClientboundEntityLookAndRelativeMove: WrapPacketHandlers(&protocol.EntityLookAndRelativeMove{}, HandleEntityLookAndRelativeMove),
			protocol.ClientboundEntityTeleport:            WrapPacketHandlers(&protocol.EntityTeleport{}, HandleEntityTeleport),
			protocol.ClientboundPlayerAbilities:           WrapPacketHandlers(&protocol.PlayerAbilities{}, HandlePlayerAbilities),
			protocol.ClientboundDisconnect:                WrapPacketHandlers(&protocol.Disconnect{}, HandleDisconnect),
		},
	}
)

func WrapPacketHandlers(packet protocol.Packet, handlers ...PacketHandler) RawPacketHandler {
	return func(rawPacket *Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
		err = packet.Read(rawPacket.Packet)
		if err != nil {
			return
		}

		for _, handler := range handlers {
			result, next, err = handler(packet, tunnel)
			if err != nil {
				return
			}
		}

		return
	}
}

func HandleHandshake(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	handshake := packet.(*protocol.Handshake)
	switch handshake.NextState {
	case 1:
		tunnel.State = ConnStateStatus
	case 2:
		tunnel.State = ConnStateLogin
	}

	sessionID := strings.Split(string(handshake.ServerAddress), ".")[0]
	CurrentTunnelPool.RegisterTunnel(sessionID, tunnel)

	host, sPort, err := net.SplitHostPort(tunnel.Server.Socket.RemoteAddr().String())
	if err != nil {
		return
	}

	port, err := strconv.Atoi(sPort)
	if err != nil {
		return
	}

	handshake.ServerAddress = pk.String(fmt.Sprintf("%s ", host))
	handshake.ServerPort = pk.UnsignedShort(port)

	return handshake.Marshal(), true, nil
}

func HandleEncryptionRequest(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	encryptionRequest := packet.(*protocol.EncryptionRequest)
	err = tunnel.WriteClient(encryptionRequest.Marshal())
	if err != nil {
		next = true
		return
	}

	<-tunnel.AuxiliaryChannelAvailable
	err = tunnel.AuxiliaryChannel.SendMessage(EncryptionDataRequest, nil)
	if err != nil {
		return
	}

	key := <-tunnel.EnableEncryptionS2C
	s2ce, s2cd := newSymmetricEncryption(key)
	tunnel.Server.SetCipher(s2ce, s2cd)
	return
}

func HandleEncryptionResponse(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	encryptionResponse := packet.(*protocol.EncryptionResponse)
	candidates := <-tunnel.EnableEncryptionC2S

	var key []byte
	found := false

	for _, candidate := range candidates {
		_, decrypt := newSymmetricEncryption(candidate)
		verifyToken := make([]byte, len(tunnel.VerifyToken))
		decrypt.XORKeyStream(verifyToken, tunnel.VerifyToken)

		if bytes.Compare(verifyToken, tunnel.VerifyToken) == 0 {
			key = candidate
			found = true
			break
		}
	}

	if !found {
		tunnel.Disconnect(chat.Text("decryption failure"))
		return
	}

	err = tunnel.WriteServer(encryptionResponse.Marshal())
	if err != nil {
		return
	}

	c2se, c2sd := newSymmetricEncryption(key)
	tunnel.Client.SetCipher(c2se, c2sd)
	tunnel.EnableEncryptionS2C <- key

	err = tunnel.WriteClient((&protocol.SetCompression{Threshold: CompressionThreshold}).Marshal())
	if err != nil {
		return
	}

	tunnel.Client.SetThreshold(CompressionThreshold)
	return
}

func HandleSetCompression(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	setCompression := packet.(*protocol.SetCompression)
	tunnel.Server.SetThreshold(int(setCompression.Threshold))
	return
}

func HandleLoginSuccess(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	tunnel.State = ConnStatePlay
	return packet.Marshal(), true, nil
}

func HandleJoinGame(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	joinGame := packet.(*protocol.JoinGame)
	tunnel.EntityID = int32(joinGame.EntityID)
	tunnel.resetEntities()
	return packet.Marshal(), true, nil
}

func HandlePlayerPositionAndLook(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	playerPositionAndLook := packet.(*protocol.PlayerPositionAndLook)
	flags := playerPositionAndLook.Flags
	x, y, z := playerPositionAndLook.X, playerPositionAndLook.Y, playerPositionAndLook.Z
	yaw, pitch := playerPositionAndLook.Yaw, playerPositionAndLook.Pitch

	if flags&0x01 > 0 {
		tunnel.Location.X += float64(x)
	} else {
		tunnel.Location.X = float64(x)
	}

	if flags&0x02 > 0 {
		tunnel.Location.Y += float64(y)
	} else {
		tunnel.Location.Y = float64(y)
	}

	if flags&0x04 > 0 {
		tunnel.Location.Z += float64(z)
	} else {
		tunnel.Location.Z = float64(z)
	}

	tunnel.Location.Yaw, tunnel.Location.Pitch = byte(yaw), byte(pitch)
	return packet.Marshal(), true, nil
}

func HandleSpawnPlayer(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
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

	tunnel.initPlayer(int(spawnPlayer.EntityID), player)
	return pk.Packet{}, true, nil
}

func HandleSpawnMob(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
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

	tunnel.initMob(int(spawnMob.EntityID), mob)
	return pk.Packet{}, true, nil
}

func HandleEntityVelocity(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	entityVelocity := packet.(*protocol.EntityVelocity)
	antiKnockback := tunnel.Modules[ModuleAntiKnockback].(*AntiKnockback)

	if tunnel.IsModuleEnabled(ModuleAntiKnockback) && int32(entityVelocity.EntityID) == tunnel.EntityID {
		entityVelocity.VX = pk.Short(antiKnockback.X)
		entityVelocity.VY = pk.Short(antiKnockback.Y)
		entityVelocity.VZ = pk.Short(antiKnockback.Z)
	}

	return entityVelocity.Marshal(), true, nil
}

func HandleDestroyEntities(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	destroyEntities := packet.(*protocol.DestroyEntities)
	entityIDs := make([]int, 0)
	for _, entityID := range destroyEntities.EntityIDs {
		entityIDs = append(entityIDs, int(entityID))
	}

	tunnel.destroyEntities(entityIDs)
	return packet.Marshal(), true, nil
}

func HandleEntityRelativeMove(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	entityRelativeMove := packet.(*protocol.EntityRelativeMove)
	dx, dy, dz := entityRelativeMove.DX, entityRelativeMove.DY, entityRelativeMove.DZ
	tunnel.entityRelativeMove(int(entityRelativeMove.EntityID), float64(dx)/32, float64(dy)/32, float64(dz)/32)
	return packet.Marshal(), true, nil
}

func HandleEntityLookAndRelativeMove(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	entityLookAndRelativeMove := packet.(*protocol.EntityLookAndRelativeMove)
	dx, dy, dz := entityLookAndRelativeMove.DX, entityLookAndRelativeMove.DY, entityLookAndRelativeMove.DZ
	tunnel.entityRelativeMove(int(entityLookAndRelativeMove.EntityID), float64(dx)/32, float64(dy)/32, float64(dz)/32)
	return packet.Marshal(), true, nil
}

func HandleEntityTeleport(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	entityTeleport := packet.(*protocol.EntityTeleport)
	x, y, z := entityTeleport.X, entityTeleport.Y, entityTeleport.Z
	yaw, pitch := entityTeleport.Yaw, entityTeleport.Pitch
	tunnel.entityTeleport(int(entityTeleport.EntityID), float64(x)/32, float64(y)/32, float64(z)/32, byte(yaw), byte(pitch))
	return packet.Marshal(), true, nil
}

func HandlePlayerAbilities(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	if tunnel.IsModuleEnabled(ModuleFlight) {
		go func(conn *MinecraftTunnel) {
			time.Sleep(100 * time.Millisecond)
			flight := conn.Modules[ModuleFlight].(*Flight)
			err = flight.Update()
		}(tunnel)
	}

	return packet.Marshal(), true, nil
}

func HandleDisconnect(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	disconnect := packet.(*protocol.Disconnect)
	log.Println("disconnected from server:", disconnect.Reason.String())
	err = tunnel.WriteClient(packet.Marshal())
	if err != nil {
		return
	}

	tunnel.Close()
	return
}

func HandleChatMessage(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	chatMessage := packet.(*protocol.ChatMessage)
	handled := HandleCommand(string(chatMessage.Message), tunnel)
	return packet.Marshal(), !handled, nil
}

func HandlePlayer(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	player := packet.(*protocol.Player)
	if tunnel.IsModuleEnabled(ModuleNoFall) {
		player.OnGround = true
	}
	return player.Marshal(), true, nil
}

func HandlePlayerPosition(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	playerPosition := packet.(*protocol.PlayerPosition)
	location := tunnel.Location
	location.X, location.Y, location.Z = float64(playerPosition.X), float64(playerPosition.Y), float64(playerPosition.Z)

	if tunnel.IsModuleEnabled(ModuleNoFall) {
		playerPosition.OnGround = true
	}

	return playerPosition.Marshal(), true, nil
}

func HandleServerPlayerPositionAndLook(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	playerPosition := packet.(*protocol.ServerPlayerPositionAndLook)
	location := tunnel.Location
	location.X, location.Y, location.Z = float64(playerPosition.X), float64(playerPosition.Y), float64(playerPosition.Z)
	tunnel.Location.Yaw, tunnel.Location.Pitch = byte(playerPosition.Yaw), byte(playerPosition.Pitch)

	if tunnel.IsModuleEnabled(ModuleNoFall) {
		playerPosition.OnGround = true
	}

	return playerPosition.Marshal(), true, nil
}

func HandleServerPlayerAbilities(packet protocol.Packet, tunnel *MinecraftTunnel) (result pk.Packet, next bool, err error) {
	playerAbilities := packet.(*protocol.ServerPlayerAbilities)
	tunnel.IsFlying = playerAbilities.Flags&0x02 > 0
	return packet.Marshal(), true, nil
}

func HandlePluginMessage(packet pk.Packet, srcName string) error {
	var (
		Channel pk.String
		Data    pk.PluginMessageData
	)

	err := packet.Scan(&Channel, &Data)
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("accepted plugin message from %s in channel %s:\n%s", srcName, Channel, hex.Dump(Data)))
	return nil
}
