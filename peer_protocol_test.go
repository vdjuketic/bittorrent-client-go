package main

import (
	"testing"
)

func TestHasPiece(t *testing.T) {
	tests := []struct {
		name     string
		bitfield Bitfield
		index    int
		expected bool
	}{
		{
			name:     "First bit set",
			bitfield: Bitfield{0b10000000}, // binary: 10000000
			index:    0,
			expected: true,
		},
		{
			name:     "First bit not set",
			bitfield: Bitfield{0b01000000}, // binary: 01000000
			index:    0,
			expected: false,
		},
		{
			name:     "Middle bit set",
			bitfield: Bitfield{0b00100000}, // binary: 00100000
			index:    2,
			expected: true,
		},
		{
			name:     "Last bit set",
			bitfield: Bitfield{0b00000001}, // binary: 00000001
			index:    7,
			expected: true,
		},
		{
			name:     "Bit in second byte set",
			bitfield: Bitfield{0b00000000, 0b10000000}, // binary: 00000000 10000000
			index:    8,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.bitfield.hasPiece(tt.index)
			if result != tt.expected {
				t.Errorf("hasPiece(%d) = %v, expected %v", tt.index, result, tt.expected)
			}
		})
	}
}
