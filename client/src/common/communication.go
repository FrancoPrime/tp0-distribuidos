package common

import (
	"encoding/binary"
	"fmt"
	"net"
)

func sendBet(conn net.Conn, bet Bet) error {
	betMessage := fmt.Sprintf("%s;%s;%s;%s;%s;%s;", bet.AgencyID, bet.Nombre, bet.Apellido, bet.Documento, bet.Nacimiento, bet.Numero)
	return sendMessage(conn, betMessage)
}

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
