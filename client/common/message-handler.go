package common

import (
	"encoding/binary"
	"net"
)

const PACKET_SIZE = 4

type MessageHandler struct {
	conn net.Conn
}

func NewMessageHandler(conn net.Conn) *MessageHandler {
	return &MessageHandler{
		conn: conn,
	}
}

func (m *MessageHandler) sendHeader(size int32, endFlag bool) error {
	flag := int8(0)
	if endFlag {
		flag = 1
	}
	
	err := binary.Write(m.conn, binary.BigEndian, int32(size))
    if err != nil {
		return err
	}
		
	err = binary.Write(m.conn, binary.BigEndian, int8(flag))
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageHandler) sendMessage(msg []byte, endFlag bool) error {
	err := m.sendHeader(int32(len(msg)), endFlag)

	if err != nil {
		return err
	}
			
	totalSent := 0

	for totalSent < len(msg) {
		n, err := m.conn.Write((msg[totalSent:]))
		if err != nil {
			return err
		}
		totalSent += n
	}

	return nil
}

func (m *MessageHandler) readHeader() (uint32, error) {
	buf := make([]byte, PACKET_SIZE)

	n, err := m.conn.Read(buf[:PACKET_SIZE])
	if err != nil {
		return 0, err
	}

	size := binary.BigEndian.Uint32(buf[:n])
	return size, nil
}

func (m *MessageHandler) receiveMessage() (string, error) {
	size, err := m.readHeader()
	if err != nil {
		return "", err
	}

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
