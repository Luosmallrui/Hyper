package models

import "fmt"

const (
	DiscountTagPoints  = 1 // 积分立减
	DiscountTagPayment = 2 // 买单立减
	DiscountTagNewUser = 4 // 新人优惠
)

// DiscountTagBits validates API tag IDs and converts them to the bit mask
// persisted on activities.discount_tags.
func DiscountTagBits(ids []int) (int, error) {
	bits := 0
	for _, id := range ids {
		switch id {
		case DiscountTagPoints, DiscountTagPayment, DiscountTagNewUser:
			bits |= id
		default:
			return 0, fmt.Errorf("优惠标签无效: %d", id)
		}
	}
	return bits, nil
}

func DiscountTagIDs(bits int) []int {
	ids := make([]int, 0, 3)
	for _, id := range []int{DiscountTagPoints, DiscountTagPayment, DiscountTagNewUser} {
		if bits&id == id {
			ids = append(ids, id)
		}
	}
	return ids
}

func DiscountTagNames(bits int) []string {
	names := make([]string, 0, 3)
	for _, tag := range []struct {
		id   int
		name string
	}{
		{DiscountTagPoints, "积分立减"},
		{DiscountTagPayment, "买单立减"},
		{DiscountTagNewUser, "新人优惠"},
	} {
		if bits&tag.id == tag.id {
			names = append(names, tag.name)
		}
	}
	return names
}
