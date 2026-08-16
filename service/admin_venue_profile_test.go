package service

import (
	"Hyper/models"
	"Hyper/types"
	"testing"
)

func TestFillAdminVenueProfileFromLegacy(t *testing.T) {
	profile := &types.OrganizerVenueProfileInput{Gallery: []string{}}
	legacy := models.Activity{
		PosterList: "list.png", PosterDetail: "detail.png", PosterLong: "long.png",
		Description: "旧场地介绍", Address: "成都市武侯区", Latitude: 30.6, Longitude: 104.0,
	}
	fillAdminVenueProfileFromLegacy(profile, legacy)
	if profile.CoverImage != "list.png" || profile.Description != "旧场地介绍" || profile.Address != "成都市武侯区" {
		t.Fatalf("legacy profile not filled: %+v", profile)
	}
	if len(profile.Gallery) != 3 {
		t.Fatalf("expected legacy posters in gallery, got %+v", profile.Gallery)
	}
}

func TestFillAdminVenueProfileFromLegacyKeepsCurrentFields(t *testing.T) {
	profile := &types.OrganizerVenueProfileInput{CoverImage: "current.png", Gallery: []string{"current-gallery.png"}}
	fillAdminVenueProfileFromLegacy(profile, models.Activity{PosterList: "legacy.png", PosterDetail: "detail.png"})
	if profile.CoverImage != "current.png" || len(profile.Gallery) != 1 || profile.Gallery[0] != "current-gallery.png" {
		t.Fatalf("current profile should win: %+v", profile)
	}
}
