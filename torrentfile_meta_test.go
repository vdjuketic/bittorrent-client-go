package main

import (
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
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

// Helper function to capture output
func captureOutput(f func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	os.Stdout = orig
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestPrintTree_SingleFile(t *testing.T) {
	torrent := TorrentMeta{
		Name: "singlefile.txt",
		Keys: []File{},
	}

	output := captureOutput(torrent.printTree)

	expected := "singlefile.txt\n"
	if output != expected {
		t.Errorf("Expected %q but got %q", expected, output)
	}
}

func TestPrintTree_MultiFile(t *testing.T) {
	torrent := TorrentMeta{
		Name: "multifile",
		Keys: []File{
			{path: []string{"dir1", "file1.txt"}},
			{path: []string{"dir2", "file2.txt"}},
			{path: []string{"dir3", "file3.txt"}},
		},
	}

	output := captureOutput(torrent.printTree)

	expected := strings.Join([]string{
		"dir1/file1.txt",
		"dir2/file2.txt",
		"dir3/file3.txt",
		"",
	}, "\n")
	if output != expected {
		t.Errorf("Expected %q but got %q", expected, output)
	}
}
