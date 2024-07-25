package main

type Tracker struct {
	InfoHash   string
	PeerId     string
	Port       int
	Uploaded   int
	Downloaded int
	Left       int
	Compact    int
}

func fromTorrentMeta(torrentMeta TorrentMeta) Tracker {
	tracker := Tracker{}
	tracker.InfoHash = torrentMeta.InfoHash
	// Hardcoded currently, change when appropriate
	tracker.PeerId = "00112233445566778899"
	tracker.Port = 6881
	tracker.Uploaded = 0
	tracker.Downloaded = 0
	tracker.Left = torrentMeta.Length
	tracker.Compact = 1

	return tracker
}
