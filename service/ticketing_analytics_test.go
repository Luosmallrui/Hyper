package service

import (
	"Hyper/models"
	"strings"
	"testing"
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
