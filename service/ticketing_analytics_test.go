package service

import (
	"Hyper/models"
	"Hyper/types"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestActivityVisitorKey(t *testing.T) {
	if got := activityVisitorKey(42, "ignored-for-logged-in-user"); got != "u:42" {
		t.Fatalf("logged-in visitor key = %q, want u:42", got)
	}
	first := activityVisitorKey(0, "mini-program-device-id")
	second := activityVisitorKey(0, "mini-program-device-id")
	if first == "" || first != second || !strings.HasPrefix(first, "g:") {
		t.Fatalf("guest visitor key must be stable hashed value, got %q and %q", first, second)
	}
	if strings.Contains(first, "mini-program-device-id") {
		t.Fatalf("guest visitor key must not persist the raw identifier: %q", first)
	}
	if got := activityVisitorKey(0, ""); got != "" {
		t.Fatalf("empty guest visitor key = %q, want empty", got)
	}
}

func TestWithdrawStatusForOrder(t *testing.T) {
	tests := []struct {
		name    string
		status  int8
		amounts orderWithdrawAmounts
		want    string
	}{
		{name: "pending allocation", status: models.TicketOrderStatusUsable, amounts: orderWithdrawAmounts{PendingAmount: 100}, want: "pending_withdraw"},
		{name: "settled allocation", status: models.TicketOrderStatusUsed, amounts: orderWithdrawAmounts{SettledAmount: 100}, want: "withdrawn"},
		{name: "available paid order", status: models.TicketOrderStatusRefundReject, want: "available"},
		{name: "refunded order", status: models.TicketOrderStatusRefundSuccess, want: "unavailable"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := withdrawStatusForOrder(test.status, test.amounts); got != test.want {
				t.Fatalf("withdrawStatusForOrder() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeSalesChannel(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		defaultWeChat bool
		want          string
		wantErr       bool
	}{
		{name: "new order defaults to wechat", defaultWeChat: true, want: salesChannelWeChat},
		{name: "empty filter", defaultWeChat: false, want: ""},
		{name: "wechat mini program alias", value: "wechat_mini_program", want: salesChannelWeChat},
		{name: "douyin mini program alias", value: "douyin_mini_program", want: salesChannelDouyin},
		{name: "web", value: "web", want: salesChannelWeb},
		{name: "unknown", value: "kuaishou", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeSalesChannel(test.value, test.defaultWeChat)
			if (err != nil) != test.wantErr {
				t.Fatalf("NormalizeSalesChannel() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("NormalizeSalesChannel() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestActivityUpdatesUsesTypeSpecificTimeFields(t *testing.T) {
	venueType := models.ActivityTypeVenue
	invalidTime := "not-a-datetime"
	updates, err := activityUpdates(types.ActivityCreateRequest{
		Type:      &venueType,
		StartTime: &invalidTime,
		EndTime:   &invalidTime,
	}, models.ActivityTypeVenue)
	if err != nil {
		t.Fatalf("venue must ignore activity time fields, got error: %v", err)
	}
	if _, ok := updates["start_time"]; ok {
		t.Fatal("venue update must not write start_time")
	}
	if _, ok := updates["end_time"]; ok {
		t.Fatal("venue update must not write end_time")
	}

	partyType := models.ActivityTypeParty
	_, err = activityUpdates(types.ActivityCreateRequest{
		Type:      &partyType,
		StartTime: &invalidTime,
	}, models.ActivityTypeParty)
	if err == nil {
		t.Fatal("party must validate start_time")
	}
}

func TestActivityEditNeedsReaudit(t *testing.T) {
	if activityEditNeedsReaudit(types.ActivityCreateRequest{}, map[string]any{}) {
		t.Fatal("empty edit must not trigger a re-audit")
	}
	if !activityEditNeedsReaudit(types.ActivityCreateRequest{}, map[string]any{"name": "new name"}) {
		t.Fatal("content update must trigger a re-audit")
	}
	if !activityEditNeedsReaudit(types.ActivityCreateRequest{TagIDs: []int64{}}, map[string]any{}) {
		t.Fatal("tag replacement, including clearing tags, must trigger a re-audit")
	}
	if !activityEditNeedsReaudit(types.ActivityCreateRequest{TicketSpecs: []types.TicketSpecSaveItem{}}, map[string]any{}) {
		t.Fatal("ticket specification save must trigger a re-audit")
	}
}

func TestNormalizeActivityAuditType(t *testing.T) {
	if got := normalizeActivityAuditType(models.ActivityAuditTypeReaudit); got != models.ActivityAuditTypeReaudit {
		t.Fatalf("re_audit type = %q", got)
	}
	if got := normalizeActivityAuditType("unexpected"); got != models.ActivityAuditTypeInitial {
		t.Fatalf("unknown type = %q, want initial", got)
	}
}

func TestActivityDetailResponseIncludesVenueBusinessHours(t *testing.T) {
	response := types.ActivityDetailResponse{BusinessHours: "19:30-02:30"}
	if response.BusinessHours != "19:30-02:30" {
		t.Fatalf("business_hours = %q", response.BusinessHours)
	}
}

func TestTicketSpecUpdatesKeepsZeroValues(t *testing.T) {
	updates := ticketSpecUpdates(&models.TicketSpec{IsEnabled: 0, Price: 0, Stock: 0})
	for _, key := range []string{"is_enabled", "price", "stock"} {
		if _, ok := updates[key]; !ok {
			t.Fatalf("ticket update must include %s even when it is zero", key)
		}
	}
}

func TestTicketSpecIDUsesJSONString(t *testing.T) {
	const unsafeID int64 = 9007199254740993
	encoded, err := json.Marshal(models.TicketSpec{ID: unsafeID})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"id":"9007199254740993"`) {
		t.Fatalf("ticket spec id must be serialized as a string, got %s", encoded)
	}

	var input types.TicketSpecSaveItem
	if err := json.Unmarshal([]byte(`{"id":"9007199254740993","name":"test","price":1}`), &input); err != nil {
		t.Fatalf("string ticket spec id must be accepted: %v", err)
	}
	if input.ID != unsafeID {
		t.Fatalf("ticket spec id = %d, want %d", input.ID, unsafeID)
	}
}

func TestVenueValidityWindow(t *testing.T) {
	now := time.Date(2026, time.August, 11, 10, 0, 0, 0, time.Local)
	start, end := venueValidityWindow(now)
	if !start.Equal(now) {
		t.Fatalf("venue start = %v, want %v", start, now)
	}
	wantEnd := now.AddDate(20, 0, 0)
	if !end.Equal(wantEnd) {
		t.Fatalf("venue end = %v, want %v", end, wantEnd)
	}
}

func TestIsActivityPublic(t *testing.T) {
	tests := []struct {
		name     string
		activity models.Activity
		want     bool
	}{
		{name: "online and visible", activity: models.Activity{Status: models.ActivityStatusOnline}, want: true},
		{name: "online but hidden", activity: models.Activity{Status: models.ActivityStatusOnline, IsHidden: 1}, want: false},
		{name: "pending", activity: models.Activity{Status: models.ActivityStatusPending}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isActivityPublic(test.activity); got != test.want {
				t.Fatalf("isActivityPublic() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestUnavailableActivityDetail(t *testing.T) {
	resp := unavailableActivityDetail(8)
	if resp.ID != 8 {
		t.Fatalf("activity id = %d, want 8", resp.ID)
	}
	if resp.Name != unavailableActivityName || resp.IsHidden != 1 || resp.HiddenReason != unavailableActivityName {
		t.Fatalf("unexpected unavailable activity response: %+v", resp.Activity)
	}
	if resp.TicketSpecs == nil || resp.Tags == nil {
		t.Fatal("unavailable activity must return empty collections instead of null")
	}
}

func TestApplyOrderActivityListItemForMissingActivity(t *testing.T) {
	var item types.TicketOrderListItem
	applyOrderActivityListItem(&item, 13, 0, "", time.Time{}, time.Time{}, "", 0, "")
	if item.Activity.ID != 13 || item.Activity.Name != unavailableActivityName || !item.Activity.IsHidden || item.Activity.HiddenReason != unavailableActivityName {
		t.Fatalf("unexpected missing activity list item: %+v", item.Activity)
	}
}
