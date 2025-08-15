package torrent

import (
	"bittorrent_client/internal/peers"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (tf TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(tf.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (tf TorrentFile) RequestPeersFromTracker(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	trackerURL, err := tf.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}
	client := http.Client{Timeout: 30 * time.Second}
	res, err := client.Get(trackerURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	trackerRes := bencodeTrackerResp{}
	err = bencode.Unmarshal(res.Body, &trackerRes)
	if err != nil {
		return nil, err
	}
	peers, err := peers.GetPeers([]byte(trackerRes.Peers))

	return peers, err
}
