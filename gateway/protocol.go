package gateway

import "time"

type MessageType string

const (
	ProtocolVersion = "1.0"

	MessageTypeInput  MessageType = "input"
	MessageTypeOutput MessageType = "output"
	MessageTypeResize MessageType = "resize"
	MessageTypeError  MessageType = "error"
)

type TerminalMessage struct {
	Version   string      `json:"version"`
	Type      MessageType `json:"type"`
	Data      string      `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

type ResizeMetadata struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

func NewOutputMessage(data string) TerminalMessage {
	return TerminalMessage{
		Type:      MessageTypeOutput,
		Data:      data,
		Timestamp: time.Now(),
	}
}

func NewErrorMessage(err string) TerminalMessage {
	return TerminalMessage{
		Type:      MessageTypeError,
		Data:      err,
		Timestamp: time.Now(),
	}
}
