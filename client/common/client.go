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

const AGENCY_FILE = "./agency.csv"
const DELIMITER = ","
const END_OF_BET = "\n"

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
	file *os.File
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

// Converts a bet to a byte array
func (c * Client) serialize_bet(bet []string) []byte {
	bet = append([]string{string(c.config.ID)}, bet...)
	data := strings.Join(bet, DELIMITER)
	data += END_OF_BET
	
	log.Debugf("action: serialize_bet | client_id: %v | data: %v", c.config.ID, data)

	return []byte(data)
}

// sendBets Sends a batch of bets to the server. The batch size is not larger than 8kb
func (c * Client) sendBets(reader *csv.Reader, messageHandler *MessageHandler) (bool, error) {
	eofReached := false
	batchMsg := []byte{}

	for i := 0; i < c.config.BatchAmount; i++ {
		line, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				eofReached = true
				break
			}
			return eofReached, err
		}

		batchMsg = append(batchMsg, c.serialize_bet(line)...)
	}

	err := messageHandler.sendMessage(batchMsg, eofReached)

	if err != nil {
		return eofReached, err
	}

	return eofReached, nil
}

// getBetsReader Returns a csv.Reader to read the bets file
func (c *Client) getBetsReader() *csv.Reader {
	file, err := os.OpenFile(AGENCY_FILE, os.O_CREATE|os.O_RDONLY, 0777)
    if err != nil {
        log.Errorf("Error opening file: %v", err)
		return nil
    }

	c.file = file
	reader := csv.NewReader(file)
	return reader
}

// catchSignal Catches the signal SIGTERM to exit gracefully
func (c *Client) catchSignal() {
	sigChannel := make(chan os.Signal, 1)
   	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			s := <-sigChannel
			c.signalHandler(s)
		} 
	}()
}

// StartClientLoop Send messages batches of bets to the server
func (c *Client) StartClientLoop() {
	c.catchSignal()

	if err := c.createClientSocket(); err != nil {
		return
	}

	reader := c.getBetsReader()
	if reader == nil {
		c.conn.Close()
		return
	}
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
			c.file.Close()
			return
		}
	}

	if _, err := messageHandler.receiveMessage() ; err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	} else {
		log.Infof("action: apuestas_enviadas | result: success | client_id: %v", c.config.ID)
	}

	c.conn.Close()
	c.file.Close()
}

func (c *Client) signalHandler(signal os.Signal) {
	c.running = false

	if c.conn != nil {
		c.conn.Close()
		log.Debugf("action: close_connections | result: success | client_id: %v", c.config.ID)
	}

	if c.file != nil {
		c.file.Close()
		log.Debugf("action: close_file | result: sucess | client_id: %v", c.config.ID)
	}
	log.Debugf("action: exit | result: success | client_id: %v", c.config.ID)
}
