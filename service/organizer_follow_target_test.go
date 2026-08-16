package service

import (
	"Hyper/models"
	"context"
	"testing"
)

func TestResolveOrganizerFollowTargetForCanonicalVenue(t *testing.T) {
	target, err := resolveOrganizerFollowTarget(context.Background(), nil, models.Organizer{ID: 1, Type: models.OrganizerTypeVenue})
	if err != nil {
		t.Fatal(err)
	}
	if target != models.ContentFollowTargetVenue {
		t.Fatalf("expected venue target, got %s", target)
	}
}
