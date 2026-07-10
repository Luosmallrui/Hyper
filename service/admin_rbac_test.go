package service

import "testing"

func TestRequiredAdminPermission(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   string
	}{
		{"GET", "/v1/admin/dashboard", "admin.dashboard"},
		{"GET", "/v1/admin/orders", "admin.orders"},
		{"POST", "/v1/admin/orders/T1/refund/approve", "admin.orders"},
		{"GET", "/v1/admin/finance/platform-flows", "admin.finance"},
		{"GET", "/v1/admin/notes", "admin.content"},
		{"POST", "/v1/admin/admins", "*"},
		{"PUT", "/v1/admin/roles/1", "*"},
	}
	for _, test := range tests {
		if got := RequiredAdminPermission(test.method, test.path); got != test.want {
			t.Fatalf("%s %s: got %q, want %q", test.method, test.path, got, test.want)
		}
	}
}

func TestNormalizeAdminPermissions(t *testing.T) {
	permissions, err := normalizeAdminPermissions(`["admin.orders", "admin.orders", "admin.finance"]`)
	if err != nil || len(permissions) != 2 {
		t.Fatalf("expected normalized permissions, got %#v, err=%v", permissions, err)
	}
	if _, err := normalizeAdminPermissions(`["admin.unknown"]`); err == nil {
		t.Fatal("expected unknown permission to be rejected")
	}
	if _, err := normalizeAdminPermissions(`["*", "admin.orders"]`); err == nil {
		t.Fatal("expected mixed super-admin permissions to be rejected")
	}
}
