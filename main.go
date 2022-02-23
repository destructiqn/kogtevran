package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/destructiqn/kogtevran/net"
	"github.com/destructiqn/kogtevran/net/CFB8"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
	"github.com/destructiqn/kogtevran/proxy"
)

var ServerPort = 25565

func GetRemoteAddr() string {
	return fmt.Sprintf("%s:%d", GetConnectAddress(), ServerPort)
}

func main() {
	go func() {
		_ = http.ListenAndServe("127.0.0.1:8080", http.HandlerFunc(proxy.WebsocketHandler))
	}()

	proxyServer, err := net.ListenMC("127.0.0.1:25565")
	if err != nil {
		panic(err)
	}

	log.Println("server is now listening for connections")
	for {
		client, err := proxyServer.Accept()
		if err != nil {
			panic(err)
		}

		server, err := net.DialMC(GetRemoteAddr())
		if err != nil {
			panic(err)
		}

		conn := proxy.WrapConn(server, &client)
		proxy.RegisterDefaultModules(conn)

		go pipe(conn, protocol.ConnS2C)
		go pipe(conn, protocol.ConnC2S)
	}
}

func pipe(conn *proxy.MinecraftTunnel, typ int) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()

	srcName, dstName := "client", "server"
	src, _ := conn.Client, conn.Server
	if typ == protocol.ConnS2C {
		srcName, dstName = dstName, srcName
		src, _ = conn.Server, conn.Client
	}

	direction := fmt.Sprintf("%s -> %s", srcName, dstName)
	var packets int
	var err error
	var lastPacket *pk.Packet
	for {
		var packet pk.Packet
		err = src.ReadPacket(&packet)
		if err != nil {
			if conn.Closed {
				log.Println("closing", direction)
				break
			}

			if err == io.EOF {
				return
			}

			log.Println(direction, "error reading packet", err)
			break
		}

		wrappedPacket := protocol.WrapPacket(packet, typ)
		stateHandlerPool := ClientboundHandlers
		if typ == protocol.ConnC2S {
			stateHandlerPool = ServerboundHandlers
		}

		next := true
		if stateHandler, ok := stateHandlerPool[conn.State]; ok {
			if handler, ok := stateHandler[packet.ID]; ok {
				result, err := handler(wrappedPacket, conn)
				if err != nil {
					log.Println(direction, "error handling packet", protocol.FormatPacket(packet.ID, typ), err)
					next = true
					continue
				}

				if result == nil {
					continue
				}

				if result.IsModified {
					packet = result.Packet
				}

				if next {
					next = result.ShouldPass
				}
			}
		}

		if next {
			write := conn.WriteClient
			if typ == protocol.ConnC2S {
				write = conn.WriteServer
			}

			err = write(packet)
			if err != nil {
				log.Println(direction, "error writing packet", wrappedPacket.Name, "to", dstName)
				break
			}

			packets++
		}

		if packet.ID != protocol.ClientboundDisconnect {
			lastPacket = &packet
		}
	}

	if err != nil {
		fmt.Println("error in", srcName, "->", dstName, "connection:", err)
		conn.Close()
	}

	if lastPacket != nil {
		fmt.Println("total packets handled:", packets)
		fmt.Println("last packet was", protocol.FormatPacket(lastPacket.ID, typ), "")
		fmt.Println(fmt.Sprintf("last packet dump:\n%s", hex.Dump(lastPacket.Data)))
	}
}

func GetConnectAddress() string {
	addresses := []string{
		"54.36.120.15",
		"46.105.114.88",
		"46.105.113.87",
		"5.39.71.168",
		"5.39.71.186",
		"46.105.114.126",
		"51.178.178.68",
		"54.37.81.60",
		"46.105.114.103",
		"46.105.114.104",
		"178.33.226.100",
		"54.37.83.112",
		"178.33.226.137",
		"5.39.71.183",
		"46.105.114.5",
	}

	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(addresses))
	return addresses[i]
}

func newSymmetricEncryption(key []byte) (eStream, dStream cipher.Stream) {
	b, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	dStream = CFB8.NewCFB8Decrypt(b, key)
	eStream = CFB8.NewCFB8Encrypt(b, key)
	return
}
