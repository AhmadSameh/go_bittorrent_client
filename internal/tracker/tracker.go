package tracker

import (
	"bittorrent_client/internal/torrentfile"
	"net/url"
	"strconv"
)

type Tracker struct {
	URL    string
	PeerID [20]byte
	Port   int
}

func (tr Tracker) BuildTrackerURL(torrent torrentfile.TorrentFile) (string, error) {
	base, err := url.Parse(torrent.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(torrent.InfoHash[:])},
		"peer_id":    []string{string(tr.PeerID[:])},
		"port":       []string{strconv.Itoa(tr.Port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(torrent.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}
