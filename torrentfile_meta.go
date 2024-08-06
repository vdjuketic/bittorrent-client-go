package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

type TorrentMeta struct {
	Announce      string
	InfoHash      string
	InfoHashBytes []byte
	Pieces        []string
	PieceLength   int
	Length        int
	Keys          []File
	Name          string
	CreatedBy     string
}

type File struct {
	length int
	path   []string
}

func fromBencode(bencode string) TorrentMeta {
	decoded, err := decodeBencode(bencode)
	if err != nil {
		fmt.Println("Failed to parse torrent file.")
		panic(err)
	}

	decodedTorrent := decoded.(map[string]interface{})
	decodedInfo := decodedTorrent["info"].(map[string]interface{})

	meta := TorrentMeta{}
	meta.Announce = fmt.Sprint(decodedTorrent["announce"])
	meta.InfoHash = getInfoHash(decodedTorrent["info"])
	meta.Pieces = getPieceHashes(decodedInfo["pieces"].(string))
	meta.PieceLength = decodedInfo["piece length"].(int)
	meta.Name = fmt.Sprint(decodedInfo["name"])
	meta.CreatedBy = fmt.Sprint(decodedTorrent["created by"])
	meta.Keys = getKeys(decodedInfo)
	meta.Length = getLength(decodedInfo)

	meta.InfoHashBytes, err = hex.DecodeString(meta.InfoHash)
	if err != nil {
		fmt.Println("failed to decode info hash")
		panic(err)
	}

	return meta
}

func getKeys(decodedInfo map[string]interface{}) []File {
	// if length is provided it's a single file torrent
	// if not then it's a multi file torrent with file structure provided in keys
	_, ok := decodedInfo["length"]
	if !ok {
		return decodedInfo["keys"].([]File)
	}
	return nil
}

func getLength(decodedInfo map[string]interface{}) int {
	_, ok := decodedInfo["length"]
	if ok {
		// if length is provided it's a single file torrent with that length
		return decodedInfo["length"].(int)
	} else {
		// if length is not provided it's a multi file torrent
		// it's length is the sum of length of all individual files
		sumLength := 0
		for _, file := range decodedInfo["keys"].([]File) {
			sumLength += file.length
		}
		return sumLength
	}
}

func getInfoHash(infoDict interface{}) string {
	encoded, err := encodeBencode(infoDict)
	if err != nil {
		fmt.Println("failed to encode info hash")
		panic(err)
	}
	return convertToPieceHash(encoded)
}

func convertToPieceHash(piece []byte) string {
	hash := sha1.Sum(piece)
	var result string
	for _, number := range hash {
		result += fmt.Sprintf("%02x", number)
	}
	return result
}

func getPieceHashes(pieces string) []string {
	piecesAsBytes := []byte(pieces)
	piecesAsHexStr := hex.EncodeToString(piecesAsBytes)
	return splitString(piecesAsHexStr, 40)
}

func splitString(s string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

func getPieceLength(pieceNum int, torrentMeta TorrentMeta) int {
	numOfPieces := len(torrentMeta.Pieces)
	if numOfPieces-1 != pieceNum {
		return torrentMeta.PieceLength
	}

	return torrentMeta.Length % torrentMeta.PieceLength
}

func (t TorrentMeta) printTree() {
	if len(t.Keys) == 0 {
		// single file torrent
		fmt.Println(t.Name)
	} else {
		// multi file torrent
		for _, file := range t.Keys {
			fullPath := strings.Join(file.path[:], "/")
			fmt.Println(fullPath)
		}
	}
}
