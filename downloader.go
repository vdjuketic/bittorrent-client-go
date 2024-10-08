package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/k0kubun/go-ansi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
)

const defaultBlockSize int = 16 * 1024

const (
	WAITING     = "waiting"
	IN_PROGRESS = "in progress"
	COMPLETE    = "complete"
)

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

type PeerRequestMessage struct {
	lengthPrefix uint32
	id           uint8
	index        uint32
	begin        uint32
	length       uint32
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
	progressBar := getProgressBar(int(numJobs))

	jobs := make(chan Piece, numJobs)
	results := make(chan Result, numJobs)
	errors := make(chan Piece, numJobs)

	var wg sync.WaitGroup

	// Create jobs for each piece
	for _, piece := range pieces {
		jobs <- piece
		log.Debug().Msg(fmt.Sprintf("added job for piece %d to job queue", piece.number))
	}

	// Create a goroutine for each peer
	for i := range peers {
		wg.Add(1)
		go func(worker Peer) {
			defer wg.Done()
			downloadTorrentPieceWorker(torrentMeta, worker, jobs, errors, results)
		}(peers[i])
	}

	// Create a worker to add back failed jobs to job queue
	wg.Add(1)
	go func() {
		defer wg.Done()
		addBackFailedJobs(jobs, errors)
	}()

	// Check if all pieces are downloaded and stop all workers
	for {
		finishedJobs := len(results)

		progressBar.Set(finishedJobs)

		if finishedJobs == numJobs {
			log.Debug().Msg("All jobs finished, stopping workers")
			close(jobs)
			close(errors)
			close(results)
			break
		}
		time.Sleep(time.Millisecond * 10)
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
		log.Debug().Msg(fmt.Sprintf("added piece %d back to job queue", piece.number))
	}
	log.Debug().Msg("Stopping addBackFailedJobs")
}

func downloadTorrentPieceWorker(torrentMeta TorrentMeta, peer Peer, jobs <-chan Piece, errors chan<- Piece, results chan<- Result) {
	for piece := range jobs {
		log.Debug().Msg(fmt.Sprintf("[Peer %d] started downloading piece: %d", peer.id, piece.number))
		piece.status = IN_PROGRESS
		result, err := downloadTorrentPiece(torrentMeta, peer.address, piece.number)
		if err != nil {
			piece.status = WAITING
			errors <- piece
			log.Debug().Msg(fmt.Sprintf("[Peer %d] failed downloading piece: %d - %s", peer.id, piece.number, err))
		} else {
			piece.status = COMPLETE

			res := Result{piece: piece.number, result: result}
			results <- res

			log.Debug().Msg(fmt.Sprintf("[Peer %d] downloaded piece: %d", peer.id, piece.number))
		}
	}
	log.Debug().Msg(fmt.Sprintf("[Peer %d] stopped", peer.id))
}

func downloadTorrentPiece(torrentMeta TorrentMeta, peer string, piece int) ([]byte, error) {
	conn, err := peerHandshake(peer, torrentMeta.InfoHashBytes)
	if err != nil {
		return nil, err
	}

	err = exchangePeerMessages(conn, piece)
	if err != nil {
		return nil, err
	}

	pieceOffset := 0

	pieceLength := getPieceLength(piece, torrentMeta)
	blocks := int(math.Ceil(float64(pieceLength) / float64(defaultBlockSize)))

	// pipeline block request messages
	for range blocks {
		nextLength := pieceLength - pieceOffset
		blockSize := math.Min(float64(defaultBlockSize), float64(nextLength))

		payload := PeerRequestMessage{
			lengthPrefix: 13,
			id:           request,
			index:        uint32(piece),
			begin:        uint32(pieceOffset),
			length:       uint32(blockSize),
		}
		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, payload)

		sendMessageToPeer(conn, buf.Bytes())

		pieceOffset += int(blockSize)
	}

	// receive block messages and assemble the piece
	var downloadedPiece []byte
	for range blocks {
		peerDataMessage, err := receivePieceMessageFromPeer(conn)
		if err != nil {
			return nil, errors.New("error receiving data message")
		}

		downloadedPiece = append(downloadedPiece, peerDataMessage.data...)
	}

	downloadedPieceHash := convertToPieceHash(downloadedPiece)

	if downloadedPieceHash != torrentMeta.Pieces[piece] {
		return nil, errors.New("integrity check failed")
	}

	defer conn.Close()
	return downloadedPiece, nil
}

func getProgressBar(numJobs int) *progressbar.ProgressBar {
	var bar *progressbar.ProgressBar

	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		bar = progressbar.NewOptions(numJobs,
			progressbar.OptionSetVisibility(false))
	} else {
		bar = progressbar.NewOptions(numJobs,
			progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(15),
			//TODO change when multiple file torrents are supported
			progressbar.OptionSetDescription("[cyan][1/1][reset] Downloading file..."),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[green]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}))
	}
	return bar
}
