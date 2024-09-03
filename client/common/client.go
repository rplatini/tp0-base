package common

import (
	"encoding/csv"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

const EMPTY = ""

const NOMBRE = 0
const APELLIDO = 1
const DOCUMENTO = 2
const NACIMIENTO = 3
const NUMERO = 4

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
		return err
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
				eofReached = true
				break
			}
			return eofReached, err
		}
	
		bet := NewBet(
			c.config.ID,
			line[NOMBRE],
			line[APELLIDO],
			line[DOCUMENTO],
			line[NACIMIENTO],
			line[NUMERO],
		)

		batchMsg = append(batchMsg, bet.serialize()...)
	}

	err := messageHandler.sendMessage([]byte(batchMsg), eofReached)
	return eofReached, err
}

func (c *Client) AskForWinners(messageHandler *MessageHandler) (int, error) {
	winnersAsk := "WINNERS" + DELIMITER + c.config.ID

	err := messageHandler.sendMessage([]byte(winnersAsk), true)
	if err == nil {
		winners, err := messageHandler.receiveMessage()
		// log.Debugf("DNI winners: %v", winners)

		if err == nil {
			winnersCount := 0
			if winners != EMPTY {
				winnersCount = len(strings.Split(winners, DELIMITER))
			}
			return winnersCount, err
		}
	}
		
	return -1, err
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

	if err := c.createClientSocket(); err != nil {
		return
	}
	messageHandler := &MessageHandler{conn: c.conn}

	for {
		eofReached, err := c.sendBets(reader, messageHandler)

		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			c.conn.Close()
			return
		}

		if eofReached {
			break
		}
	}

	dniWinners, err := c.AskForWinners(messageHandler)
	c.conn.Close()

	if err != nil {
		log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: ${%d}", dniWinners)
}

func (c *Client) signalHandler(signal os.Signal) {
	c.running = false

	if c.conn != nil {
		log.Debugf("action: close_connections | result: in_progress | client_id: %v", c.config.ID)
		c.conn.Close()
		log.Debugf("action: exit | result: success | client_id: %v", c.config.ID)
	}
}
