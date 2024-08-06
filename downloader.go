package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

const defaultBlockSize int = 16 * 1024

const peerHadshakeTimeout time.Duration = time.Duration(5 * time.Second)
const (
	piece      = 7
	request    = 6
	bitfield   = 5
	interested = 2
	unchoke    = 1
)
const (
	WAITING     = "waiting"
	IN_PROGRESS = "in progress"
	COMPLETE    = "complete"
)

type PeerMessage struct {
	lengthPrefix uint32
	id           uint8
	index        uint32
	begin        uint32
	length       uint32
}

type Piece struct {
	number int
	status string
}

type Peer struct {
	id      int
	address string
	status  string
}

type Result struct {
	piece  int
	result []byte
}

func downloadTorrent(file string) []byte {
	torrentMeta := fromBencode(string(file))
	torrentMeta.printTree()
	peers := []Peer{}

	for j, address := range getPeers(torrentMeta) {
		peer := Peer{j, address, "idle"}
		peers = append(peers, peer)
	}

	pieces := []Piece{}

	for i := 0; i < len(torrentMeta.Pieces); i++ {
		piece := Piece{i, WAITING}
		pieces = append(pieces, piece)
	}

	return downloadTorrentPieces(torrentMeta, pieces, peers)
}

func downloadTorrentPieces(torrentMeta TorrentMeta, pieces []Piece, peers []Peer) []byte {
	numJobs := len(pieces)
	jobs := make(chan Piece, numJobs)
	results := make(chan Result, numJobs)
	errors := make(chan Piece, numJobs)

	var wg sync.WaitGroup

	for _, piece := range pieces {
		jobs <- piece
	}

	for i := range peers {
		wg.Add(1)
		go func(worker Peer) {
			defer wg.Done()
			downloadTorrentPieceWorker(torrentMeta, worker, jobs, errors, results)
		}(peers[i])
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		addBackFailedJobs(jobs, errors)
	}()

	for {
		if len(results) == numJobs {
			fmt.Println("Stopping workers")
			close(jobs)
			close(errors)
			close(results)
			break
		}
		time.Sleep(time.Second * 3)
	}

	wg.Wait()

	var totalResults []Result
	for r := range results {
		totalResults = append(totalResults, r)
	}

	sort.Slice(totalResults, func(i, j int) bool {
		return totalResults[i].piece < totalResults[j].piece
	})

	var res []byte
	for _, r := range totalResults {
		res = append(res, r.result...)
	}

	return res
}

func addBackFailedJobs(jobs chan<- Piece, errors <-chan Piece) {
	for piece := range errors {
		jobs <- piece
	}
	fmt.Println("Stopping addBackFailedJobs")
}

func downloadTorrentPieceWorker(torrentMeta TorrentMeta, peer Peer, jobs <-chan Piece, errors chan<- Piece, results chan<- Result) {
	for piece := range jobs {
		fmt.Printf("[Peer %d] started downloading piece: %d\n", peer.id, piece.number)
		piece.status = IN_PROGRESS
		result, err := downloadTorrentPiece(torrentMeta, peer.address, piece.number)
		if err != nil {
			piece.status = WAITING
			errors <- piece
			fmt.Printf("[Peer %d] failed downloading piece: %d\n", peer.id, piece.number)
		} else {
			piece.status = COMPLETE

			res := Result{piece: piece.number, result: result}
			results <- res

			fmt.Printf("[Peer %d] downloaded piece: %d\n", peer.id, piece.number)
		}
	}
	fmt.Printf("[Peer %d] stopped\n", peer.id)
}

func downloadTorrentPiece(torrentMeta TorrentMeta, peer string, piece int) ([]byte, error) {
	conn, err := peerHandshake(peer, torrentMeta.InfoHashBytes)
	if err != nil {
		return nil, fmt.Errorf("[%s] Handshake failed", peer)
	}

	err = exchangePeerMessages(conn, peer)
	if err != nil {
		return nil, err
	}

	pieceOffset := 0
	var downloadedPiece []byte

	pieceLength := getPieceLength(piece, torrentMeta)

	for pieceOffset < pieceLength {
		nextLength := pieceLength - pieceOffset
		blockSize := math.Min(float64(defaultBlockSize), float64(nextLength))

		payload := PeerMessage{
			lengthPrefix: 13,
			id:           request,
			index:        uint32(piece),
			begin:        uint32(pieceOffset),
			length:       uint32(blockSize),
		}
		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, payload)

		sendMessageToPeer(conn, buf.Bytes())

		data, err := receiveDataMessageFromPeer(conn)
		if err != nil {
			return nil, fmt.Errorf("[%s] Error receiving data message", peer)
		}

		downloadedPiece = append(downloadedPiece, data...)

		pieceOffset += int(blockSize)
	}

	downloadedPieceHash := convertToPieceHash(downloadedPiece)

	if downloadedPieceHash != torrentMeta.Pieces[piece] {
		return nil, fmt.Errorf("[%s] Integrity check failed", peer)
	}

	defer conn.Close()
	return downloadedPiece, nil
}
