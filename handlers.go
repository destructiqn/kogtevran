package main

import (
	"encoding/hex"
	"fmt"
	"github.com/Tnze/go-mc/chat"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"log"
	"time"
)

type PacketHandler func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error)
type PacketHandlerPool map[int32]PacketHandler

var (
	HandlersS2C = PacketHandlerPool{
		0x01: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			if encryptionResponse {
				return packet.Packet, true, nil
			}

			log.Println("accepting encryption request")
			err = conn.WriteClient(packet.Packet)
			if err != nil {
				next = true
				return
			}

			key := <-enableEncryption
			s2ce, s2cd := newSymmetricEncryption(key)
			conn.Server.SetCipher(s2ce, s2cd) // for server -> client
			log.Println("enabled s2c encryption")
			return
		},
		0x03: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			if compression {
				return packet.Packet, true, nil
			}

			log.Println("handling compression")
			var threshold pk.VarInt
			err = packet.Scan(&threshold)
			if err != nil {
				return packet.Packet, false, err
			}

			compression = true
			log.Println("using s2c compression threshold", threshold)
			conn.Server.SetThreshold(int(threshold))
			enableCompression <- int(threshold)
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
					err = conn.WriteClient(pk.Marshal(0x39, pk.Byte(0x04), pk.Float(0.1), pk.Float(0.1)))
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
			if handshake {
				return packet.Packet, true, nil
			}
			log.Println("handling handshake")
			handshake = true
			return pk.Marshal(0x00, pk.VarInt(47), pk.String(fmt.Sprintf("%s ", RemoteHost)), pk.UnsignedShort(RemotePort), pk.VarInt(2)), true, nil
		},
		0x01: func(packet *Packet, conn *WrappedConn) (result pk.Packet, next bool, err error) {
			if !encryptionResponse {
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
				enableEncryption <- key
				encryptionResponse = true

				threshold := <-enableCompression
				log.Println("using c2s compression threshold", threshold)
				conn.Client.SetThreshold(threshold)
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
			return pk.Marshal(0x03, pk.Boolean(false)), true, nil
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
