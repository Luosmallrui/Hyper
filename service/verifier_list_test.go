package service

import (
	"context"
	"testing"
)

func TestListVerifiedRejectsMissingVerifierID(t *testing.T) {
	service := &TicketingService{}
	if _, err := service.ListVerified(context.Background(), 0, 1, 20); err == nil {
		t.Fatal("expected missing verifier ID to be rejected")
	}
}

func TestListVerifiedByUserRejectsAnonymousUser(t *testing.T) {
	service := &TicketingService{}
	if _, err := service.ListVerifiedByUser(context.Background(), 0, 1, 20); err == nil {
		t.Fatal("expected anonymous user to be rejected")
	}
}
