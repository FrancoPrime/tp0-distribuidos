package common

import (
	"fmt"
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

func GetBetFromEnv(agencyID string) Bet {
	return Bet{
		AgencyID:   agencyID,
		Nombre:     os.Getenv("NOMBRE"),
		Apellido:   os.Getenv("APELLIDO"),
		Documento:  os.Getenv("DOCUMENTO"),
		Nacimiento: os.Getenv("NACIMIENTO"),
		Numero:     os.Getenv("NUMERO"),
	}
}

// Serialize Serializes the bet to a string
func (b Bet) Serialize() string {
	return fmt.Sprintf("%s;%s;%s;%s;%s;%s;", b.AgencyID, b.Nombre, b.Apellido, b.Documento, b.Nacimiento, b.Numero)
}

// wasBetSuccessful Checks if the response from the server was a successful message
func wasBetSuccessful(response string) bool {
	return strings.EqualFold(response, SucessfulBetResponse)
}
