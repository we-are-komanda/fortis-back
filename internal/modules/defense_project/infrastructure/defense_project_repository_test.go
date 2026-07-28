package infrastructure

import "testing"

func TestNullableEnterpriseID(t *testing.T) {
	if got := nullableEnterpriseID(""); got != nil {
		t.Fatalf("expected empty enterprise ID to become nil, got %v", got)
	}

	const id = "1d864f98-a1b7-4ce5-babd-c5fc22a74fb4"
	if got := nullableEnterpriseID(id); got != id {
		t.Fatalf("expected non-empty enterprise ID to be preserved, got %v", got)
	}
}
