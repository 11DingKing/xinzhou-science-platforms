package policy

import (
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"testing"
	"time"
)

func TestDeclarationValidation(t *testing.T) {
	now := time.Now()
	valid := Declaration{SKU: "s", Channel: "online", Region: "county", DeclaredAt: now, ExpiresAt: now.Add(time.Hour)}
	if err := valid.Validate(now); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []Declaration{{}, {SKU: "s", Channel: "online", Region: "county", DeclaredAt: now, ExpiresAt: now.Add(-time.Second)}, {SKU: "s", Channel: "online", Region: "county", DeclaredAt: now, ExpiresAt: now.Add(-time.Hour)}} {
		if bad.Validate(now) == nil {
			t.Fatalf("invalid declaration accepted: %+v", bad)
		}
	}
}
func TestCompareDetectsDisclosureDifferences(t *testing.T) {
	d := Declaration{SKU: "s", Channel: "online", Region: "city", Material: "steel", Configuration: "full"}
	s := Sample{SKU: "s", Channel: "online", Region: "county", Material: "aluminum", Configuration: "lite", Score: 55}
	got := Compare(d, s)
	if got.Allowed || got.Severity != "critical" || len(got.Reasons) != 4 {
		t.Fatalf("decision=%+v", got)
	}
	if len(MergeReasons(got.Reasons, []string{"region differs", "new"})) != 5 {
		t.Fatal("reasons not merged")
	}
}
func TestPolicyHelpers(t *testing.T) {
	if NormalizeRegion("  Hu  Bei ") == "hubei" {
		t.Fatal("spaces should remain semantic")
	}
	if NormalizeRegion("  Hu  Bei ") != "hu bei" {
		t.Fatal("normalize failed")
	}
	for score, want := range map[int]string{10: "critical", 50: "high", 70: "medium", 99: "low"} {
		if EscalationLevel(score) != want {
			t.Fatalf("score %d", score)
		}
	}
	if RequirePublished(domain.ProductVersion{Status: domain.VersionDraft}) == nil {
		t.Fatal("draft accepted")
	}
	if RequireBatchReviewable(domain.Batch{Status: domain.BatchPending}, time.Now()) == nil {
		t.Fatal("pending batch accepted")
	}
}
