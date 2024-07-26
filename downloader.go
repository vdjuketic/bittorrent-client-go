package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

const peerHadshakeTimeout time.Duration = time.Duration(5 * time.Second)

func downloadTorrent(file string) {
	torrentMeta := fromBencode(string(file))
	peers := getPeers(torrentMeta)
	peerHandshake(peers[0], torrentMeta.InfoHashBytes)
}

func getPeers(torrentMeta TorrentMeta) []string {
	tracker := fromTorrentMeta(torrentMeta)

	params := getTrackerRequestQueryParams(tracker)
	url := fmt.Sprintf("%s?%s", torrentMeta.Announce, params)

	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		panic("failed to get response from tracker")
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		panic("failed to read response from tracker")
	}

	decodedBody, err := decodeBencode(string(body))
	if err != nil {
		fmt.Println(err)
		panic("failed to decode response from tracker")
	}

	peersField := decodedBody.(map[string]interface{})["peers"]
	//add timeout based on interval field
	if peersField == nil {
		getPeers(torrentMeta)
	}

	peersString := peersField.(string)

	peers := make([]string, 0)
	for k := 0; k < len(peersString); k += 6 {
		peer := strconv.Itoa(int(peersString[k])) + "." + strconv.Itoa(int(peersString[k+1])) + "." + strconv.Itoa(int(peersString[k+2])) + "." + strconv.Itoa(int(peersString[k+3])) + ":" + strconv.Itoa(int((binary.BigEndian.Uint16)([]byte(peersString[k+4:k+6]))))
		peers = append(peers, peer)
	}

	for _, peer := range peers {
		fmt.Println(peer)
	}

	return peers
}

func peerHandshake(peerUrl string, infoHash []byte) {
	conn, err := net.Dial("tcp", peerUrl)
	if err != nil {
		fmt.Println(err)
		panic("error establishing connection to peer")
	}
	err = conn.SetReadDeadline(time.Now().Add(peerHadshakeTimeout))
	if err != nil {
		fmt.Println(err)
		panic("set deadline failed")
	}

	defer conn.Close()

	handshakeMsg := createHandshakeMessage(infoHash)

	_, err = conn.Write(handshakeMsg)
	if err != nil {
		fmt.Println(err)
		panic("error sending handshake to peer")
	}

	buf := make([]byte, 68)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		panic("error reading handshake from peer")
	}
	fmt.Printf("Peer ID: %s\n", hex.EncodeToString(buf[48:]))
}

func createHandshakeMessage(infoHash []byte) []byte {
	pstrlen := byte(19)
	pstr := []byte("BitTorrent protocol")
	reserved := make([]byte, 8)
	handshake := append([]byte{pstrlen}, pstr...)
	handshake = append(handshake, reserved...)
	handshake = append(handshake, infoHash...)
	handshake = append(handshake, []byte{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9}...)
	return handshake
}
