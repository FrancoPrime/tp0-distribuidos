package common

import (
	"encoding/json"
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

func wasBetSuccessful(response string) bool {
	return strings.EqualFold(response, SucessfulBetResponse)
}

func ParseArrayFromJSON(bytes []byte) ([]string, error) {
	result := []string{}
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
