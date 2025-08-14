package p2p

import (
	"bittorrent_client/internal/client"
	"bittorrent_client/internal/message"
	"bittorrent_client/internal/peers"
	"log"
	"time"
)

const maxBackLog = 5

const maxBlockSize = 16384

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

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state pieceProgress) readMessage() error {
	msg, err := state.client.ReadMessage()
	if err != nil {
		return err
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		index, err := msg.ParsHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		downloaded, err := msg.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += downloaded
		state.backlog--
	}
	return nil
}

func (t Torrent) calculatePieceSize(index int) int {
	begin := index * t.PieceLength
	end := begin + t.PieceLength
	return end - begin
}

func attemptDownloadPiece(client *client.Client, workPiece *workContainer) ([]byte, error) {
	state := pieceProgress{
		index:  workPiece.index,
		client: client,
		buf:    make([]byte, workPiece.length),
	}

	client.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer client.Conn.SetDeadline(time.Time{})

	for state.downloaded < workPiece.length {
		if !state.client.Choked {
			for state.backlog < maxBackLog && state.requested < workPiece.length {
				blockSize := maxBlockSize
				if workPiece.length-state.requested < blockSize {
					blockSize = workPiece.length - state.requested
				}
				err := state.client.SendRequest(workPiece.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func (t Torrent) downloadPiece(peer peers.Peer, workBuf chan *workContainer, results chan *resultsContainer) {
	client, err := client.ConnectWithPeer(peer, t.PeerID, t.InfoHash)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer client.Conn.Close()
	log.Printf("Completed handshake with %s\n", peer.IP)

	client.SendUnchoke()
	client.SendInterested()

	for workPiece := range workBuf {
		if !client.Bitfield.HasPiece(workPiece.index) {
			workBuf <- workPiece
			continue
		}

		buf, err := attemptDownloadPiece(client, workPiece)
		if err != nil {
			log.Println("Exiting", err)
			workBuf <- workPiece
			return
		}

		err = checkIntegrity(workPiece, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", workPiece.index)
			workBuf <- workPiece
			continue
		}

		client.SendHave(workPiece.index)

		results <- &resultsContainer{workPiece.index, buf}
	}
}

func (t Torrent) Download() ([]byte, error) {
	log.Println("Downloading", t.Name)
	workBuf := make(chan *workContainer, len(t.PieceHashes))
	results := make(chan *resultsContainer)

	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workBuf <- &workContainer{index, hash, length}
	}

	for _, peer := range t.Peers {
		go t.downloadPiece(peer, workBuf, results)
	}

	buf := make([]byte, t.Length)
	downloadedPiece := 0
	for downloadedPiece < len(t.PieceHashes) {
		res := <-results
		begin := res.index * t.PieceLength
		end := begin + t.PieceLength
		copy(buf[begin:end], res.buf)
		downloadedPiece++
	}
	close(workBuf)

	return buf, nil
}
