package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/destructiqn/kogtevran/metrics"
	"github.com/destructiqn/kogtevran/net"
	"github.com/destructiqn/kogtevran/net/CFB8"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
	"github.com/destructiqn/kogtevran/proxy"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var ServerPort = 25565

func GetRemoteAddr() string {
	return fmt.Sprintf("%s:%d", GetConnectAddress(), ServerPort)
}

func main() {
	metrics.RegisterMetrics()
	go func() {
		err := http.ListenAndServe("0.0.0.0:9090", promhttp.Handler())
		if err != nil {
			log.Fatalln("prometheus listener error:", err)
		}
	}()

	go func() {
		var err error
		if certPath, ok := os.LookupEnv("KV_CERT_PATH"); ok {
			keyPath := os.Getenv("KV_CERT_KEY_PATH")
			err = http.ListenAndServeTLS("0.0.0.0:8080", certPath, keyPath, http.HandlerFunc(proxy.WebsocketHandler))
		} else {
			err = http.ListenAndServe("0.0.0.0:8080", http.HandlerFunc(proxy.WebsocketHandler))
		}

		if err != nil {
			log.Fatalln("http listener error:", err)
		}
	}()

	proxyServer, err := net.ListenMC("0.0.0.0:25565")
	if err != nil {
		log.Fatalln("error starting proxy listener")
	}

	log.Println("server is now listening for connections")
	for {
		client, err := proxyServer.Accept()
		if err != nil {
			log.Println("error accepting connection:", err)
			continue
		}

		server, err := net.DialMC(GetRemoteAddr())
		if err != nil {
			log.Println("error connecting to vimeworld:", err)
			continue
		}

		conn := proxy.WrapConn(server, &client)

		go pipe(conn, protocol.ConnS2C)
		go pipe(conn, protocol.ConnC2S)
	}
}

func pipe(conn *proxy.MinecraftTunnel, typ int) {
	defer func() {
		conn.Close()
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
	for {
		var packet pk.Packet
		err = src.ReadPacket(&packet)
		if err != nil {
			if conn.Closed {
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
	}

	if err != nil {
		log.Println("error in", direction, "connection:", err)
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
