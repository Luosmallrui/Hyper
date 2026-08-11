package dao

import (
	"Hyper/models"
	"testing"
)

func TestIsContentFollowTarget(t *testing.T) {
	for _, targetType := range []string{
		models.ContentFollowTargetActivity,
		models.ContentFollowTargetVenue,
		models.ContentFollowTargetOrganizer,
		models.ContentFollowTargetParty,
	} {
		if !isContentFollowTarget(targetType) {
			t.Fatalf("target type %q should be supported", targetType)
		}
	}
	if isContentFollowTarget("unknown") {
		t.Fatal("unknown target type should not be supported")
	}
}
