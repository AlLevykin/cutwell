package utils

import (
	"testing"
)

func TestDecoder_Decode(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    string
		wantErr bool
	}{
		{"Hello", "a4aa1685e126ea2d3f6f960cc00f7af5949340d8d7", "Hello", false},
		{"Error", "Error", "", true},
	}
	d := NewDecoder()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.Decode(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Decode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecoder_Encode(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"Hello", "Hello", "a4aa1685e126ea2d3f6f960cc00f7af5949340d8d7"},
	}
	d := NewDecoder()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := d.Encode(tt.arg); got != tt.want {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDecoder(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"constructor"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDecoder(); got == nil {
				t.Errorf("NewDecoder() = %v, want %v", got, "not nil value")
			}
		})
	}
}
