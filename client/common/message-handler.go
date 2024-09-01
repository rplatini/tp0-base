package common

import (
	"encoding/binary"
	"fmt"
	"net"
)

const HEADER_SIZE = 4

type MessageHandler struct {
	conn net.Conn
}

func NewMessageHandler(conn net.Conn) *MessageHandler {
	return &MessageHandler{
		conn: conn,
	}
}

func (m *MessageHandler) sendMessage(msg []byte, endFlag bool) error {
	flag := int8(0)
	if endFlag {
		flag = 1
	}

	size := len(msg)
	
	err := binary.Write(m.conn, binary.BigEndian, int32(size))
    if err != nil {
		return err
		}
		
	err = binary.Write(m.conn, binary.BigEndian, int8(flag))
	if err != nil {
		return err
	}
			
	totalSent := 0

	for totalSent < size {
		n, err := m.conn.Write((msg[totalSent:]))
		if err != nil {
			return err
		}
		totalSent += n
	}

	fmt.Println("Total bytes sent: ", totalSent)
	return nil
}

func (m *MessageHandler) receiveMessage() (string, error) {
	size_buf := make([]byte, HEADER_SIZE)

    n, err := m.conn.Read(size_buf[:HEADER_SIZE])
    if err != nil {
        return "", err
    }

    size := binary.BigEndian.Uint32(size_buf[:n])

	buf := make([]byte, size)
	totalRead := 0

	for uint32(totalRead) < size {
		n, err := m.conn.Read(buf[totalRead:])
		if err != nil {
			return "", err
		}
		totalRead += n
	}
	return string(buf), nil
}

func (m *MessageHandler) close() {
	m.conn.Close()
}
