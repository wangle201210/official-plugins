// settlement_test.go holds the database-gated integration tests for the operator
// settlement service: the dashboard aggregates, the batch certificate issuance
// (cohort selection, idempotency and the rejection paths), the shared-device risk
// view and the settlement archive create/list round-trip. The tests are skipped
// unless LINA_TEST_PGSQL_LINK is set.

package settlement

import (
	"context"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

// TestDashboardAggregates verifies every dashboard figure is counted on the
// database side across the C1-C5 tables.
func TestDashboardAggregates(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New(nil, Config{})

	user1 := insertUserRow(t, ctx, "玩家A", "student", "fp-1")
	user2 := insertUserRow(t, ctx, "玩家B", "friend", "fp-2")
	activeNiu := insertNiuRow(t, ctx, "活跃牛", "active")
	insertNiuRow(t, ctx, "未激活牛", "inactive")
	insertActivationRow(t, ctx, user1, activeNiu, 1, 1)
	insertActivationRow(t, ctx, user2, activeNiu, 0, 2)
	insertFeedingRow(t, ctx, user1, 100)
	insertFeedingRow(t, ctx, user2, 50)
	insertStealRow(t, ctx, user1, user2, 1)
	insertGiftRow(t, ctx, user1, user2, 1)
	insertCheckinRow(t, ctx, user1, 1)
	insertCheckinRow(t, ctx, user1, 2)
	certHonor := insertHonorDefRow(t, ctx, "certificate", "participation", 0)
	if _, err := dao.UserHonor.Ctx(ctx).Data(do.UserHonor{
		UserId:  user1,
		HonorId: certHonor,
	}).Insert(); err != nil {
		t.Fatalf("seed user honor failed: %v", err)
	}

	dashboard, err := svc.Dashboard(ctx)
	if err != nil {
		t.Fatalf("Dashboard returned error: %v", err)
	}
	checks := map[string]struct{ got, want int64 }{
		"playerCount":             {dashboard.PlayerCount, 2},
		"activatedNiuCount":       {dashboard.ActivatedNiuCount, 1},
		"totalNiuCount":           {dashboard.TotalNiuCount, 2},
		"firstActivatorCount":     {dashboard.FirstActivatorCount, 1},
		"feedingCount":            {dashboard.FeedingCount, 2},
		"feedTotalEffect":         {dashboard.FeedTotalEffect, 150},
		"stealCount":              {dashboard.StealCount, 1},
		"giftCount":               {dashboard.GiftCount, 1},
		"checkinCount":            {dashboard.CheckinCount, 2},
		"certificateGrantedCount": {dashboard.CertificateGrantedCount, 1},
	}
	for name, c := range checks {
		if c.got != c.want {
			t.Fatalf("dashboard.%s = %d, want %d", name, c.got, c.want)
		}
	}
}

// TestIssueParticipationIdempotent verifies a participation certificate is issued
// to all players and a re-run grants nobody new.
func TestIssueParticipationIdempotent(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New(nil, Config{})

	insertUserRow(t, ctx, "A", "student", "")
	insertUserRow(t, ctx, "B", "student", "")
	insertUserRow(t, ctx, "C", "friend", "")
	cert := insertHonorDefRow(t, ctx, "certificate", "participation", 0)

	first, err := svc.IssueCertificates(ctx, cert)
	if err != nil {
		t.Fatalf("first issue returned error: %v", err)
	}
	if first.Eligible != 3 || first.Issued != 3 || first.Skipped != 0 {
		t.Fatalf("first issue = %+v, want eligible=3 issued=3 skipped=0", first)
	}
	if got := countUserHonors(t, ctx, cert); got != 3 {
		t.Fatalf("granted rows after first issue = %d, want 3", got)
	}

	second, err := svc.IssueCertificates(ctx, cert)
	if err != nil {
		t.Fatalf("second issue returned error: %v", err)
	}
	if second.Eligible != 3 || second.Issued != 0 || second.Skipped != 3 {
		t.Fatalf("second issue = %+v, want eligible=3 issued=0 skipped=3", second)
	}
	if got := countUserHonors(t, ctx, cert); got != 3 {
		t.Fatalf("granted rows after second issue = %d, want 3 (idempotent)", got)
	}
}

// TestIssueFeedCountCohort verifies a feed_count certificate is issued only to the
// players whose feeding count meets the threshold.
func TestIssueFeedCountCohort(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New(nil, Config{})

	heavy := insertUserRow(t, ctx, "Heavy", "student", "")
	light := insertUserRow(t, ctx, "Light", "student", "")
	insertUserRow(t, ctx, "Idle", "student", "")
	// heavy feeds twice (meets threshold 2), light once (below), idle never.
	insertFeedingRow(t, ctx, heavy, 10)
	insertFeedingRow(t, ctx, heavy, 10)
	insertFeedingRow(t, ctx, light, 10)
	cert := insertHonorDefRow(t, ctx, "certificate", "feed_count", 2)

	result, err := svc.IssueCertificates(ctx, cert)
	if err != nil {
		t.Fatalf("issue returned error: %v", err)
	}
	if result.Eligible != 1 || result.Issued != 1 {
		t.Fatalf("feed_count issue = %+v, want eligible=1 issued=1", result)
	}
	if got := countUserHonors(t, ctx, cert); got != 1 {
		t.Fatalf("granted rows = %d, want 1", got)
	}
}

// TestIssueRejectsNonCertificate verifies a non-certificate honor is rejected.
func TestIssueRejectsNonCertificate(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New(nil, Config{})

	insertUserRow(t, ctx, "A", "student", "")
	badge := insertHonorDefRow(t, ctx, "badge", "participation", 0)

	_, err := svc.IssueCertificates(ctx, badge)
	assertBizCode(t, err, "PLUGIN_SICAU_NIU_SETTLEMENT_NOT_CERTIFICATE")
	if got := countUserHonors(t, ctx, badge); got != 0 {
		t.Fatalf("granted rows = %d, want 0 after rejection", got)
	}
}

// TestIssueRejectsCollectionUnlock verifies a collection-based certificate is
// rejected from batch settlement.
func TestIssueRejectsCollectionUnlock(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New(nil, Config{})

	insertUserRow(t, ctx, "A", "student", "")
	cert := insertHonorDefRow(t, ctx, "certificate", "full_complete", 0)

	_, err := svc.IssueCertificates(ctx, cert)
	assertBizCode(t, err, "PLUGIN_SICAU_NIU_SETTLEMENT_UNLOCK_UNSUPPORTED")
}

// TestIssueHonorNotFound verifies a missing honor is rejected.
func TestIssueHonorNotFound(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New(nil, Config{})

	_, err := svc.IssueCertificates(ctx, 99999)
	assertBizCode(t, err, "PLUGIN_SICAU_NIU_SETTLEMENT_HONOR_NOT_FOUND")
}

// TestRiskDeviceClusters verifies the risk view surfaces only shared-device
// clusters with their members.
func TestRiskDeviceClusters(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New(nil, Config{})

	// Two players share fp-x; one player is alone on fp-y; one has no fingerprint.
	insertUserRow(t, ctx, "Dup1", "student", "fp-x")
	insertUserRow(t, ctx, "Dup2", "student", "fp-x")
	insertUserRow(t, ctx, "Solo", "student", "fp-y")
	insertUserRow(t, ctx, "Blank", "student", "")

	clusters, err := svc.RiskDeviceClusters(ctx)
	if err != nil {
		t.Fatalf("RiskDeviceClusters returned error: %v", err)
	}
	if len(clusters.List) != 1 {
		t.Fatalf("expected 1 shared-device cluster, got %d", len(clusters.List))
	}
	cluster := clusters.List[0]
	if cluster.Fingerprint != "fp-x" || cluster.Count != 2 || len(cluster.Members) != 2 {
		t.Fatalf("unexpected cluster: fp=%q count=%d members=%d", cluster.Fingerprint, cluster.Count, len(cluster.Members))
	}
}

// TestArchiveCreateAndList verifies an archive freezes the dashboard snapshot and
// is returned by the list.
func TestArchiveCreateAndList(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New(nil, Config{})

	insertUserRow(t, ctx, "A", "student", "")
	insertUserRow(t, ctx, "B", "student", "")

	id, err := svc.CreateArchive(ctx, "结算公示 2026")
	if err != nil {
		t.Fatalf("CreateArchive returned error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("CreateArchive returned non-positive id %d", id)
	}

	archives, err := svc.ListArchives(ctx)
	if err != nil {
		t.Fatalf("ListArchives returned error: %v", err)
	}
	if len(archives.List) != 1 {
		t.Fatalf("expected 1 archive, got %d", len(archives.List))
	}
	archive := archives.List[0]
	if archive.Title != "结算公示 2026" {
		t.Fatalf("unexpected archive title: %q", archive.Title)
	}
	if !strings.Contains(archive.Snapshot, "\"playerCount\":2") {
		t.Fatalf("snapshot does not carry frozen playerCount: %q", archive.Snapshot)
	}
	if archive.ArchivedAt == nil || *archive.ArchivedAt <= 0 {
		t.Fatalf("archive archivedAt not populated: %v", archive.ArchivedAt)
	}
}

// assertBizCode fails the test unless err is a structured business error whose
// runtime code equals wantCode.
func assertBizCode(t *testing.T, err error, wantCode string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected business error with code %s, got nil", wantCode)
	}
	bizErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error with code %s, got %T: %v", wantCode, err, err)
	}
	if bizErr.RuntimeCode() != wantCode {
		t.Fatalf("expected code %s, got %s", wantCode, bizErr.RuntimeCode())
	}
}
