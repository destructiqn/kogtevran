package proxy

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

type AuxiliaryOperationCode int

// Clientbound Operations
const (
	KeepAliveRequest AuxiliaryOperationCode = iota
	EncryptionDataRequest
)

// Serverbound operations
const (
	KeepAliveResponse AuxiliaryOperationCode = iota
	Handshake
	EncryptionDataResponse
)

type AuxiliaryChannel struct {
	Conn       *websocket.Conn
	TunnelPair *TunnelPair
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
			log.Println(err)
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
	case Handshake:
		var handshake AuxiliaryHandshake
		err := mapstructure.Decode(message.Payload, &handshake)
		if err != nil {
			return err
		}

		pair := &TunnelPair{
			Auxiliary: c,
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
	}

	return nil
}

type WebsocketMessage struct {
	OperationCode AuxiliaryOperationCode `json:"op"`
	Payload       interface{}            `json:"payload"`
}

type AuxiliaryHandshake struct {
	Username string `json:"username"`
}

type AuxiliaryEncryptionData struct {
	SharedSecret [][]byte `mapstructure:"candidates"`
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

	channel := AuxiliaryChannel{Conn: conn}
	channel.Handle()

	log.Println("closing auxiliary connection from", r.RemoteAddr)
	conn.Close()
}
