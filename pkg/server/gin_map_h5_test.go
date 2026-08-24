package server

import "testing"

func TestMapH5DirUsesEnvironmentOverride(t *testing.T) {
	t.Setenv("MAP_H5_DIR", "/tmp/hyper-map-h5")
	if got := mapH5Dir(); got != "/tmp/hyper-map-h5" {
		t.Fatalf("mapH5Dir() = %q, want environment override", got)
	}
}
