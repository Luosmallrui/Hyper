package service

import (
	"Hyper/types"
	"testing"
)

func TestNormalizeMarkerIcon(t *testing.T) {
	valid := "https://cdn.hypercn.cn/marker-icons/qiche.png"
	got, err := normalizeMarkerIcon(valid)
	if err != nil || got != valid {
		t.Fatalf("valid CDN icon rejected: got=%q err=%v", got, err)
	}
	if _, err := normalizeMarkerIcon("https://example.com/icon.png"); err == nil {
		t.Fatal("external icon host must be rejected")
	}
	if _, err := normalizeMarkerIcon("http://cdn.hypercn.cn/icon.png"); err == nil {
		t.Fatal("non-HTTPS icon must be rejected")
	}
}

func TestValidateVenueProfileInput(t *testing.T) {
	valid := types.OrganizerVenueProfileInput{
		Address: "成都市武侯区天府三街", BusinessHours: "19:30-次日02:30",
		Latitude: 30.657, Longitude: 104.066,
	}
	if err := validateVenueProfileInput(valid); err != nil {
		t.Fatalf("valid venue profile rejected: %v", err)
	}
	valid.Address = ""
	if err := validateVenueProfileInput(valid); err == nil {
		t.Fatal("venue without address must be rejected")
	}
}
