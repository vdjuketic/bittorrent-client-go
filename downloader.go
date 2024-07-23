package main

import (
	"fmt"
	//"net/http"
)

func downloadTorrent(torrentMeta TorrentMeta) {
	peerUrls := getPeers(torrentMeta)
	fmt.Println(peerUrls)
}

func getPeers(torrentMeta TorrentMeta) []string {
	//tracker := fromTorrentMeta(torrentMeta)

	//response := http.Request()

	return []string{}
}
