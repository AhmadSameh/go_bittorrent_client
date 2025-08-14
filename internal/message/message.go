package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	MsgChoke         uint8 = 0
	MsgUnchoke       uint8 = 1
	MsgInterested    uint8 = 2
	MsgNotInterested uint8 = 3
	MsgHave          uint8 = 4
	MsgBitfield      uint8 = 5
	MsgRequest       uint8 = 6
	MsgPiece         uint8 = 7
	MsgCancel        uint8 = 8
)

type Message struct {
	ID      uint8
	Payload []byte
}

func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}

func (m Message) Serialize() []byte {
	length := uint32(1 + len(m.Payload))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

func (m Message) ParsHave(msg *Message) (int, error) {
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("expected payload length 4, got length %d", len(msg.Payload))
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

func (m Message) ParsePiece(index int, buf []byte, msg *Message) (int, error) {

	pieceIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if pieceIndex != index {
		return 0, fmt.Errorf("expected index %d, got %d", index, pieceIndex)
	}
	pieceOffset := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if pieceOffset >= len(buf) {
		return 0, fmt.Errorf("begin offset too high. %d >= %d", pieceOffset, len(buf))
	}
	data := msg.Payload[8:]
	if pieceOffset+len(data) > len(buf) {
		return 0, fmt.Errorf("data too long [%d] for offset %d with length %d", len(data), pieceOffset, len(buf))
	}
	copy(buf[pieceOffset:], data)
	return len(data), nil
}

func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}
	id := uint8(messageBuf[0])
	payload := messageBuf[1:]

	return &Message{
		ID:      id,
		Payload: payload,
	}, nil
}
