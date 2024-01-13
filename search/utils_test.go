package search

import "testing"

func TestUnit_isValidIndexName(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{name: "Case1", arg: "Aaaa", want: false},
		{name: "Case2", arg: "0-aaa", want: false},
		{name: "Case3", arg: "123_aaa", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidIndexName(tt.arg); got != tt.want {
				t.Errorf("isValidIndexName() = %v, want %v", got, tt.want)
			}
		})
	}
}
