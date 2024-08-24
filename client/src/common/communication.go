package common

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
)

func sendMessage(conn net.Conn, bet Bet) error {
	betJSONBytes, err := json.Marshal(bet)
	if err != nil {
		return fmt.Errorf("failed to jsonify bet: %w", err)
	}

	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, uint16(len(betJSONBytes)))
	betJSONBytes = append(lengthBytes, betJSONBytes...)

	remaining := len(betJSONBytes)
	for remaining > 0 {
		bytesWritten, err := conn.Write(betJSONBytes[len(betJSONBytes)-remaining:])
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
