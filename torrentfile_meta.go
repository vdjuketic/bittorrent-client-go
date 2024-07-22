package main

import (
	"encoding/json"
	"fmt"
)

type TorrentMeta struct {
	Announce    string
	InfoHash    []byte
	PieceHashes [][]byte
	PieceLength int `json:"piece length"`
	Length      int
	Name        string
	Info        string
	CreatedBy   string `json:"created by"`
}

type BencodeTorrent struct {
	Announce string
	Info     TorrentInfo
}

type TorrentInfo struct {
	//Pieces      string
	PieceLength int `json:"piece length"`
	Length      int
	Name        string
}

func fromString(source string) TorrentMeta {
	torrentMeta := TorrentMeta{}

	err := json.Unmarshal([]byte(source), &torrentMeta)
	if err != nil {
		fmt.Println("Couldn't unmarshal torrent meta.")
		panic(err)
	}

	torrentInfo := TorrentInfo{}
	err = json.Unmarshal([]byte(torrentMeta.Info), &torrentInfo)
	if err != nil {
		fmt.Println("Couldn't unmarshal torrent info.")
		panic(err)
	}

	return torrentMeta
}
