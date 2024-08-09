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

func TestConvertToPieceHash(t *testing.T) {
	tests := []struct {
		name     string
		piece    []byte
		expected string
	}{
		{
			name:     "Empty slice",
			piece:    []byte{},
			expected: "da39a3ee5e6b4b0d3255bfef95601890afd80709", // SHA-1 hash of an empty string
		},
		{
			name:     "Single byte slice",
			piece:    []byte{0x61},                               // 'a'
			expected: "86f7e437faa5a7fce15d1ddcb9eaeaea377667b8", // SHA-1 hash of "a"
		},
		{
			name:     "String",
			piece:    []byte("The quick brown fox jumps over the lazy dog"),
			expected: "2fd4e1c67a2d28fced849ee1bb76e7391b93eb12", // SHA-1 hash of "The quick brown fox jumps over the lazy dog"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToPieceHash(tt.piece)
			if result != tt.expected {
				t.Errorf("convertToPieceHash(%v) = %v, expected %v", tt.piece, result, tt.expected)
			}
		})
	}
}

func TestGetPieceHashes(t *testing.T) {
	tests := []struct {
		name     string
		pieces   string
		expected []string
	}{
		{
			name:     "Empty string",
			pieces:   "",
			expected: []string{},
		},
		{
			name:   "Hashes",
			pieces: "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghij", // 50 bytes -> 100 hex chars
			expected: []string{
				"6162636465666768696a6162636465666768696a", // hex encoding of "abcdefghijabcdefghij"
				"6162636465666768696a6162636465666768696a", // hex encoding of "abcdefghijabcdefghij"
				"6162636465666768696a",                     // hex encoding of "abcdefghij"
			},
		},
		{
			name:   "Exact multiple of hash length",
			pieces: "01234567890123456789012345678901", // 32 bytes -> 64 hex chars
			expected: []string{
				"3031323334353637383930313233343536373839", // hex encoding of "01234567890123456789"
				"303132333435363738393031",                 // hex encoding of "012345678901"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPieceHashes(tt.pieces)
			if len(result) != len(tt.expected) {
				t.Errorf("getPieceHashes(%q) returned %v, expected %v", tt.pieces, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("getPieceHashes(%q)[%d] = %v, expected %v", tt.pieces, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestGetLength(t *testing.T) {
	tests := []struct {
		name        string
		decodedInfo map[string]interface{}
		expected    int
	}{
		{
			name: "Single file torrent with length",
			decodedInfo: map[string]interface{}{
				"length": 1024,
			},
			expected: 1024,
		},
		{
			name: "Multi file torrent",
			decodedInfo: map[string]interface{}{
				"keys": []File{
					{length: 1024},
					{length: 2048},
					{length: 4096},
				},
			},
			expected: 7168, // sum of all file lengths
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getLength(tt.decodedInfo)
			if result != tt.expected {
				t.Errorf("getLength(%v) = %d, expected %d", tt.decodedInfo, result, tt.expected)
			}
		})
	}
}
