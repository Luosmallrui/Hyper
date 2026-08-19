package service

import (
	"Hyper/types"
	"testing"
)

func TestExplicitActivityLocationValidation(t *testing.T) {
	address := "成都市武侯区音乐公园"
	latitude, longitude := 30.6352, 104.0431
	req := types.ActivityCreateRequest{Address: &address, Latitude: &latitude, Longitude: &longitude}
	if !hasExplicitActivityLocation(req) {
		t.Fatal("location should be explicit")
	}
	if err := validateExplicitActivityLocation(req); err != nil {
		t.Fatalf("valid custom activity location rejected: %v", err)
	}
	req.Longitude = nil
	if err := validateExplicitActivityLocation(req); err == nil {
		t.Fatal("partial custom location must be rejected")
	}
}

func TestMissingActivityLocationUsesVenueDefault(t *testing.T) {
	if hasExplicitActivityLocation(types.ActivityCreateRequest{}) {
		t.Fatal("no location fields should use the venue default")
	}
}
