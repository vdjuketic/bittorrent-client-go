package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackpal/bencode-go"
)

func downloadTorrent(torrentMeta TorrentMeta) {
	peerUrls := getPeers(torrentMeta)
	fmt.Println(peerUrls)
}

func getPeers(torrentMeta TorrentMeta) []string {
	tracker := fromTorrentMeta(torrentMeta)

	req, err := http.NewRequest("GET", torrentMeta.Announce, nil)
	if err != nil {
		fmt.Println("Failed to get tracker data")
		panic(err)
	}

	q := req.URL.Query()
	q.Add("info_hash", string(tracker.InfoHash[:]))
	q.Add("peer_id", tracker.PeerId)
	q.Add("port", strconv.Itoa(tracker.Port))
	q.Add("uploaded", strconv.Itoa(tracker.Uploaded))
	q.Add("downloaded", strconv.Itoa(tracker.Downloaded))
	q.Add("left", strconv.Itoa(tracker.Left))
	q.Add("compact", strconv.Itoa(tracker.Compact))
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		panic(err)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	decodedBody, err := bencode.Decode(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		panic(err)
	}
	fmt.Println("client: response body: ", decodedBody)

	return []string{}
}
