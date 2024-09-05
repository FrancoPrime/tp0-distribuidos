package common

import (
	"encoding/csv"
	"io"
	"os"
)

var LastBetRead *Bet
var File *os.File
var FileReader *csv.Reader

func InitializeFileReader(filename string) error {
	log.Infof("action: open_file | result: in_progress")
	file, err := os.Open("./" + filename)
	if err != nil {
		log.Errorf("action: open_file | result: fail")
		return err
	}
	log.Infof("action: open_file | result: success")
	File = file
	FileReader = csv.NewReader(file)
	return nil
}

func CloseFileReader() {
	if File == nil {
		return
	}
	log.Infof("action: close_file | result: in_progress")
	err := File.Close()
	if err != nil {
		log.Errorf("action: close_file | result: fail | error: %v", err)
		return
	}
	log.Info("action: close_file | result: success")
	File = nil
	FileReader = nil
}

// ProcessNextBatch Returns the next batch of bets to be sent to the server
func processNextBatch(agencyID string, maxAmount int) (string, int) {
	if FileReader == nil {
		return "", 0
	}
	batch := ""
	size := 0
	if LastBetRead != nil {
		batch += LastBetRead.Serialize()
		size++
		LastBetRead = nil
	}
	for size < maxAmount {
		record, err := FileReader.Read()
		if err == io.EOF {
			CloseFileReader()
			break
		}
		if err != nil {
			continue
		}

		bet := Bet{
			AgencyID:   agencyID,
			Nombre:     record[0],
			Apellido:   record[1],
			Documento:  record[2],
			Nacimiento: record[3],
			Numero:     record[4],
		}
		serializedBet := bet.Serialize()
		if len(batch)+len(serializedBet) > MaxPayloadSize {
			LastBetRead = &bet
			break
		}
		batch += serializedBet
		size++
	}
	return batch, size
}
