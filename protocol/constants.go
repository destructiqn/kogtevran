package protocol

type ChatPosition byte

const (
	ChatPositionChat ChatPosition = iota
	ChatPositionSystemMessage
	ChatPositionAboveHotbar
)

type ConnectionState int

const (
	ConnStateHandshake ConnectionState = iota
	ConnStateStatus
	ConnStateLogin
	ConnStatePlay
)
