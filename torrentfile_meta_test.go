package main

import (
	"reflect"
	"strconv"
	"testing"
)

func TestGetPieceLength(t *testing.T) {
	tests := []struct {
		pieceNum    int
		torrentMeta TorrentMeta
		want        int
	}{
		{
			pieceNum: 0,
			torrentMeta: TorrentMeta{
				Pieces:      []string{"piece1", "piece2", "piece3"},
				PieceLength: 256,
				Length:      700,
			},
			want: 256,
		},
		{
			pieceNum: 2,
			torrentMeta: TorrentMeta{
				Pieces:      []string{"piece1", "piece2", "piece3"},
				PieceLength: 256,
				Length:      700,
			},
			want: 188, // 700 % 256 = 188
		},
	}

	for _, tt := range tests {
		t.Run(
			// Use piece number and length to make test name unique and descriptive
			t.Name()+"/pieceNum="+strconv.Itoa(tt.pieceNum)+"/Length="+strconv.Itoa(tt.torrentMeta.Length),
			func(t *testing.T) {
				got := getPieceLength(tt.pieceNum, tt.torrentMeta)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("getPieceLength() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
