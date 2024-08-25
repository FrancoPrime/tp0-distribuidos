package common

import (
	"encoding/csv"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/op/go-logging"
)

const ExitMessage = "exit"
const ErrorMessage = "error"
const CheckWinnersMessage = "winners"

var MaxBatch = 0

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchSize     int
}

// Client Entity that encapsulates how
type Client struct {
	config     ClientConfig
	conn       net.Conn
	running    bool
	bets       []Bet
	currentBet int
	abort      chan struct{}
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:     config,
		running:    true,
		bets:       make([]Bet, 0),
		currentBet: 0,
		abort:      make(chan struct{}),
	}
	return client
}

func isErrorMessage(msg string) bool {
	return strings.EqualFold(msg, ErrorMessage)
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func (c *Client) StopClient() {
	log.Infof("action: stop_client | result: in_progress | client_id: %v",
		c.config.ID,
	)
	c.running = false
	close(c.abort)
}

func (c *Client) processNextBatch() []Bet {
	if c.currentBet >= len(c.bets) {
		return nil
	}
	start := c.currentBet
	end := c.currentBet + c.config.BatchSize
	if end > len(c.bets) {
		end = len(c.bets)
	}
	batch := c.bets[start:end]
	c.currentBet = end
	return batch
}

func (c *Client) LoadBetsFile() error {
	log.Infof("action: load_file | result: in_progress | client_id: %v",
		c.config.ID,
	)
	file, err := os.Open("./agency.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		bet := Bet{
			AgencyID:   c.config.ID,
			Nombre:     record[0],
			Apellido:   record[1],
			Documento:  record[2],
			Nacimiento: record[3],
			Numero:     record[4],
		}
		c.bets = append(c.bets, bet)
	}
	log.Infof("action: load_file | result: success | client_id: %v",
		c.config.ID,
	)
	return nil
}

func (c *Client) SendExitMessage() {
	log.Infof("action: send_exit_message | result: in_progress | client_id: %v",
		c.config.ID,
	)
	err := sendMessage(c.conn, ExitMessage+c.config.ID)
	if err != nil {
		log.Errorf("action: send_exit_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}
	log.Infof("action: send_exit_message | result: success | client_id: %v",
		c.config.ID,
	)
}

func (c *Client) StartClient() {
	c.SendAgencyBets()
	c.CheckWinners()
}

// StartClient Send messages to the client until some time threshold is met
func (c *Client) SendAgencyBets() {
	err := c.LoadBetsFile()
	if err != nil {
		log.Criticalf("action: load_file | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}
	c.createClientSocket()
	defer c.conn.Close()
	log.Infof("action: send_agency_bets | result: in_progress | client_id: %v",
		c.config.ID,
	)
	for batch := c.processNextBatch(); batch != nil; batch = c.processNextBatch() {
		if !c.running {
			log.Infof("action: stop_client | result: success | client_id: %v",
				c.config.ID,
			)
			return
		}

		err := sendBets(c.conn, batch)
		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}
		msg, err := receiveMessage(c.conn)

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Debugf("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		if isErrorMessage(msg) {
			log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: Server Abort",
				c.config.ID,
			)
			return
		} else if wasBetSuccessful(msg) {
			log.Infof("action: apuesta_enviada | result: success | cantidad: %v",
				len(batch),
			)
		} else {
			log.Infof("action: apuesta_enviada | result: fail | cantidad: %v",
				len(batch),
			)
		}
	}

	c.SendExitMessage()

	log.Infof("action: send_agency_bets | result: success | client_id: %v",
		c.config.ID,
	)
}

func (c *Client) CheckWinners() {
	log.Infof("action: check_winners | result: in_progress | client_id: %v",
		c.config.ID,
	)
	for {
		c.createClientSocket()

		err := sendMessage(c.conn, CheckWinnersMessage+c.config.ID)

		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		msg, err := receiveMessage(c.conn)
		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		if isErrorMessage(msg) {
			select {
			case <-time.After(5 * c.config.LoopPeriod):
			case <-c.abort:
				log.Infof("action: stop_client | result: success | client_id: %v",
					c.config.ID,
				)
				return
			}
			continue
		}

		DNIs, err := ParseArrayFromJSON([]byte(msg))

		if err != nil {
			log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(DNIs))
		break
	}
}
