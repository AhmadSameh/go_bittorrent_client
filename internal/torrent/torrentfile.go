package torrent

import (
	"bittorrent_client/internal/p2p"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

const Port uint16 = 6881 // Default port for BitTorrent

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PiecesHash  [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Info     bencodeInfo `bencode:"info"`
	Announce string      `bencode:"announce"`
}

func (bto bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	var tf TorrentFile
	var err error
	tf.Announce = bto.Announce
	tf.PieceLength = bto.Info.PieceLength
	tf.Length = bto.Info.Length
	tf.Name = bto.Info.Name
	tf.InfoHash, err = bto.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}
	tf.PiecesHash, err = bto.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}
	return tf, nil
}

func (info bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, info)
	if err != nil {
		return [20]byte{}, err
	}
	hashed := sha1.Sum(buf.Bytes())
	return hashed, nil
}

func (info bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	const hashLen = 20
	pieces := []byte(info.Pieces)
	if len(pieces)%hashLen != 0 {
		err := fmt.Errorf("received malformed pieces of length %d", len(pieces))
		return nil, err
	}
	numHashes := len(pieces) / hashLen
	piecedHashes := make([][20]byte, numHashes)
	for i := range numHashes {
		copy(piecedHashes[i][:], pieces[i*hashLen:(i+1)*hashLen])
	}
	return piecedHashes, nil
}

func OpenTorrent(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()
	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return TorrentFile{}, err
	}

	return bto.toTorrentFile()
}

func (tf TorrentFile) DownloadTorrent(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}
	peers, err := tf.RequestPeersFromTracker(peerID, Port)
	if err != nil {
		return err
	}

	tr := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PiecesHash,
		PieceLength: tf.PieceLength,
		Length:      tf.Length,
		Name:        tf.Name,
	}

	buf, err := tr.Download()
	if err != nil {
		return err
	}

	f, err := os.Create(tf.Name)
	defer f.Close()
	if err != nil {
		return err
	}

	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	return nil
}
