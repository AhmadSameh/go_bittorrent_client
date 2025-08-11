package client

import (
	"bittorrent_client/internal/bitfield"
	"bittorrent_client/internal/handshake"
	"bittorrent_client/internal/message"
	"bittorrent_client/internal/peers"
	"bytes"
	"fmt"
	"net"
	"time"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.BitField
	peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func completeHandshake(conn net.Conn, infoHash [20]byte, peerID [20]byte) error {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	// create a handshake byte and send it serialized over the network
	req := handshake.New(infoHash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return err
	}
	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return fmt.Errorf("Expected infohash %x but got %x", res.InfoHash, infoHash)
	}
	return nil
}

func recvBitfield(conn net.Conn) (bitfield.BitField, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read()
	if err != nil {
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}
	return msg.Payload, nil
}

func InitTCP(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {
	// start connection with the peer
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		return nil, err
	}

	bf, err := recvBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bf,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil
}
