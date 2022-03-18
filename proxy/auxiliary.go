package proxy

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/destructiqn/kogtevran/license"
	"github.com/destructiqn/kogtevran/modules"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

type AuxiliaryOperationCode int

// Clientbound Operations
const (
	KeepAliveRequest AuxiliaryOperationCode = iota
	EncryptionDataRequest
	ModuleToggle
)

// Serverbound operations
const (
	KeepAliveResponse AuxiliaryOperationCode = iota
	Handshake
	EncryptionDataResponse
	ModuleToggleAck
)

const KeepAliveInterval = 20 * time.Second

type AuxiliaryChannel struct {
	Conn          *websocket.Conn
	TunnelPair    *TunnelPair
	lastKeepAlive *time.Time
	close         chan bool
}

func (c *AuxiliaryChannel) Close() error {
	go func() {
		c.close <- true
	}()

	return c.Conn.Close()
}

func (c *AuxiliaryChannel) Handle() {
	for {
		var message WebsocketMessage
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			return
		}

		err = c.HandleMessage(&message)
		if err != nil {
			log.Println("error in auxiliary connection:", err)
			_ = c.Close()
			return
		}
	}
}

func (c *AuxiliaryChannel) HandleKeepAlive() {
	ticker := time.NewTicker(KeepAliveInterval)
	for {
		select {
		case <-ticker.C:
			if c.lastKeepAlive != nil && time.Now().Sub(*c.lastKeepAlive) > KeepAliveInterval*2 {
				log.Println("dropping connection from", c.Conn.RemoteAddr(), "due to keep alive timeout")
				if c.TunnelPair.Primary != nil {
					c.TunnelPair.Primary.Close()
				}
				return
			}

			err := c.SendMessage(KeepAliveRequest, nil)
			if err != nil {
				log.Println("unable to write keep alive request:", err)
				if c.TunnelPair.Primary != nil {
					c.TunnelPair.Primary.Close()
				}
				return
			}
		case <-c.close:
			return
		}
	}
}

func (c *AuxiliaryChannel) SendMessage(operation AuxiliaryOperationCode, payload interface{}) error {
	return c.Conn.WriteJSON(WebsocketMessage{
		OperationCode: operation,
		Payload:       payload,
	})
}

func (c *AuxiliaryChannel) HandleMessage(message *WebsocketMessage) error {
	if c.TunnelPair == nil && message.OperationCode != Handshake {
		return fmt.Errorf("expected handshake for request with unknown source, but got %d", message.OperationCode)
	}

	switch message.OperationCode {
	case KeepAliveResponse:
		now := time.Now()
		c.lastKeepAlive = &now
	case Handshake:
		var handshake AuxiliaryHandshake
		err := mapstructure.Decode(message.Payload, &handshake)
		if err != nil {
			return err
		}

		licenseData, err := license.GetLicense(handshake.AuthKey)
		if err != nil {
			return err
		}

		pair := &TunnelPair{
			Auxiliary: c,
			License:   licenseData,
		}

		host, _, err := net.SplitHostPort(c.Conn.RemoteAddr().String())
		if err != nil {
			return err
		}

		id := TunnelPairID{
			Username:   handshake.Username,
			RemoteAddr: host,
		}

		c.TunnelPair = pair
		CurrentTunnelPool.RegisterPair(id, pair)
	case EncryptionDataResponse:
		var encryptionData AuxiliaryEncryptionData
		err := mapstructure.Decode(message.Payload, &encryptionData)
		if err != nil {
			return err
		}

		if c.TunnelPair == nil {
			return errors.New("cannot find corresponding minecraft tunnel")
		}

		c.TunnelPair.Primary.EnableEncryptionC2S <- encryptionData.SharedSecret
	case ModuleToggleAck:
		var moduleData AuxiliaryToggleModuleAck
		err := mapstructure.Decode(message.Payload, &moduleData)
		if err != nil {
			return err
		}

		if c.TunnelPair == nil {
			return errors.New("cannot find corresponding minecraft tunnel")
		}

		moduleHandler := c.TunnelPair.Primary.ModuleHandler
		module, ok := moduleHandler.GetModule(moduleData.Identifier)
		if !ok {
			return errors.New("unknown module")
		}

		clientModule, ok := module.(*modules.ClientModule)
		if !ok {
			return errors.New("this module is not handled by client")
		}

		clientModule.SetEnabled(moduleData.Status)
		return moduleHandler.UpdateModule(clientModule)
	}

	return nil
}

type WebsocketMessage struct {
	OperationCode AuxiliaryOperationCode `json:"op"`
	Payload       interface{}            `json:"payload"`
}

type AuxiliaryHandshake struct {
	Username string `json:"username"`
	AuthKey  string `json:"authKey"`
}

type AuxiliaryEncryptionData struct {
	SharedSecret [][]byte `mapstructure:"candidates"`
}

type AuxiliaryToggleModule struct {
	Identifier string `json:"identifier"`
}

type AuxiliaryToggleModuleAck struct {
	Identifier string `json:"identifier"`
	Status     bool   `json:"status"`
}

var WebsocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := WebsocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("accepted auxiliary connection from", r.RemoteAddr)

	channel := AuxiliaryChannel{Conn: conn, close: make(chan bool)}
	go channel.HandleKeepAlive()
	channel.Handle()

	log.Println("closing auxiliary connection from", r.RemoteAddr)
	_ = channel.Close()
}
