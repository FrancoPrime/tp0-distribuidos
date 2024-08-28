package common

import (
	"net"
	"strings"
	"time"

	"github.com/op/go-logging"
)

const ExitMessage = "exit"
const ErrorMessage = "error"

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

// isErrorMessage Checks if the message received from the server is an error message
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

// StopClient Puts the client as non running. Aborts current loop if exists.
func (c *Client) StopClient() {
	log.Infof("action: stop_client | result: in_progress | client_id: %v",
		c.config.ID,
	)
	c.running = false
	close(c.abort)
}

// ProcessNextBatch Returns the next batch of bets to be sent to the server
func (c *Client) processNextBatch() (string, int) {
	if c.currentBet >= len(c.bets) {
		return "", 0
	}
	start := c.currentBet
	end := c.currentBet + c.config.BatchSize
	size := 0
	batch := ""
	for i := start; i < end && i < len(c.bets); i++ {
		bet := c.bets[i].Serialize()
		if len(batch)+len(bet) > MaxPayloadSize {
			break
		}
		batch += bet
		size++
	}
	c.currentBet += size
	return batch, size
}

// LoadBetsFile Loads the bets file from the filesystem
func (c *Client) LoadBetsFile() error {
	log.Infof("action: load_file | result: in_progress | client_id: %v",
		c.config.ID,
	)
	bets, err := getBetsFromFile(c.config.ID)
	if err != nil {
		log.Errorf("action: load_file | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.bets = bets
	log.Infof("action: load_file | result: success | client_id: %v",
		c.config.ID,
	)
	return nil
}

// SendExitMessage Informs the server all bets were placed
func (c *Client) SendExitMessage() {
	log.Infof("action: send_exit_message | result: in_progress | client_id: %v",
		c.config.ID,
	)
	err := sendMessage(c.conn, ExitMessage)
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

// CheckMessageResult Checks if the message received from the server is an error message
func (c *Client) CheckMessageResult(batchSize int) bool {
	msg, err := receiveMessage(c.conn)

	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return false
	}

	log.Debugf("action: receive_message | result: success | client_id: %v | msg: %v",
		c.config.ID,
		msg,
	)

	if isErrorMessage(msg) {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: Server Abort",
			c.config.ID,
		)
		return false
	} else if wasBetSuccessful(msg) {
		log.Infof("action: apuesta_enviada | result: success | cantidad: %v",
			batchSize,
		)
	} else {
		log.Infof("action: apuesta_enviada | result: fail | cantidad: %v",
			batchSize,
		)
	}
	return true
}

// StartClient Starts the client. It loads the bets file and sends them in batch to the server
func (c *Client) StartClient() {
	err := c.LoadBetsFile()
	if err != nil {
		return
	}
	c.createClientSocket()
	defer c.conn.Close()
	for batch, size := c.processNextBatch(); size > 0; batch, size = c.processNextBatch() {
		if !c.running {
			log.Infof("action: stop_client | result: success | client_id: %v",
				c.config.ID,
			)
			return
		}

		err := sendMessage(c.conn, batch)
		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		shouldContinue := c.CheckMessageResult(size)
		if !shouldContinue {
			return
		}
	}

	c.SendExitMessage()

	log.Infof("action: client_finished | result: success | client_id: %v",
		c.config.ID,
	)
}
