package common

import (
	"encoding/binary"
	"fmt"
	"net"
)

var MaxPayloadSize = 8*1024 - 2

// SendMessage Sends a message to the server with the communication protocol defined
func sendMessage(conn net.Conn, message string) error {
	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, uint16(len(message)))
	messageBytes := append(lengthBytes, message...)

	remaining := len(messageBytes)
	for remaining > 0 {
		bytesWritten, err := conn.Write(messageBytes[len(messageBytes)-remaining:])
		if err != nil {
			return fmt.Errorf("failed to write to connection: %w", err)
		}

		remaining -= bytesWritten
	}

	return nil
}

// ReceiveMessage Receives a message from the server with the communication protocol defined
func receiveMessage(conn net.Conn) (string, error) {
	// Read the 2-byte length field
	lengthBytes := make([]byte, 2)
	remaining := len(lengthBytes)

	for remaining > 0 {
		n, err := conn.Read(lengthBytes[len(lengthBytes)-remaining:])
		if err != nil {
			return "", fmt.Errorf("failed to read length from connection: %w", err)
		}
		remaining -= n
	}

	length := binary.BigEndian.Uint16(lengthBytes)

	messageBytes := make([]byte, length)
	remaining = int(length)

	for remaining > 0 {
		n, err := conn.Read(messageBytes[len(messageBytes)-remaining:])
		if err != nil {
			return "", fmt.Errorf("failed to read message from connection: %w", err)
		}
		remaining -= n
	}

	return string(messageBytes), nil
}
