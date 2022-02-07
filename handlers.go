package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/Tnze/go-mc/chat"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

type PacketHandler func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error)
type PacketHandlerPool map[int32]PacketHandler

const (
	CompressionThreshold = 1024
)

var (
	HandlersS2C = PacketHandlerPool{
		0x01: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			switch conn.State {
			case ConnStateLogin:
				log.Println("accepting encryption request")
				err = conn.WriteClient(packet.Packet)
				if err != nil {
					next = true
					return
				}

				key := <-conn.EnableEncryption
				s2ce, s2cd := newSymmetricEncryption(key)
				conn.Server.SetCipher(s2ce, s2cd) // for server -> client
				return
			case ConnStatePlay:
				var (
					EntityID         pk.Int
					GameMode         pk.UnsignedByte
					Dimension        pk.Byte
					Difficulty       pk.UnsignedByte
					MaxPlayers       pk.UnsignedByte
					LevelType        pk.String
					ReducedDebugInfo pk.Boolean
				)

				err = packet.Scan(&EntityID, &GameMode, &Dimension, &Difficulty, &MaxPlayers, &LevelType, &ReducedDebugInfo)
				if err != nil {
					return
				}

				conn.EntityID = int32(EntityID)
				conn.resetEntities()
			}

			return packet.Packet, true, nil
		},
		0x02: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			if conn.State == ConnStateLogin {
				conn.State = ConnStatePlay
			}

			return packet.Packet, true, nil
		},
		0x03: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			switch conn.State {
			case ConnStateLogin:
				if conn.State != ConnStateLogin {
					return packet.Packet, true, nil
				}

				log.Println("handling compression")
				var threshold pk.VarInt
				err = packet.Scan(&threshold)
				if err != nil {
					return
				}

				log.Println("using s2c compression threshold", threshold)
				conn.Server.SetThreshold(int(threshold))
				return
			}

			return packet.Packet, true, nil
		},
		// Player Position And Look
		0x08: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				X, Y, Z    pk.Double
				Yaw, Pitch pk.Float
				Flags      pk.Byte
			)

			err = packet.Scan(&X, &Y, &Z, &Yaw, &Pitch, &Flags)
			if err != nil {
				return
			}

			if Flags&0x01 > 0 {
				conn.Location.X += float64(X)
			} else {
				conn.Location.X = float64(X)
			}

			if Flags&0x02 > 0 {
				conn.Location.Y += float64(Y)
			} else {
				conn.Location.Y = float64(Y)
			}

			if Flags&0x04 > 0 {
				conn.Location.Z += float64(Z)
			} else {
				conn.Location.Z = float64(Z)
			}

			conn.Location.Yaw, conn.Location.Pitch = byte(Yaw), byte(Pitch)
			return packet.Packet, true, nil
		},
		// Spawn Player
		0x0C: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				EntityID    pk.VarInt
				PlayerUUID  pk.UUID
				X, Y, Z     pk.Int
				Yaw, Pitch  pk.Angle
				CurrentItem pk.Short
			)

			err = packet.Scan(&EntityID, &PlayerUUID, &X, &Y, &Z, &Yaw, &Pitch, &CurrentItem)
			if err != nil {
				return
			}

			player := &Player{
				DefaultEntity{Location: &Location{
					X:     float64(X) / 32,
					Y:     float64(Y) / 32,
					Z:     float64(Z) / 32,
					Yaw:   byte(Yaw),
					Pitch: byte(Pitch),
				}},
			}

			conn.initPlayer(int(EntityID), player)
			return packet.Packet, true, nil
		},
		0x0F: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				EntityID                        pk.VarInt
				Type                            pk.UnsignedByte
				X, Y, Z                         pk.Int
				Yaw, Pitch, HeadPitch           pk.Angle
				VelocityX, VelocityY, VelocityZ pk.Short
			)

			err = packet.Scan(&EntityID, &Type, &X, &Y, &Z, &Yaw, &Pitch, &HeadPitch, &VelocityX, &VelocityY, &VelocityZ)
			if err != nil {
				return
			}

			mob := &Mob{
				DefaultEntity: DefaultEntity{Location: &Location{
					X:     float64(X) / 32,
					Y:     float64(Y) / 32,
					Z:     float64(Z) / 32,
					Yaw:   byte(Yaw),
					Pitch: byte(Pitch),
				}},
				Type: MobType(Type),
			}

			conn.initMob(int(EntityID), mob)
			return packet.Packet, true, nil
		},
		// Entity Velocity
		0x12: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				EntityID pk.VarInt
				X, Y, Z  pk.Short
			)

			err = packet.Scan(&EntityID, &X, &Y, &Z)
			if err != nil {
				return
			}

			antiKnockback := conn.Modules[ModuleAntiKnockback].(*AntiKnockback)

			if !conn.IsModuleEnabled(ModuleAntiKnockback) || int32(EntityID) != conn.EntityID {
				return packet.Packet, true, nil
			}

			return pk.Marshal(0x12, EntityID, pk.Short(antiKnockback.X), pk.Short(antiKnockback.Y), pk.Short(antiKnockback.Z)), true, nil
		},
		// Destroy Entities
		0x13: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				Count     pk.VarInt
				EntityIDs []pk.VarInt
			)

			err = packet.Scan(&Count, &pk.Ary{
				Len: &Count,
				Ary: &EntityIDs,
			})
			if err != nil {
				return
			}

			entityIDs := make([]int, 0)
			for _, entityID := range EntityIDs {
				entityIDs = append(entityIDs, int(entityID))
			}

			conn.destroyEntities(entityIDs)
			return packet.Packet, true, nil
		},
		// Entity Relative Move
		0x15: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				EntityID pk.VarInt
				X, Y, Z  pk.Byte
				OnGround pk.Boolean
			)

			err = packet.Scan(&EntityID, &X, &Y, &Z, &OnGround)
			if err != nil {
				return
			}

			conn.entityRelativeMove(int(EntityID), float64(X)/32, float64(Y)/32, float64(Z)/32)
			return packet.Packet, true, nil
		},
		// Entity Look And Relative Move
		0x17: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				EntityID   pk.VarInt
				X, Y, Z    pk.Byte
				Yaw, Pitch pk.Angle
				OnGround   pk.Boolean
			)

			err = packet.Scan(&EntityID, &X, &Y, &Z, &Yaw, &Pitch, &OnGround)
			if err != nil {
				return
			}

			conn.entityRelativeMove(int(EntityID), float64(X)/32, float64(Y)/32, float64(Z)/32)
			return packet.Packet, true, nil
		},
		// Entity Teleport
		0x18: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				EntityID   pk.VarInt
				X, Y, Z    pk.Int
				Yaw, Pitch pk.Angle
				OnGround   pk.Boolean
			)

			err = packet.Scan(&EntityID, &X, &Y, &Z, &Yaw, &Pitch, &OnGround)
			if err != nil {
				return
			}

			conn.entityTeleport(int(EntityID), float64(X)/32, float64(Y)/32, float64(Z)/32, byte(Yaw), byte(Pitch))
			return packet.Packet, true, nil
		},
		0x39: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				Flags               pk.Byte
				FlyingSpeed         pk.Float
				FieldOfViewModifier pk.Float
			)

			err = packet.Scan(&Flags, &FlyingSpeed, &FieldOfViewModifier)
			if err != nil {
				return
			}

			if conn.IsModuleEnabled(ModuleFlight) {
				go func(conn *WrappedConn) {
					time.Sleep(100 * time.Millisecond)
					flight := conn.Modules[ModuleFlight].(*Flight)
					err = flight.Update()
				}(conn)
			}

			return packet.Packet, true, err
		},
		//0x3F: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
		//	err = HandlePluginMessage(packet.Packet, "server")
		//	return packet.Packet, true, err
		//},
		0x40: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var Reason chat.Message
			err = packet.Scan(&Reason)
			if err != nil {
				return
			}

			log.Println("disconnected from server:", Reason.String())
			err = conn.WriteClient(packet.Packet)
			if err != nil {
				return packet.Packet, false, err
			}

			conn.Disconnect()
			return
		},
	}

	HandlersC2S = PacketHandlerPool{
		0x00: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			if conn.State != ConnStateHandshake {
				return packet.Packet, true, nil
			}
			log.Println("handling handshake")
			conn.State = ConnStateLogin
			return pk.Marshal(0x00, pk.VarInt(47), pk.String(fmt.Sprintf("%s ", RemoteHost)), pk.UnsignedShort(RemotePort), pk.VarInt(2)), true, nil
		},
		0x01: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			if conn.State == ConnStateLogin {
				log.Println("accepting encryption response")
				key := <-secretChannel
				log.Println("accepted shared secret:", key)

				err = conn.WriteServer(packet.Packet)
				if err != nil {
					return packet.Packet, true, err
				}

				c2se, c2sd := newSymmetricEncryption(key)
				conn.Client.SetCipher(c2se, c2sd) // for client -> server
				log.Println("enabled c2s encryption")
				conn.EnableEncryption <- key

				err = conn.WriteClient(pk.Marshal(0x03, pk.VarInt(CompressionThreshold)))
				if err != nil {
					return
				}

				log.Println("using c2s compression threshold", CompressionThreshold)
				conn.Client.SetThreshold(CompressionThreshold)
				return
			}

			var Message pk.String
			err = packet.Scan(&Message)
			if err != nil {
				log.Println("error scanning chat message:", err)
				return
			}

			handled := HandleCommand(string(Message), conn)
			return packet.Packet, !handled, nil
		},
		0x03: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var OnGround pk.Boolean

			err = packet.Scan(&OnGround)
			if err != nil {
				return
			}

			return pk.Marshal(0x03, OnGround || pk.Boolean(conn.IsModuleEnabled(ModuleNoFall))), true, nil
		},
		0x04: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				X, Y, Z  pk.Double
				OnGround pk.Boolean
			)

			err = packet.Scan(&X, &Y, &Z, &OnGround)
			if err != nil {
				return
			}

			conn.Location.X, conn.Location.Y, conn.Location.Z = float64(X), float64(Y), float64(Z)

			if !conn.IsModuleEnabled(ModuleNoFall) {
				return packet.Packet, true, nil
			}

			return pk.Marshal(0x04, X, Y, Z, OnGround || pk.Boolean(conn.IsModuleEnabled(ModuleNoFall))), true, nil
		},
		// Player Position And Look
		0x06: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var (
				X, Y, Z    pk.Double
				Yaw, Pitch pk.Float
				OnGround   pk.Boolean
			)

			err = packet.Scan(&X, &Y, &Z, &Yaw, &Pitch, &OnGround)
			if err != nil {
				return
			}

			conn.Location.X, conn.Location.Y, conn.Location.Z = float64(X), float64(Y), float64(Z)
			conn.Location.Yaw, conn.Location.Pitch = byte(Yaw), byte(Pitch)

			if !conn.IsModuleEnabled(ModuleNoFall) {
				return packet.Packet, true, nil
			}

			return pk.Marshal(0x06, X, Y, Z, Yaw, Pitch, pk.Boolean(true)), true, nil
		},
		0x13: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			var Flags pk.Byte
			err = packet.Scan(&Flags)
			if err != nil {
				return
			}

			conn.IsFlying = Flags&0x02 > 0
			return packet.Packet, true, nil
		},
		//0x17: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
		//	err = HandlePluginMessage(packet.Packet, "client")
		//	return packet.Packet, true, err
		//},
	}
)

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
