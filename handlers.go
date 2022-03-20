package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/Tnze/go-mc/chat"
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/metrics"
	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
	"github.com/destructiqn/kogtevran/proxy"

	"github.com/destructiqn/kogtevran/modules/antiknockback"
	"github.com/destructiqn/kogtevran/modules/fastbreak"
	"github.com/destructiqn/kogtevran/modules/flight"
	"github.com/destructiqn/kogtevran/modules/longjump"
	"github.com/destructiqn/kogtevran/modules/nobadeffects"
	"github.com/destructiqn/kogtevran/modules/nofall"
	"github.com/destructiqn/kogtevran/modules/unlimitedcps"

	"github.com/prometheus/client_golang/prometheus"
)

type RawPacketHandler func(packet *protocol.WrappedPacket, tunnel *proxy.MinecraftTunnel) (result *generic.HandlerResult, err error)
type PacketHandler func(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error)
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
			protocol.ServerboundLoginStart:         WrapPacketHandlers(&protocol.LoginStart{}, HandleLoginStart),
			protocol.ServerboundEncryptionResponse: WrapPacketHandlers(&protocol.EncryptionResponse{}, HandleEncryptionResponse),
		},

		protocol.ConnStatePlay: ProtocolStateHandler{
			protocol.ServerboundChatMessage: WrapPacketHandlers(&protocol.ChatMessage{},
				proxy.HandleChatMessage,
			),
			protocol.ServerboundPlayer: WrapPacketHandlers(&protocol.Player{},
				proxy.HandlePlayer, nofall.HandlePlayer,
			),
			protocol.ServerboundPlayerPosition: WrapPacketHandlers(&protocol.PlayerPosition{},
				longjump.HandlePlayerPosition, proxy.HandlePlayerPosition, nofall.HandlePlayerPosition,
			),
			protocol.ServerboundPlayerLook: WrapPacketHandlers(&protocol.PlayerLook{},
				proxy.HandlePlayerLook, nofall.HandlePlayerLook,
			),
			protocol.ServerboundPlayerPositionAndLook: WrapPacketHandlers(&protocol.ServerPlayerPositionAndLook{},
				longjump.HandleServerPlayerPositionAndLook, proxy.HandleServerPlayerPositionAndLook, nofall.HandleServerPlayerPositionAndLook,
			),
			protocol.ServerboundPlayerDigging: WrapPacketHandlers(&protocol.PlayerDigging{},
				fastbreak.HandlePlayerDigging,
			),
			protocol.ServerboundCloseWindow: WrapPacketHandlers(&protocol.ServerCloseWindow{},
				proxy.HandleCloseWindow,
			),
			protocol.ServerboundPlayerAbilities: WrapPacketHandlers(&protocol.ServerPlayerAbilities{},
				proxy.HandleServerPlayerAbilities,
			),
			protocol.ServerboundPluginMessage: WrapPacketHandlers(&protocol.PluginMessage{},
				HandlePluginMessage("Texteria", proxy.HandleKeyboardPacketCandidate),
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
				proxy.HandleJoinGame, unlimitedcps.HandleJoinGame,
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
			protocol.ClientboundEntityEffect: WrapPacketHandlers(&protocol.EntityEffect{},
				nobadeffects.HandleEntityEffect,
			),
			// protocol.ClientboundOpenWindow: WrapPacketHandlers(&protocol.OpenWindow{},
			// 	proxy.HandleOpenWindow, cheststealer.HandleOpenWindow,
			// ),
			// protocol.ClientboundCloseWindow: WrapPacketHandlers(&protocol.CloseWindow{},
			// 	proxy.HandleCloseWindow,
			// ),
			// protocol.ClientboundSetSlot: WrapPacketHandlers(&protocol.SetSlot{},
			// 	proxy.HandleSetSlot, cheststealer.HandleSetSlot,
			// ),
			// protocol.ClientboundWindowItems: WrapPacketHandlers(&protocol.WindowItems{},
			// 	proxy.HandleWindowItems, cheststealer.HandleWindowItems,
			// ),
			protocol.ClientboundPlayerAbilities: WrapPacketHandlers(&protocol.PlayerAbilities{},
				flight.HandlePlayerAbilities,
			),
			protocol.ClientboundDisconnect: WrapPacketHandlers(&protocol.Disconnect{},
				HandleDisconnect,
			),
		},
	}
)

func WrapPacketHandlers(packet protocol.Packet, handlers ...PacketHandler) RawPacketHandler {
	return func(rawPacket *protocol.WrappedPacket, tunnel *proxy.MinecraftTunnel) (result *generic.HandlerResult, err error) {
		err = packet.Read(rawPacket.Packet)
		if err != nil {
			return
		}

		for _, handler := range handlers {
			result, err = handler(packet, tunnel)
			if err != nil {
				return
			}
		}

		return
	}
}

func HandleHandshake(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	handshake := packet.(*protocol.Handshake)
	metrics.HandshakeCount.With(prometheus.Labels{"state": fmt.Sprintf("%d", handshake.NextState)}).Inc()

	switch handshake.NextState {
	case 1:
		tunnel.SetState(protocol.ConnStateStatus)
	case 2:
		tunnel.SetState(protocol.ConnStateLogin)
	}

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

	return generic.ModifyPacket(handshake.Marshal()), nil
}

func HandleLoginStart(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	loginStart := packet.(*protocol.LoginStart)
	minecraftTunnel := tunnel.(*proxy.MinecraftTunnel)
	log.Println(loginStart.Name, "is connecting from", minecraftTunnel.Client.Socket.RemoteAddr())
	minecraftTunnel.PlayerHandler.PlayerName = string(loginStart.Name)

	host, _, err := net.SplitHostPort(minecraftTunnel.Client.Socket.RemoteAddr().String())
	if err != nil {
		return
	}

	id := proxy.TunnelPairID{
		Username:   string(loginStart.Name),
		RemoteAddr: host,
	}

	tunnelPair, ok := proxy.CurrentTunnelPool.GetPair(id)
	if !ok {
		minecraftTunnel.Disconnect(chat.Text("unknown session"))
		return generic.RejectPacket(), nil
	}

	if tunnelPair.License == nil || !tunnelPair.License.IsRelated(tunnel) {
		if tunnelPair.License != nil {
			log.Println("unrelated license: got license data", tunnelPair.License, "for connection from", tunnel.GetRemoteAddr())
		}

		minecraftTunnel.Disconnect(chat.Text("license validation failure"))
		return generic.RejectPacket(), nil
	}

	tunnelPair.Primary = minecraftTunnel
	minecraftTunnel.TunnelPair = tunnelPair
	minecraftTunnel.PairID = id

	proxy.RegisterDefaultModules(minecraftTunnel)
	proxy.UpdateConnectionMetrics()

	log.Println("linked minecraft connection for", loginStart.Name, "with auxiliary connection from", tunnelPair.Auxiliary.Conn.RemoteAddr())
	return generic.PassPacket(), nil
}

func HandleEncryptionRequest(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	minecraftTunnel := tunnel.(*proxy.MinecraftTunnel)
	encryptionRequest := packet.(*protocol.EncryptionRequest)

	err = tunnel.WriteClient(encryptionRequest.Marshal())
	if err != nil {
		return
	}

	if minecraftTunnel.TunnelPair == nil || minecraftTunnel.TunnelPair.Auxiliary == nil {
		return
	}

	err = minecraftTunnel.TunnelPair.Auxiliary.SendMessage(proxy.EncryptionDataRequest, proxy.AuxiliaryEncryptionRequest{
		PublicKey: encryptionRequest.PublicKey,
		ServerID:  string(encryptionRequest.ServerID),
	})
	if err != nil {
		return
	}

	key := <-minecraftTunnel.EnableEncryptionS2C
	s2ce, s2cd := newSymmetricEncryption(key)
	minecraftTunnel.Server.SetCipher(s2ce, s2cd)
	return
}

func HandleEncryptionResponse(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	minecraftTunnel := tunnel.(*proxy.MinecraftTunnel)
	encryptionResponse := packet.(*protocol.EncryptionResponse)
	sharedSecret := <-minecraftTunnel.EnableEncryptionC2S

	err = tunnel.WriteServer(encryptionResponse.Marshal())
	if err != nil {
		return
	}

	c2se, c2sd := newSymmetricEncryption(sharedSecret)
	minecraftTunnel.Client.SetCipher(c2se, c2sd)
	minecraftTunnel.EnableEncryptionS2C <- sharedSecret

	err = tunnel.WriteClient((&protocol.SetCompression{Threshold: CompressionThreshold}).Marshal())
	if err != nil {
		return
	}

	minecraftTunnel.Client.SetThreshold(CompressionThreshold)
	return
}

func HandleSetCompression(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	minecraftTunnel := tunnel.(*proxy.MinecraftTunnel)
	setCompression := packet.(*protocol.SetCompression)
	minecraftTunnel.Server.SetThreshold(int(setCompression.Threshold))
	return generic.RejectPacket(), nil
}

func HandleLoginSuccess(_ protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	tunnel.SetState(protocol.ConnStatePlay)
	return generic.PassPacket(), nil
}

func HandleDisconnect(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	disconnect := packet.(*protocol.Disconnect)
	log.Println(tunnel.GetPlayerHandler().GetPlayerName(), "was disconnected from server:", disconnect.Reason.String())
	err = tunnel.WriteClient(packet.Marshal())
	if err != nil {
		return
	}

	tunnel.Close()
	return generic.RejectPacket(), nil
}

func HandlePluginMessage(targetChannel string, handler PluginMessageHandler) PacketHandler {
	return func(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
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
			var next bool
			data, next, err = handler(data, tunnel)
			messageData := pk.PluginMessageData(data)
			if next {
				return generic.ModifyPacket(pk.Marshal(packetID, pk.String(channel), &messageData)), nil
			} else {
				return generic.RejectPacket(), nil
			}
		}

		return generic.PassPacket(), nil
	}
}
