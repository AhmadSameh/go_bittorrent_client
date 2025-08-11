package handshake

import (
	"fmt"
	"io"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func New(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (h Handshake) Serialize() []byte {
	handshake := make([]byte, len(h.Pstr)+49)
	handshake[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(handshake[curr:], h.Pstr)
	curr += copy(handshake[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(handshake[curr:], h.InfoHash[:])
	curr += copy(handshake[curr:], h.PeerID[:])
	return handshake
}

func Read(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrlen := int(lengthBuf[0])
	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return nil, err
	}
	handshakeBuf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}
	var infoHash, peerID [20]byte
	const reservedBytes = 8
	const pstrOffset = 20
	pstr := string(handshakeBuf[:pstrlen])

	copy(infoHash[:], handshakeBuf[pstrlen+reservedBytes:pstrlen+reservedBytes+pstrOffset])
	copy(peerID[:], handshakeBuf[pstrlen+reservedBytes+pstrOffset:])

	return &Handshake{
		Pstr:     pstr,
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}
