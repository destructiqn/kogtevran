package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/Tnze/go-mc/chat"
	"github.com/ruscalworld/vimeinterceptor/generic"
	"github.com/ruscalworld/vimeinterceptor/modules/antiknockback"
	"github.com/ruscalworld/vimeinterceptor/modules/flight"
	"github.com/ruscalworld/vimeinterceptor/modules/nofall"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
	"github.com/ruscalworld/vimeinterceptor/proxy"
)

type RawPacketHandler func(packet *protocol.WrappedPacket, tunnel *proxy.MinecraftTunnel) (result pk.Packet, next bool, err error)
type PacketHandler func(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error)
type ProtocolStateHandler map[int32]RawPacketHandler
type ProtocolStateHandlerPool map[protocol.ConnectionState]ProtocolStateHandler
type PluginMessageHandler func(data []byte, tunnel generic.Tunnel) (result []byte, next bool, err error)

const CompressionThreshold = 1024

var (
	ServerboundHandlers = ProtocolStateHandlerPool{
		protocol.ConnStateHandshake: ProtocolStateHandler{
			protocol.ServerboundHandshake: WrapPacketHandlers(&protocol.Handshake{}, HandleHandshake),
		},

		protocol.ConnStateLogin: ProtocolStateHandler{
			protocol.ServerboundEncryptionResponse: WrapPacketHandlers(&protocol.EncryptionResponse{}, HandleEncryptionResponse),
		},

		protocol.ConnStatePlay: ProtocolStateHandler{
			protocol.ServerboundChatMessage: WrapPacketHandlers(&protocol.ChatMessage{},
				proxy.HandleChatMessage,
			),
			protocol.ServerboundPlayer: WrapPacketHandlers(&protocol.Player{},
				nofall.HandlePlayer,
			),
			protocol.ServerboundPlayerPosition: WrapPacketHandlers(&protocol.PlayerPosition{},
				nofall.HandlePlayerPosition, proxy.HandlePlayerPosition,
			),
			protocol.ServerboundPlayerPositionAndLook: WrapPacketHandlers(&protocol.ServerPlayerPositionAndLook{},
				nofall.HandleServerPlayerPositionAndLook, proxy.HandleServerPlayerPositionAndLook,
			),
			protocol.ServerboundCloseWindow: WrapPacketHandlers(&protocol.ServerCloseWindow{},
				proxy.HandleCloseWindow,
			),
			protocol.ServerboundPlayerAbilities: WrapPacketHandlers(&protocol.ServerPlayerAbilities{},
				proxy.HandleServerPlayerAbilities,
			),
		},
	}

	ClientboundHandlers = ProtocolStateHandlerPool{
		protocol.ConnStateLogin: ProtocolStateHandler{
			protocol.ClientboundEncryptionRequest:   WrapPacketHandlers(&protocol.EncryptionRequest{}, HandleEncryptionRequest),
			protocol.ClientboundLoginSuccess:        WrapPacketHandlers(&protocol.LoginSuccess{}, HandleLoginSuccess),
			protocol.ClientboundLoginSetCompression: WrapPacketHandlers(&protocol.SetCompression{}, HandleSetCompression),
		},

		protocol.ConnStatePlay: ProtocolStateHandler{
			protocol.ClientboundJoinGame: WrapPacketHandlers(&protocol.JoinGame{},
				proxy.HandleJoinGame,
			),
			protocol.ClientboundPlayerPositionAndLook: WrapPacketHandlers(&protocol.PlayerPositionAndLook{},
				proxy.HandlePlayerPositionAndLook,
			),
			protocol.ClientboundSpawnPlayer: WrapPacketHandlers(&protocol.SpawnPlayer{},
				proxy.HandleSpawnPlayer,
			),
			protocol.ClientboundSpawnMob: WrapPacketHandlers(&protocol.SpawnMob{},
				proxy.HandleSpawnMob,
			),
			protocol.ClientboundEntityVelocity: WrapPacketHandlers(&protocol.EntityVelocity{},
				antiknockback.HandleEntityVelocity,
			),
			protocol.ClientboundDestroyEntities: WrapPacketHandlers(&protocol.DestroyEntities{},
				proxy.HandleDestroyEntities,
			),
			protocol.ClientboundEntityRelativeMove: WrapPacketHandlers(&protocol.EntityRelativeMove{},
				proxy.HandleEntityRelativeMove,
			),
			protocol.ClientboundEntityLookAndRelativeMove: WrapPacketHandlers(&protocol.EntityLookAndRelativeMove{},
				proxy.HandleEntityLookAndRelativeMove,
			),
			protocol.ClientboundEntityTeleport: WrapPacketHandlers(&protocol.EntityTeleport{},
				proxy.HandleEntityTeleport,
			),
			protocol.ClientboundOpenWindow: WrapPacketHandlers(&protocol.OpenWindow{},
				proxy.HandleOpenWindow,
			),
			protocol.ClientboundCloseWindow: WrapPacketHandlers(&protocol.CloseWindow{},
				proxy.HandleCloseWindow,
			),
			protocol.ClientboundSetSlot: WrapPacketHandlers(&protocol.SetSlot{},
				proxy.HandleSetSlot,
			),
			protocol.ClientboundWindowItems: WrapPacketHandlers(&protocol.WindowItems{},
				proxy.HandleWindowItems,
			),
			protocol.ClientboundPlayerAbilities: WrapPacketHandlers(&protocol.PlayerAbilities{},
				flight.HandlePlayerAbilities,
			),
			//protocol.ClientboundPluginMessage: WrapPacketHandlers(&protocol.PluginMessage{},
			//	HandlePluginMessage("Texteria", proxy.HandleClientboundTexteriaPacket),
			//),
			protocol.ClientboundDisconnect: WrapPacketHandlers(&protocol.Disconnect{},
				HandleDisconnect,
			),
		},
	}
)

func WrapPacketHandlers(packet protocol.Packet, handlers ...PacketHandler) RawPacketHandler {
	return func(rawPacket *protocol.WrappedPacket, tunnel *proxy.MinecraftTunnel) (result pk.Packet, next bool, err error) {
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

func HandleHandshake(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	handshake := packet.(*protocol.Handshake)
	switch handshake.NextState {
	case 1:
		tunnel.SetState(protocol.ConnStateStatus)
	case 2:
		tunnel.SetState(protocol.ConnStateLogin)
	}

	sessionID := strings.Split(string(handshake.ServerAddress), ".")[0]
	proxy.CurrentTunnelPool.RegisterTunnel(sessionID, tunnel.(*proxy.MinecraftTunnel))

	host, sPort, err := net.SplitHostPort(tunnel.(*proxy.MinecraftTunnel).Server.Socket.RemoteAddr().String())
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

func HandleEncryptionRequest(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	minecraftTunnel := tunnel.(*proxy.MinecraftTunnel)
	encryptionRequest := packet.(*protocol.EncryptionRequest)
	err = tunnel.WriteClient(encryptionRequest.Marshal())
	if err != nil {
		next = true
		return
	}

	<-minecraftTunnel.AuxiliaryChannelAvailable
	err = minecraftTunnel.AuxiliaryChannel.SendMessage(proxy.EncryptionDataRequest, nil)
	if err != nil {
		return
	}

	key := <-minecraftTunnel.EnableEncryptionS2C
	s2ce, s2cd := newSymmetricEncryption(key)
	minecraftTunnel.Server.SetCipher(s2ce, s2cd)
	return
}

func HandleEncryptionResponse(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	minecraftTunnel := tunnel.(*proxy.MinecraftTunnel)
	encryptionResponse := packet.(*protocol.EncryptionResponse)
	candidates := <-minecraftTunnel.EnableEncryptionC2S

	var key []byte
	found := false

	for _, candidate := range candidates {
		_, decrypt := newSymmetricEncryption(candidate)
		verifyToken := make([]byte, len(minecraftTunnel.VerifyToken))
		decrypt.XORKeyStream(verifyToken, minecraftTunnel.VerifyToken)

		if bytes.Compare(verifyToken, minecraftTunnel.VerifyToken) == 0 {
			key = candidate
			found = true
			break
		}
	}

	if !found {
		minecraftTunnel.Disconnect(chat.Text("decryption failure"))
		return
	}

	err = tunnel.WriteServer(encryptionResponse.Marshal())
	if err != nil {
		return
	}

	c2se, c2sd := newSymmetricEncryption(key)
	minecraftTunnel.Client.SetCipher(c2se, c2sd)
	minecraftTunnel.EnableEncryptionS2C <- key

	err = tunnel.WriteClient((&protocol.SetCompression{Threshold: CompressionThreshold}).Marshal())
	if err != nil {
		return
	}

	minecraftTunnel.Client.SetThreshold(CompressionThreshold)
	return
}

func HandleSetCompression(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	minecraftTunnel := tunnel.(*proxy.MinecraftTunnel)
	setCompression := packet.(*protocol.SetCompression)
	minecraftTunnel.Server.SetThreshold(int(setCompression.Threshold))
	return
}

func HandleLoginSuccess(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	tunnel.SetState(protocol.ConnStatePlay)
	return packet.Marshal(), true, nil
}

func HandleDisconnect(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	disconnect := packet.(*protocol.Disconnect)
	log.Println("disconnected from server:", disconnect.Reason.String())
	err = tunnel.WriteClient(packet.Marshal())
	if err != nil {
		return
	}

	tunnel.Close()
	return
}

func HandlePluginMessage(targetChannel string, handler PluginMessageHandler) PacketHandler {
	return func(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
		var (
			data     []byte
			channel  string
			packetID int32
		)

		switch packet.(type) {
		case *protocol.PluginMessage:
			pluginMessage := packet.(*protocol.PluginMessage)
			data, channel = pluginMessage.Data, string(pluginMessage.Channel)
			packetID = protocol.ClientboundPluginMessage
		case *protocol.ServerPluginMessage:
			pluginMessage := packet.(*protocol.ServerPluginMessage)
			data, channel = pluginMessage.Data, string(pluginMessage.Channel)
			packetID = protocol.ServerboundPluginMessage
		}

		if targetChannel == channel {
			data, next, err = handler(data, tunnel)
			messageData := pk.PluginMessageData(data)
			return pk.Marshal(packetID, pk.String(channel), &messageData), next, err
		}

		return packet.Marshal(), true, nil
	}
}
