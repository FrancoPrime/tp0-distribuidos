package common

import (
	"net"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config  ClientConfig
	conn    net.Conn
	running bool
	abort   chan struct{}
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:  config,
		running: true,
		abort:   make(chan struct{}),
	}
	return client
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

// CheckBetResult Checks the result of the bet sent
func (c *Client) CheckBetResult(bet Bet) {
	msg, err := receiveMessage(c.conn)
	c.conn.Close()

	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	if wasBetSuccessful(msg) {
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			bet.Documento,
			bet.Numero,
		)
	} else {
		log.Infof("action: apuesta_enviada | result: fail | dni: %v | numero: %v",
			bet.Documento,
			bet.Numero,
		)
	}

	log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
		c.config.ID,
		msg,
	)
}

// StartClient Starts the client. It sends a bet to the server
func (c *Client) StartClient() {
	if !c.running {
		log.Infof("action: stop_client | result: success | client_id: %v",
			c.config.ID,
		)
		return
	}
	c.createClientSocket()

	bet := GetBetFromEnv(c.config.ID)
	err := sendBet(c.conn, bet)
	if err != nil {
		log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}
	c.CheckBetResult(bet)
}
