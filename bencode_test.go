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
