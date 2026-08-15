package service

import "testing"

func TestPlatformBoolEnabled(t *testing.T) {
	tests := []struct {
		value    string
		fallback bool
		want     bool
	}{
		{value: "", fallback: true, want: true},
		{value: "", fallback: false, want: false},
		{value: "1", want: true},
		{value: "true", want: true},
		{value: "on", want: true},
		{value: "0", want: false},
		{value: "false", want: false},
	}
	for _, tt := range tests {
		if got := platformBoolEnabled(tt.value, tt.fallback); got != tt.want {
			t.Fatalf("platformBoolEnabled(%q, %t) = %t, want %t", tt.value, tt.fallback, got, tt.want)
		}
	}
}
