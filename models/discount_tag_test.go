package models

import "testing"

func TestDiscountTagBits(t *testing.T) {
	tests := []struct {
		name    string
		ids     []int
		want    int
		wantErr bool
	}{
		{name: "empty", ids: []int{}, want: 0},
		{name: "multiple tags", ids: []int{DiscountTagPoints, DiscountTagNewUser}, want: 5},
		{name: "duplicate tags", ids: []int{DiscountTagPoints, DiscountTagPoints}, want: 1},
		{name: "unknown tag", ids: []int{8}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DiscountTagBits(tt.ids)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DiscountTagBits() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("DiscountTagBits() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDiscountTagDisplayValues(t *testing.T) {
	bits := DiscountTagPoints | DiscountTagNewUser
	ids := DiscountTagIDs(bits)
	names := DiscountTagNames(bits)
	if len(ids) != 2 || ids[0] != DiscountTagPoints || ids[1] != DiscountTagNewUser {
		t.Fatalf("DiscountTagIDs() = %v, want [1 4]", ids)
	}
	if len(names) != 2 || names[0] != "积分立减" || names[1] != "新人优惠" {
		t.Fatalf("DiscountTagNames() = %v, want [积分立减 新人优惠]", names)
	}
}
