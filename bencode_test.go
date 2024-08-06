package main

import (
	"fmt"
	"reflect"
	"testing"
)

// Function under test: encodeBencode
func TestEncodeBencode(t *testing.T) {
	tests := []struct {
		data    interface{}
		want    []byte
		wantErr bool
	}{
		{
			data:    "hello",
			want:    []byte("5:hello"),
			wantErr: false,
		},
		{
			data:    42,
			want:    []byte("i42e"),
			wantErr: false,
		},
		{
			data:    []interface{}{"a", "b", "c"},
			want:    []byte("l1:a1:b1:ce"),
			wantErr: false,
		},
		{
			data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			want:    []byte("d4:key16:value14:key26:value2e"),
			wantErr: false,
		},
		{
			data:    struct{}{},
			want:    nil,
			wantErr: true,
		},
		{
			data: map[string]interface{}{
				"key": []interface{}{"nested", "list"},
			},
			want:    []byte("d3:keyl4:list6:nestedee"),
			wantErr: false,
		},
		{
			data: map[string]interface{}{
				"key": map[string]interface{}{"nestedKey": "nestedValue"},
			},
			want:    []byte("d3:keyd9:nestedKey11:nestedValueee"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("data=%v", tt.data), func(t *testing.T) {
			got, err := encodeBencode(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeBencode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeBencode() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDecodeBencode(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
		hasError bool
	}{
		// Test cases for string decoding
		{"4:spam", "spam", false},
		// Test cases for integer decoding
		{"i3e", 3, false},
		{"i-3e", -3, false},
		// Test cases for list decoding
		{"l4:spami42ee", []interface{}{"spam", 42}, false},
		// Test cases for dictionary decoding
		{"d3:cow3:moo4:spam4:eggse", map[string]interface{}{"cow": "moo", "spam": "eggs"}, false},
		// Test case for nested structures
		{"d4:spaml1:a1:bee", map[string]interface{}{"spam": []interface{}{"a", "b"}}, false},
		// Test cases for invalid bencode
		{"", "", true},
		{"i3", "", true},
		{"l4:spami42e", "", true},
		{"d3:cow3:moo4:spam4:egg", "", true},
	}

	for _, test := range tests {
		result, err := decodeBencode(test.input)
		if (err != nil) != test.hasError {
			t.Errorf("decodeBencode(%q) expected error: %v, got: %v", test.input, test.hasError, err)
		}
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("decodeBencode(%q) expected %v, got %v", test.input, test.expected, result)
		}
	}
}
