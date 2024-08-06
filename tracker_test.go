package main

import (
	"reflect"
	"testing"
)

func TestFromTorrentMeta(t *testing.T) {
	tests := []struct {
		torrentMeta TorrentMeta
		want        Tracker
	}{
		{
			torrentMeta: TorrentMeta{
				InfoHashBytes: []byte("12345678901234567890"),
				Length:        1000,
			},
			want: Tracker{
				InfoHash:   []byte("12345678901234567890"),
				PeerId:     "00112233445566778899",
				Port:       6881,
				Uploaded:   0,
				Downloaded: 0,
				Left:       1000,
				Compact:    1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.torrentMeta.InfoHashBytes), func(t *testing.T) {
			got := fromTorrentMeta(tt.torrentMeta)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromTorrentMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTrackerRequestQueryParams(t *testing.T) {
	tests := []struct {
		tracker Tracker
		want    string
	}{
		{
			tracker: Tracker{
				InfoHash:   []byte("12345678901234567890"),
				PeerId:     "00112233445566778899",
				Port:       6881,
				Uploaded:   0,
				Downloaded: 0,
				Left:       1000,
				Compact:    1,
			},
			want: "compact=1&downloaded=0&info_hash=12345678901234567890&left=1000&peer_id=00112233445566778899&port=6881&uploaded=0",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.tracker.InfoHash), func(t *testing.T) {
			got := tt.tracker.getTrackerRequestQueryParams()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTrackerRequestQueryParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeersStringToIpList(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input: string([]byte{
				192, 168, 1, 1, 0x1F, 0x90, // 192.168.1.1:8080
				127, 0, 0, 1, 0x00, 0x50, // 127.0.0.1:80
			}),
			expected: []string{
				"192.168.1.1:8080",
				"127.0.0.1:80",
			},
		},
		{
			input: string([]byte{
				10, 0, 0, 1, 0x1F, 0x90, // 10.0.0.1:8080
			}),
			expected: []string{
				"10.0.0.1:8080",
			},
		},
		{
			input:    string([]byte{}),
			expected: []string{},
		},
	}

	for _, test := range tests {
		result := peersStringToIpList(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("Expected length %d but got %d", len(test.expected), len(result))
		}
		for i := range result {
			if result[i] != test.expected[i] {
				t.Errorf("Expected %s but got %s", test.expected[i], result[i])
			}
		}
	}
}
