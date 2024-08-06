package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Tracker struct {
	InfoHash   []byte
	PeerId     string
	Port       int
	Uploaded   int
	Downloaded int
	Left       int
	Compact    int
}

func fromTorrentMeta(torrentMeta TorrentMeta) Tracker {
	tracker := Tracker{}
	tracker.InfoHash = torrentMeta.InfoHashBytes
	tracker.PeerId = "00112233445566778899"
	tracker.Port = 6881
	tracker.Uploaded = 0
	tracker.Downloaded = 0
	tracker.Left = torrentMeta.Length
	tracker.Compact = 1

	return tracker
}

func (t Tracker) getTrackerRequestQueryParams() string {
	params := url.Values{}
	params.Add("info_hash", string(t.InfoHash))
	params.Add("peer_id", t.PeerId)
	params.Add("port", fmt.Sprint(t.Port))
	params.Add("uploaded", fmt.Sprint(t.Uploaded))
	params.Add("downloaded", fmt.Sprint(t.Downloaded))
	params.Add("left", fmt.Sprint(t.Left))
	params.Add("compact", fmt.Sprint(t.Compact))

	return params.Encode()
}

func (t Tracker) getPeers(trackerUrl string) []string {
	params := t.getTrackerRequestQueryParams()
	url := fmt.Sprintf("%s?%s", trackerUrl, params)

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

	fmt.Println(decodedBody.(map[string]interface{}))
	peersField := decodedBody.(map[string]interface{})["peers"]

	//add timeout based on interval field
	if peersField == nil {
		t.getPeers(trackerUrl)
	}

	peersString := peersField.(string)

	return peersStringToIpList(peersString)
}

func peersStringToIpList(peersString string) []string {
	peers := make([]string, 0)
	for k := 0; k < len(peersString); k += 6 {
		peer := strconv.Itoa(int(peersString[k])) + "." +
			strconv.Itoa(int(peersString[k+1])) + "." +
			strconv.Itoa(int(peersString[k+2])) + "." +
			strconv.Itoa(int(peersString[k+3])) + ":" +
			strconv.Itoa(int((binary.BigEndian.Uint16)([]byte(peersString[k+4:k+6]))))
		peers = append(peers, peer)
	}
	return peers
}
