package peers

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func GetPeers(peersEncoded []byte) ([]Peer, error) {
	const peerIPLen = 4
	const peerPortLen = 2
	const peerLen = peerIPLen + peerPortLen
	if len(peersEncoded)%peerLen != 0 {
		err := fmt.Errorf("received malformed peers")
		return nil, err
	}
	peersNumber := len(peersEncoded) / peerLen
	peers := make([]Peer, peersNumber)
	for i := range peersNumber {
		offset := i * peerLen
		peers[i].IP = net.IP(peersEncoded[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16(peersEncoded[offset+4 : offset+6])
	}
	return peers, nil
}
