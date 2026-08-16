package handler

import (
	"mime/multipart"
	"strings"
	"testing"
)

func TestBuildUploadKeyAcceptsVenueImages(t *testing.T) {
	header := &multipart.FileHeader{Filename: "venue.png"}
	for _, uploadType := range []string{"venue_cover", "venue_gallery"} {
		key, err := buildUploadKey(uploadType, header)
		if err != nil {
			t.Fatalf("%s should be accepted: %v", uploadType, err)
		}
		if !strings.HasPrefix(key, "ticketing/"+uploadType+"/") {
			t.Fatalf("unexpected key for %s: %s", uploadType, key)
		}
	}
}
