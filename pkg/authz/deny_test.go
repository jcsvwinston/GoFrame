package authz

import (
	"testing"
)

// TestEnforcer_DefaultDeny_NoPolicyMatchesDenies pins the baseline
// expectation: a fresh enforcer with no policies denies every request.
// This was already true before the deny-override migration, but is
// load-bearing for the new model — a regression here would mean the
// `some(where (p.eft == allow))` half of the effect formula stopped
// firing.
func TestEnforcer_DefaultDeny_NoPolicyMatchesDenies(t *testing.T) {
	e := newTestEnforcer(t)
	if e.Can("anyone", "/anything", "read") {
		t.Fatal("default-deny must hold when no policy matches")
	}
}

// TestEnforcer_Deny_OverridesAllow exercises the core new behaviour:
// an explicit deny rule on (sub, obj, act) blocks the request even when
// a matching allow rule exists. Operators rely on this to block a
// specific user without revoking the broader role's allow.
func TestEnforcer_Deny_OverridesAllow(t *testing.T) {
	e := newTestEnforcer(t)
	if err := e.AddPolicy("admin", "/api/*", "*"); err != nil {
		t.Fatalf("AddPolicy: %v", err)
	}
	if err := e.AddRole("alice", "admin"); err != nil {
		t.Fatalf("AddRole: %v", err)
	}
	if !e.Can("alice", "/api/users/1", "delete") {
		t.Fatal("alice (admin) should be allowed before any deny rule")
	}

	if err := e.Deny("alice", "/api/users/1", "delete"); err != nil {
		t.Fatalf("Deny: %v", err)
	}
	if e.Can("alice", "/api/users/1", "delete") {
		t.Fatal("deny rule on alice must override the admin role's allow")
	}

	// Other actions for alice still allowed — the deny is scoped.
	if !e.Can("alice", "/api/users/1", "read") {
		t.Fatal("scoped deny must not affect unrelated actions")
	}

	// Other users with admin role unaffected.
	if err := e.AddRole("bob", "admin"); err != nil {
		t.Fatalf("AddRole: %v", err)
	}
	if !e.Can("bob", "/api/users/1", "delete") {
		t.Fatal("deny on alice must not propagate to bob")
	}
}

// TestEnforcer_Deny_WithoutMatchingAllow denies as well — the deny
// rule does not magically create an allow path; the request still
// fails the "some allow" half of the effect formula.
func TestEnforcer_Deny_WithoutMatchingAllow(t *testing.T) {
	e := newTestEnforcer(t)
	if err := e.Deny("alice", "/api/users/1", "read"); err != nil {
		t.Fatalf("Deny: %v", err)
	}
	if e.Can("alice", "/api/users/1", "read") {
		t.Fatal("deny without any allow still results in deny (no implicit allow)")
	}
}

// TestEnforcer_RemovePolicy_RemovesBothEffects pins the contract of
// RemovePolicy: a single call drops both the allow and the deny
// variants matching (sub, obj, act). Operators call RemovePolicy to
// "stop applying this rule" without having to know how it was added.
func TestEnforcer_RemovePolicy_RemovesBothEffects(t *testing.T) {
	e := newTestEnforcer(t)
	if err := e.AddPolicy("alice", "/data", "read"); err != nil {
		t.Fatalf("AddPolicy: %v", err)
	}
	if err := e.Deny("alice", "/data", "read"); err != nil {
		t.Fatalf("Deny: %v", err)
	}

	// With both present the deny wins.
	if e.Can("alice", "/data", "read") {
		t.Fatal("deny-override should block during this phase")
	}

	if err := e.RemovePolicy("alice", "/data", "read"); err != nil {
		t.Fatalf("RemovePolicy: %v", err)
	}
	// Both gone → default-deny.
	if e.Can("alice", "/data", "read") {
		t.Fatal("after RemovePolicy nothing should remain that grants access")
	}
}

// TestEnforcer_RemovePolicy_OnlyDenyRevealsAllow targets a subtler
// admin scenario: an operator wrote a deny override; later they want
// to lift just the deny while keeping the allow. RemovePolicy removes
// both today (documented behaviour); to lift only the deny they would
// add the allow back. The test fixes that behaviour so the choice is
// intentional.
func TestEnforcer_RemovePolicy_OnlyDenyRevealsAllow(t *testing.T) {
	e := newTestEnforcer(t)
	if err := e.AddPolicy("alice", "/data", "read"); err != nil {
		t.Fatalf("AddPolicy: %v", err)
	}
	if err := e.Deny("alice", "/data", "read"); err != nil {
		t.Fatalf("Deny: %v", err)
	}
	if err := e.RemovePolicy("alice", "/data", "read"); err != nil {
		t.Fatalf("RemovePolicy: %v", err)
	}
	// Both effects removed by a single RemovePolicy call — operator
	// must re-add the allow to grant access again.
	if e.Can("alice", "/data", "read") {
		t.Fatal("RemovePolicy must drop both allow and deny variants")
	}
	if err := e.AddPolicy("alice", "/data", "read"); err != nil {
		t.Fatalf("re-add: %v", err)
	}
	if !e.Can("alice", "/data", "read") {
		t.Fatal("after re-adding the allow alice should be granted")
	}
}
