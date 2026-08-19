package types

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestVerifiedListItemIncludesNavigationAndBuyerPhone(t *testing.T) {
	body, err := json.Marshal(VerifiedListItem{
		OrderNo: "T202608180001", ActivityID: 10, PosterList: "https://cdn.hypercn.cn/poster.png",
		BuyerPhoneMasked: "138****5678", VerifiedAt: time.Date(2026, 8, 18, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"order_no":"T202608180001"`, `"activity_id":10`, `"poster_list"`, `"buyer_phone_masked":"138****5678"`} {
		if !strings.Contains(string(body), field) {
			t.Fatalf("verified-list contract missing %s: %s", field, body)
		}
	}
}

func TestOrganizerRealtimeOrderAndVerifierFields(t *testing.T) {
	verifiedAt := time.Date(2026, 8, 18, 10, 0, 0, 0, time.UTC)
	body, err := json.Marshal(OrganizerOrderListItem{
		BuyerPhoneMasked: "138****5678", PosterList: "https://cdn.hypercn.cn/poster.png", VerifiedAt: &verifiedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"buyer_phone_masked":"138****5678"`, `"poster_list"`, `"verified_at"`} {
		if !strings.Contains(string(body), field) {
			t.Fatalf("organizer order contract missing %s: %s", field, body)
		}
	}

	verifierBody, err := json.Marshal(OrganizerVerifierItem{VerifiedCount: 12})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(verifierBody), `"verified_count":12`) {
		t.Fatalf("verifier contract missing verified_count: %s", verifierBody)
	}
}
