package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type TrackerRequest struct {
	InfoHash   []byte
	PeerId     string
	Port       int
	Uploaded   int
	Downloaded int
	Left       int
	Compact    int
}

type Tracker struct {
	Complete       int
	Incomplete     int
	Interval       int
	MinInterval    int
	Peers          []string
	TrackerRequest TrackerRequest
}

func fromTorrentMeta(torrentMeta TorrentMeta) Tracker {
	tracker := Tracker{}

	request := TrackerRequest{}
	request.InfoHash = torrentMeta.InfoHashBytes
	request.PeerId = "00112233445566778899"
	request.Port = 6881
	request.Uploaded = 0
	request.Downloaded = 0
	request.Left = torrentMeta.Length
	request.Compact = 1

	tracker.TrackerRequest = request
	tracker.Peers, tracker.Interval = getTrackerData(tracker, torrentMeta.Announce)

	return tracker
}

// TODO periodically repeat this call according to interval field to refresh peer data
// Returns interval and list of peers
func getTrackerData(tracker Tracker, trackerUrl string) ([]string, int) {
	params := tracker.getTrackerRequestQueryParams()
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

	peersField := decodedBody.(map[string]interface{})["peers"]

	if peersField == nil {
		getTrackerData(tracker, trackerUrl)
	}

	peersString := peersField.(string)

	return peersStringToIpList(peersString), decodedBody.(map[string]interface{})["interval"].(int)
}

func (t Tracker) getTrackerRequestQueryParams() string {
	params := url.Values{}
	params.Add("info_hash", string(t.TrackerRequest.InfoHash))
	params.Add("peer_id", t.TrackerRequest.PeerId)
	params.Add("port", fmt.Sprint(t.TrackerRequest.Port))
	params.Add("uploaded", fmt.Sprint(t.TrackerRequest.Uploaded))
	params.Add("downloaded", fmt.Sprint(t.TrackerRequest.Downloaded))
	params.Add("left", fmt.Sprint(t.TrackerRequest.Left))
	params.Add("compact", fmt.Sprint(t.TrackerRequest.Compact))

	return params.Encode()
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
