package utils

import "testing"

func TestRandString(t *testing.T) {
	tests := []struct {
		name string
		arg  int
		want int
	}{
		{"n=1", 1, 1},
		{"n=2", 2, 2},
		{"n=3", 3, 3},
		{"n=4", 4, 4},
		{"n=5", 5, 5},
		{"n=10", 10, 10},
		{"n=100", 100, 100},
		{"n=1000", 1000, 1000},
		{"n=10000", 10000, 10000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandString(tt.arg); len(got) != tt.want {
				t.Errorf("RandString() = %v, want %v", got, tt.want)
			}
		})
	}
}
