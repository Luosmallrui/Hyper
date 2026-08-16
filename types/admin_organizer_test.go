package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAdminOrganizerItemIncludesProfileRevisionStatus(t *testing.T) {
	body, err := json.Marshal(AdminOrganizerItem{
		AuditKind: "profile_revision", PendingProfileStatus: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"pending_profile_status":1`) {
		t.Fatalf("missing pending profile status: %s", body)
	}
}

func TestAdminOrganizerDetailIncludesBothVenueProfiles(t *testing.T) {
	body, err := json.Marshal(AdminOrganizerDetail{
		FollowerCount: 3,
		VenueProfile:  &OrganizerVenueProfileInput{Address: "旧地址"},
		PendingProfileRevision: &OrganizerVenueProfileRevision{
			OrganizerVenueProfileInput: OrganizerVenueProfileInput{Address: "新地址"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"follower_count":3`, `"venue_profile"`, `"pending_profile_revision"`, "旧地址", "新地址"} {
		if !strings.Contains(string(body), field) {
			t.Fatalf("missing %s: %s", field, body)
		}
	}
}
