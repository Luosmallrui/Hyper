package dao

import (
	"reflect"
	"testing"
)

func TestParseContentTagIDs(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []int64
		wantErr bool
	}{
		{name: "empty", raw: "", want: []int64{}},
		{name: "sorted and deduplicated", raw: "9,2,9,5", want: []int64{2, 5, 9}},
		{name: "whitespace", raw: " 3, 8 ", want: []int64{3, 8}},
		{name: "invalid text", raw: "1,nope", wantErr: true},
		{name: "invalid zero", raw: "0", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseContentTagIDs(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseContentTagIDs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseContentTagIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}
