package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

const SucessfulBetResponse = "success"

type Bet struct {
	AgencyID   string `json:"id"`
	Nombre     string `json:"nombre"`
	Apellido   string `json:"apellido"`
	Documento  string `json:"documento"`
	Nacimiento string `json:"nacimiento"`
	Numero     string `json:"numero"`
}

// Serialize Serializes the bet to a string
func (b Bet) Serialize() string {
	return fmt.Sprintf("%s;%s;%s;%s;%s;%s;", b.AgencyID, b.Nombre, b.Apellido, b.Documento, b.Nacimiento, b.Numero)
}

// wasBetSuccessful Checks if the response from the server was a successful message
func wasBetSuccessful(response string) bool {
	return strings.EqualFold(response, SucessfulBetResponse)
}

// GetBetsFromFile Reads the file agency.csv and returns a slice of Bet
func getBetsFromFile(agencyID string) ([]Bet, error) {
	file, err := os.Open("./agency.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	bets := make([]Bet, 0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		bet := Bet{
			AgencyID:   agencyID,
			Nombre:     record[0],
			Apellido:   record[1],
			Documento:  record[2],
			Nacimiento: record[3],
			Numero:     record[4],
		}
		bets = append(bets, bet)
	}
	return bets, nil
}
