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

	"github.com/ruscalworld/vimeinterceptor/net"
	"github.com/ruscalworld/vimeinterceptor/net/CFB8"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

var ServerPort = 25565

func GetRemoteAddr() string {
	return fmt.Sprintf("%s:%d", GetConnectAddress(), ServerPort)
}

func main() {
	go func() {
		_ = http.ListenAndServe(":8080", http.HandlerFunc(WebsocketHandler))
	}()

	proxy, err := net.ListenMC(":25565")
	if err != nil {
		panic(err)
	}

	log.Println("server is now listening for connections")
	for {
		client, err := proxy.Accept()
		if err != nil {
			panic(err)
		}

		server, err := net.DialMC(GetRemoteAddr())
		if err != nil {
			panic(err)
		}

		conn := WrapConn(server, &client)
		RegisterDefaultModules(conn)

		go pipe(conn, ConnS2C)
		go pipe(conn, ConnC2S)
	}
}

func pipe(conn *MinecraftTunnel, typ int) {
	srcName, dstName := "client", "server"
	src, _ := conn.Client, conn.Server
	if typ == ConnS2C {
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
				return
			}

			if err == io.EOF {
				return
			}

			log.Println(direction, "error reading packet", err)
			break
		}

		wrappedPacket := WrapPacket(packet, typ)
		handlerPool := HandlersS2C
		if typ == ConnC2S {
			handlerPool = HandlersC2S
		}

		next := true
		if handler, ok := handlerPool[packet.ID]; ok {
			packet, next, err = handler(wrappedPacket, conn)
			if err != nil {
				log.Println(direction, "error handling packet", FormatPacket(packet.ID, typ), err)
				next = true
				continue
			}
		}

		if next {
			write := conn.WriteClient
			if typ == ConnC2S {
				write = conn.WriteServer
			}

			err = write(packet)
			if err != nil {
				log.Println(direction, "error writing packet", wrappedPacket.Name, "to", dstName)
				break
			}

			packets++
		}

		if packet.ID != 0x40 {
			lastPacket = &packet
		}
	}

	if err != nil {
		fmt.Println("error in", srcName, "->", dstName, "connection:", err)
		conn.Close()
	}

	if lastPacket != nil {
		fmt.Println("total packets handled:", packets)
		fmt.Println("last packet was", FormatPacket(lastPacket.ID, typ), "")
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
