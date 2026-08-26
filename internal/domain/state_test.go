package domain

import (
	"testing"
	"time"
)

func TestVersionLifecycle(t *testing.T) {
	cases := []struct {
		from, to VersionStatus
		ok       bool
	}{
		{VersionDraft, VersionPublished, true}, {VersionPublished, VersionWithdrawn, true},
		{VersionDraft, VersionWithdrawn, false}, {VersionPublished, VersionDraft, false},
	}
	for _, tc := range cases {
		if got := tc.from.CanTransition(tc.to); got != tc.ok {
			t.Fatalf("version transition %s -> %s = %v, want %v", tc.from, tc.to, got, tc.ok)
		}
	}
}

func TestBatchLifecycle(t *testing.T) {
	if !BatchPending.CanTransition(BatchSampling) {
		t.Fatal("pending should enter sampling")
	}
	if !BatchSampling.CanTransition(BatchCleared) {
		t.Fatal("sampling should clear")
	}
	if !BatchSampling.CanTransition(BatchFlagged) {
		t.Fatal("sampling should flag")
	}
	if BatchPending.CanTransition(BatchArchived) {
		t.Fatal("pending must not archive")
	}
	if BatchFlagged.CanTransition(BatchArchived) == false {
		t.Fatal("flagged should archive")
	}
}

func TestCaseAndRemediationTransitions(t *testing.T) {
	if !CaseOpen.CanTransition(CaseInvestigating) || !CaseInvestigating.CanTransition(CaseResolved) || !CaseResolved.CanTransition(CaseClosed) {
		t.Fatal("valid complaint path rejected")
	}
	if CaseOpen.CanTransition(CaseClosed) {
		t.Fatal("complaint may not skip investigation")
	}
	if !RemediationPlanned.CanTransition(RemediationActive) || !RemediationActive.CanTransition(RemediationDone) || !RemediationActive.CanTransition(RemediationEscalated) {
		t.Fatal("valid remediation path rejected")
	}
	if RemediationDone.CanTransition(RemediationActive) {
		t.Fatal("completed remediation reopened")
	}
}

func TestTimeWindowsAndLeaseExpiry(t *testing.T) {
	base := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	if !WithinWindow(base, base) || !WithinWindow(base.Add(-time.Second), base) || WithinWindow(base.Add(time.Second), base) {
		t.Fatal("deadline semantics incorrect")
	}
	if !LeaseExpired(base, base) || !LeaseExpired(base, base.Add(-time.Second)) || LeaseExpired(base, base.Add(time.Second)) {
		t.Fatal("lease semantics incorrect")
	}
}

func TestRequireTransition(t *testing.T) {
	if RequireTransition("a", "b", true) != nil {
		t.Fatal("allowed transition failed")
	}
	if RequireTransition("a", "b", false) == nil {
		t.Fatal("invalid transition accepted")
	}
}

func TestRoleValuesAreStable(t *testing.T) {
	if RoleReviewer == RoleMerchant {
		t.Fatal("roles must differ")
	}
	if string(RoleReviewer) != "reviewer" || string(RoleMerchant) != "merchant" {
		t.Fatal("role wire values changed")
	}
}

func TestStatusValuesAreStable(t *testing.T) {
	values := []string{string(VersionDraft), string(VersionPublished), string(BatchPending), string(BatchSampling), string(BatchCleared), string(BatchFlagged), string(CaseOpen), string(CaseInvestigating), string(CaseResolved), string(CaseClosed)}
	seen := map[string]bool{}
	for _, v := range values {
		if v == "" {
			t.Fatal("empty status")
		}
		if seen[v] {
			t.Fatalf("duplicate status %s", v)
		}
		seen[v] = true
	}
}
