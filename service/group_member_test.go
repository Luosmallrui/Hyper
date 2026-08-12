package service

import (
	"Hyper/models"
	"testing"
)

func TestGroupMemberPermissions(t *testing.T) {
	tests := []struct {
		name       string
		role       int
		wantInvite bool
		wantSetAdm bool
	}{
		{name: "owner", role: models.GroupMemberLeaderOwner, wantInvite: true, wantSetAdm: true},
		{name: "admin", role: models.GroupMemberLeaderAdmin, wantInvite: true, wantSetAdm: false},
		{name: "member", role: models.GroupMemberLeaderOrdinary, wantInvite: false, wantSetAdm: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			permissions := groupMemberPermissions(test.role)
			if permissions.CanInvite != test.wantInvite || permissions.CanSetAdmin != test.wantSetAdm {
				t.Fatalf("permissions = %+v", permissions)
			}
		})
	}
}

func TestGroupMemberCanManageAndSend(t *testing.T) {
	if !canManageGroupMember(models.GroupMemberLeaderOwner, models.GroupMemberLeaderAdmin, false) {
		t.Fatal("owner must be able to manage an admin")
	}
	if canManageGroupMember(models.GroupMemberLeaderAdmin, models.GroupMemberLeaderAdmin, false) {
		t.Fatal("admin must not manage another admin")
	}
	if canManageGroupMember(models.GroupMemberLeaderAdmin, models.GroupMemberLeaderOrdinary, true) {
		t.Fatal("member must not manage themselves")
	}
	if groupMemberCanSend(models.GroupMemberLeaderOrdinary, true, false) {
		t.Fatal("personally muted ordinary member must not send")
	}
	if groupMemberCanSend(models.GroupMemberLeaderOrdinary, false, true) {
		t.Fatal("ordinary member must not send when all members are muted")
	}
	if !groupMemberCanSend(models.GroupMemberLeaderAdmin, false, true) {
		t.Fatal("admin must send while all members are muted")
	}
}
