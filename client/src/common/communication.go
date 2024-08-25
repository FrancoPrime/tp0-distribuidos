package common

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
)

func sendBets(conn net.Conn, batch []Bet) error {
	betBatchJSONBytes, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to jsonify bet: %w", err)
	}
	return sendMessage(conn, string(betBatchJSONBytes))
}

func sendMessage(conn net.Conn, content string) error {
	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, uint16(len(content)))
	bytes := append(lengthBytes, content...)

	remaining := len(bytes)
	for remaining > 0 {
		bytesWritten, err := conn.Write(bytes[len(bytes)-remaining:])
		if err != nil {
			return fmt.Errorf("failed to write to connection: %w", err)
		}

		remaining -= bytesWritten
	}

	return nil
}

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
