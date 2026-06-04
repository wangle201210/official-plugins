// honor_certificate_test.go holds the database-gated integration tests for the
// player electronic-certificate generation: a holder gets a non-empty PNG, distinct
// holders get distinct certificates, and the type / ownership checks reject. The
// tests are skipped unless LINA_TEST_PGSQL_LINK is set.

package honor

import (
	"context"
	"testing"

	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

// seedCertificateHonor inserts a certificate honor with a unique code and returns
// its ID.
func seedCertificateHonor(t *testing.T, ctx context.Context, code string) int64 {
	t.Helper()
	id, err := dao.HonorDef.Ctx(ctx).Data(do.HonorDef{
		HonorType:  string(HonorTypeCertificate),
		Code:       code,
		Name:       "川农120纪念证书",
		UnlockType: string(UnlockTypeParticipation),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("seed certificate honor failed: %v", err)
	}
	return id
}

// grantHonor inserts a user_honor grant row.
func grantHonor(t *testing.T, ctx context.Context, userID, honorID int64) {
	t.Helper()
	if _, err := dao.UserHonor.Ctx(ctx).Data(do.UserHonor{UserId: userID, HonorId: honorID}).Insert(); err != nil {
		t.Fatalf("grant honor failed: %v", err)
	}
}

// TestPlayerCertificateRendersForHolder verifies a holder receives a non-empty PNG
// and that two distinct holders receive distinct certificates.
func TestPlayerCertificateRendersForHolder(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)
	svc := New(NewBasicCertRenderer(), Config{CampusBadge: "川农120周年"})

	cert := seedCertificateHonor(t, ctx, "cert-a")
	userA := insertUserRow(t, ctx, "证书同学A")
	userB := insertUserRow(t, ctx, "证书同学B")
	grantHonor(t, ctx, userA, cert)
	grantHonor(t, ctx, userB, cert)

	certA, err := svc.PlayerCertificate(ctx, userA, cert)
	if err != nil {
		t.Fatalf("PlayerCertificate(A) error: %v", err)
	}
	if certA.ImageBase64 == "" {
		t.Fatalf("certificate image is empty")
	}
	if certA.HonorName != "川农120纪念证书" || certA.Nickname != "证书同学A" {
		t.Fatalf("unexpected certificate fields: name=%q nickname=%q", certA.HonorName, certA.Nickname)
	}

	certB, err := svc.PlayerCertificate(ctx, userB, cert)
	if err != nil {
		t.Fatalf("PlayerCertificate(B) error: %v", err)
	}
	if certA.ImageBase64 == certB.ImageBase64 {
		t.Fatalf("distinct holders produced identical certificates")
	}
}

// TestPlayerCertificateRejectsNotOwned verifies a player without the grant is
// rejected.
func TestPlayerCertificateRejectsNotOwned(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)
	svc := New(NewBasicCertRenderer(), Config{})

	cert := seedCertificateHonor(t, ctx, "cert-b")
	user := insertUserRow(t, ctx, "未获证书的玩家")

	_, err := svc.PlayerCertificate(ctx, user, cert)
	assertBizCode(t, err, "PLUGIN_SICAU_NIU_CERTIFICATE_NOT_OWNED")
}

// TestPlayerCertificateRejectsNonCertificate verifies a non-certificate honor is
// rejected even when held.
func TestPlayerCertificateRejectsNonCertificate(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)
	svc := New(NewBasicCertRenderer(), Config{})

	badgeID, err := dao.HonorDef.Ctx(ctx).Data(do.HonorDef{
		HonorType:  string(HonorTypeBadge),
		Code:       "badge-x",
		Name:       "参与徽章",
		UnlockType: string(UnlockTypeParticipation),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("seed badge failed: %v", err)
	}
	user := insertUserRow(t, ctx, "持徽章的玩家")
	grantHonor(t, ctx, user, badgeID)

	_, err = svc.PlayerCertificate(ctx, user, badgeID)
	assertBizCode(t, err, "PLUGIN_SICAU_NIU_HONOR_NOT_CERTIFICATE")
}
