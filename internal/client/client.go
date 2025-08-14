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
		return fmt.Errorf("expected infohash %x but got %x", res.InfoHash, infoHash)
	}
	return nil
}

func recvBitfield(conn net.Conn) (bitfield.BitField, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("expected bitfield but got ID %d", msg.ID)
		return nil, err
	}
	return msg.Payload, nil
}

func (client Client) ReadMessage() (*message.Message, error) {
	msg, err := message.Read(client.Conn)
	return msg, err
}

func (client Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func (client Client) SendInterested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func (client Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func (client Client) SendRequest(index, begin, length int) error {
	req := message.FormatRequest(index, begin, length)
	_, err := client.Conn.Write(req.Serialize())
	return err
}

func ConnectWithPeer(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {
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
