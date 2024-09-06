package common

import (
	"encoding/binary"
	"net"
)

const PACKET_SIZE = 4
const FLAG_SIZE = 1

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

func (m *MessageHandler) receiveMessage() (string, int8, error) {
	size_bytes, err := m.readMessage(PACKET_SIZE)
	if err != nil {
		return "", 0, err
	}

	size := int32(binary.BigEndian.Uint32(size_bytes))

	flag_bytes, err := m.readMessage(FLAG_SIZE)
	if err != nil {
		return "",0, err
	}
	flag := int8(flag_bytes[0])

	msg, err := m.readMessage(size)
	return string(msg), flag, err
}

func (m *MessageHandler) readMessage(size int32) ([]byte, error) {
	buf := make([]byte, size)
	totalRead := 0

	for int32(totalRead) < size {
		n, err := m.conn.Read(buf[totalRead:])
		if err != nil {
			return nil, err
		}
		totalRead += n
	}
	return buf, nil
}

func (m *MessageHandler) close() {
	m.conn.Close()
}
