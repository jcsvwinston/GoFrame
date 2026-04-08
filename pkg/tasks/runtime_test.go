package tasks

import "testing"

func TestInspectRuntime_WithoutRedisURL(t *testing.T) {
	snapshot := InspectRuntime("")
	if snapshot.Enabled {
		t.Fatalf("expected disabled snapshot when redis_url is empty")
	}
	if snapshot.Reason == "" {
		t.Fatalf("expected reason for disabled snapshot")
	}
}

func TestInspectRuntime_InvalidRedisURL(t *testing.T) {
	snapshot := InspectRuntime("http://localhost:6379")
	if snapshot.Enabled {
		t.Fatalf("expected disabled snapshot for invalid redis scheme")
	}
	if snapshot.Reason == "" {
		t.Fatalf("expected reason for invalid redis scheme")
	}
}
