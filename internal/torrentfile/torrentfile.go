package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/jackpal/bencode-go"
)

const Port uint16 = 6881                                                                                                                               // Default port for BitTorrent
var PeerID [20]byte = [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14} // Example PeerID

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PiecesHash  [][20]byte
	PieceLength int
	Length      int
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

func Open(r io.Reader) (TorrentFile, error) {
	bto := bencodeTorrent{}
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return TorrentFile{}, err
	}

	return bto.toTorrentFile()
}
