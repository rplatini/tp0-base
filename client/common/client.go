package common

import (
	"encoding/csv"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

const BATCH_SIZE = 8000

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchAmount   int
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

func (c * Client) sendBets(reader *csv.Reader, messageHandler *MessageHandler) (bool, error) {
	batchMsg := []byte{}
	eofReached := false

	for i := 0; i < c.config.BatchAmount; i++ {
		line, err := reader.Read()
	
		if err != nil {
			if err == io.EOF {
				// log.Infof("Reached EOF")
				eofReached = true
				break
			}
			return eofReached, err
		}
	
		bet := NewBet(
			c.config.ID,
			line[0],
			line[1],
			line[2],
			line[3],
			line[4],
		)

		batchMsg = append(batchMsg, bet.serialize()...)
	}

	// log.Infof("msg: %s", batchMsg)

	err := messageHandler.sendMessage([]byte(batchMsg), eofReached)

	if err != nil {
		return eofReached, err
	}

	return eofReached, nil
}


// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(reader *csv.Reader) {
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

	eofReached := false
	for !eofReached {
		var err error
		eofReached, err = c.sendBets(reader, messageHandler)

		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			c.conn.Close()
			return
		}
	}
	

	msg, err := messageHandler.receiveMessage()
	c.conn.Close()

	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	log.Infof("action: apuesta_enviada | result: success | client_id: %v | confirmation: %s", c.config.ID, msg)
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
