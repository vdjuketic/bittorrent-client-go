package main

import (
	"fmt"
	"net/url"
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

func getTrackerRequestQueryParams(tracker Tracker) string {
	params := url.Values{}
	params.Add("info_hash", string(tracker.InfoHash))
	params.Add("peer_id", tracker.PeerId)
	params.Add("port", fmt.Sprint(tracker.Port))
	params.Add("uploaded", fmt.Sprint(tracker.Uploaded))
	params.Add("downloaded", fmt.Sprint(tracker.Downloaded))
	params.Add("left", fmt.Sprint(tracker.Left))
	params.Add("compact", fmt.Sprint(tracker.Compact))

	return params.Encode()
}
