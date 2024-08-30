package common

import (
	"net"
	"os"
	"os/signal"
	"syscall"
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
	config ClientConfig
	conn  net.Conn
	running bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
		running: true,
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

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(bet *Bet) {
	sigChannel := make(chan os.Signal, 1)
   	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			s := <-sigChannel
			c.signalHandler(s)
		} 
	}()
	
	if !c.running {
		log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
		return
	}

	c.createClientSocket()
	messageHandler := &MessageHandler{conn: c.conn}

	msg := bet.serialize()
	err := messageHandler.sendMessage(msg)

	if err != nil {
		log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		c.conn.Close()
		return
	}

	_, err = messageHandler.receiveMessage()
	c.conn.Close()

	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	log.Infof("action: apuesta_enviada | result: success | dni: ${%v} | numero: ${%v}", bet.documento, bet.numero)
}

func (c *Client) signalHandler(signal os.Signal) {
	log.Infof("action: signal | result: received | signal: %v", signal)
	c.running = false

	if c.conn != nil {
		log.Infof("action: closing client socket | result: in_progress | client_id: %v", c.config.ID)
		c.conn.Close()
		log.Infof("action: closing client socket | result: success | client_id: %v", c.config.ID)
	}
}
