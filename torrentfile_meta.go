package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/jackpal/bencode-go"
)

type TorrentMeta struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
	CreatedBy   string
}

type BencodeTorrent struct {
	Announce string
	Info     TorrentInfo
}

type TorrentInfo struct {
	Pieces      string
	PieceLength int `bencode:"piece length"`
	Length      int
	Name        string
	CreatedBy   string `bencode:"created by"`
}

func fromBencode(r io.Reader) TorrentMeta {
	decodedTorrent := BencodeTorrent{}
	err := bencode.Unmarshal(r, &decodedTorrent)
	if err != nil {
		fmt.Println("Failed to parse torrent file.")
		panic(err)
	}

	info := decodedTorrent.Info

	torrentMeta := TorrentMeta{}
	torrentMeta.Announce = decodedTorrent.Announce
	torrentMeta.InfoHash = getInfoHash(info)
	torrentMeta.PieceHashes = getPieceHashes(info.Pieces)
	torrentMeta.PieceLength = info.PieceLength
	torrentMeta.Length = info.Length
	torrentMeta.Name = info.Name
	torrentMeta.CreatedBy = info.CreatedBy

	fmt.Println(torrentMeta)
	return torrentMeta
}

func getPieceHashes(pieces string) [][20]byte {
	hashLen := 20
	buf := []byte(pieces)
	if len(buf)%hashLen != 0 {
		panic("Failed to split piece hashes")
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes
}

func getInfoHash(torrentInfo TorrentInfo) [20]byte {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, torrentInfo)
	if err != nil {
		fmt.Println("Failed to get info hash.")
		panic(err)
	}
	return sha1.Sum(buf.Bytes())
}
