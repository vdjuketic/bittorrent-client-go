package main

import (
	"crypto/sha1"
	"fmt"
)

type TorrentMeta struct {
	Announce    string
	InfoHash    string
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
	CreatedBy   string
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
	meta.CreatedBy = fmt.Sprint(decodedTorrent["created by"])
	meta.InfoHash = getInfoHash(decodedTorrent["info"])

	meta.Length = decodedInfo["length"].(int)
	meta.Name = fmt.Sprint(decodedInfo["name"])

	fmt.Printf("Tracker URL: %s\n", meta.Announce)
	fmt.Printf("Length: %d\n", meta.Length)
	fmt.Printf("Info Hash: %s\n", meta.InfoHash[:])

	return meta
}

// func getPieceHashes(pieces string) [][20]byte {
// 	hashLen := 20
// 	buf := []byte(pieces)
// 	if len(buf)%hashLen != 0 {
// 		panic("Failed to split piece hashes")
// 	}
// 	numHashes := len(buf) / hashLen
// 	hashes := make([][20]byte, numHashes)

// 	for i := 0; i < numHashes; i++ {
// 		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
// 	}
// 	return hashes
// }

func getInfoHash(infoDict interface{}) string {
	encoding := encodeBencode(infoDict)
	hash := sha1.Sum([]byte(encoding))
	var result string
	for _, number := range hash {
		result += fmt.Sprintf("%02x", number)
	}
	return result
}
