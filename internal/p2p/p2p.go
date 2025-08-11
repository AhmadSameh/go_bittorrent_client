package p2p

import (
	"bittorrent_client/internal/peers"
	"log"
)

type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type workContainer struct {
	index  int
	hash   [20]byte
	length int
}

type resultsContainer struct {
	index int
	buf   []byte
}

func (t Torrent) calculatePieceSize(index int) int {
	begin := index * t.PieceLength
	end := begin + t.PieceLength
	return end - begin
}

func (t Torrent) DownloadPiece(peer peers.Peer, workBuf chan *workContainer, results chan *resultsContainer) {
	
}

func (t Torrent) Download() ([]byte, error) {
	log.Println("Downloading", t.Name)
	workBuf := make(chan workContainer, len(t.PieceHashes))
	results := make(chan resultsContainer)

	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workBuf <- workContainer{index, hash, length}
	}

	for _, peer := t.Peers {
		go t.DownloadPiece(peer, workBuf, results)
	}

	buf := make([]byte, t.Length)
	downloadedPiece := 0
	for downloadedPiece < len(t.PieceHashes) {
		res := <- results
		begin := res.index * t.PieceLength
		end := begin + t.PieceLength
		copy(buf[begin:end], res.buf)
		downloadedPiece++
	}
	close(workBuf)
}
