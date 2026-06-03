// settlement_certificates.go implements the batch certificate issuance. The target
// honor must be a certificate; its unlock rule selects the eligible player cohort
// with a single set-based query (participation = all players, feed_count /
// activation_count = a grouped HAVING count), and collection rules are rejected.
// The already-granted players are excluded with one projected query and the
// remaining grants are written in a single transaction; the user_honor unique index
// makes the write idempotent even under a concurrent re-run.

package settlement

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	honorsvc "lina-plugin-sicau-niu/backend/internal/service/honor"
)

// certificateHonorType and the supported/unsupported unlock-rule strings reuse the
// honor package's stable enum constants so the settlement filter and the honor
// contract never drift.
const (
	certificateHonorType    = string(honorsvc.HonorTypeCertificate)
	unlockParticipation     = string(honorsvc.UnlockTypeParticipation)
	unlockFeedCount         = string(honorsvc.UnlockTypeFeedCount)
	unlockActivationCount   = string(honorsvc.UnlockTypeActivationCount)
)

// IssueResult is the batch certificate issuance result.
type IssueResult struct {
	// Eligible is the number of players who satisfy the certificate's unlock rule.
	Eligible int64
	// Issued is the number of newly granted players in this run.
	Issued int64
	// Skipped is the number of eligible players already granted (skipped).
	Skipped int64
}

// idRow is the temporary projection for one player ID selected by an eligibility
// or grant query.
type idRow struct {
	Id int64 `json:"id"`
}

// IssueCertificates batch-issues the certificate honor to its eligible cohort.
func (s *serviceImpl) IssueCertificates(ctx context.Context, honorID int64) (*IssueResult, error) {
	honor, err := s.loadHonor(ctx, honorID)
	if err != nil {
		return nil, err
	}
	if honor == nil {
		return nil, bizerr.NewCode(CodeSettlementHonorNotFound)
	}
	if honor.HonorType != certificateHonorType {
		return nil, bizerr.NewCode(CodeSettlementNotCertificate)
	}

	eligible, err := s.eligibleUserIDs(ctx, honor)
	if err != nil {
		return nil, err
	}
	if len(eligible) == 0 {
		return &IssueResult{}, nil
	}

	granted, err := s.grantedUserIDs(ctx, honorID, eligible)
	if err != nil {
		return nil, err
	}
	toIssue := make([]int64, 0, len(eligible))
	for _, id := range eligible {
		if !granted[id] {
			toIssue = append(toIssue, id)
		}
	}

	if len(toIssue) > 0 {
		if err = s.insertGrants(ctx, honorID, toIssue); err != nil {
			return nil, err
		}
	}
	return &IssueResult{
		Eligible: int64(len(eligible)),
		Issued:   int64(len(toIssue)),
		Skipped:  int64(len(eligible) - len(toIssue)),
	}, nil
}

// loadHonor loads the honor definition by ID, returning nil when it does not exist.
func (s *serviceImpl) loadHonor(ctx context.Context, honorID int64) (*entitymodel.HonorDef, error) {
	var honor *entitymodel.HonorDef
	if err := dao.HonorDef.Ctx(ctx).Where(dao.HonorDef.Columns().Id, honorID).Scan(&honor); err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	return honor, nil
}

// eligibleUserIDs selects the player cohort that satisfies the certificate's unlock
// rule with a single set-based query. Collection-based rules are rejected.
func (s *serviceImpl) eligibleUserIDs(ctx context.Context, honor *entitymodel.HonorDef) ([]int64, error) {
	switch honor.UnlockType {
	case unlockParticipation:
		return s.scanIDs(ctx, dao.User.Ctx(ctx).Fields(dao.User.Columns().Id+" AS id"))
	case unlockFeedCount:
		return s.scanIDs(ctx, dao.Feeding.Ctx(ctx).
			Fields(dao.Feeding.Columns().UserId+" AS id").
			Group(dao.Feeding.Columns().UserId).
			Having("COUNT(*) >= ?", thresholdOrOne(honor.Threshold)))
	case unlockActivationCount:
		return s.scanIDs(ctx, dao.Activation.Ctx(ctx).
			Fields(dao.Activation.Columns().UserId+" AS id").
			Group(dao.Activation.Columns().UserId).
			Having("COUNT(*) >= ?", thresholdOrOne(honor.Threshold)))
	default:
		// category_complete / full_complete and any unknown rule: collection rules
		// are unlocked individually in the mini-program and are not batch-settled.
		return nil, bizerr.NewCode(CodeSettlementUnlockUnsupported)
	}
}

// grantedUserIDs returns the set of eligible players already granted the honor,
// fetched in one projected query restricted to the eligible cohort.
func (s *serviceImpl) grantedUserIDs(ctx context.Context, honorID int64, eligible []int64) (map[int64]bool, error) {
	granted := make(map[int64]bool, len(eligible))
	rows, err := s.scanIDs(ctx, dao.UserHonor.Ctx(ctx).
		Where(dao.UserHonor.Columns().HonorId, honorID).
		WhereIn(dao.UserHonor.Columns().UserId, eligible).
		Fields(dao.UserHonor.Columns().UserId+" AS id"))
	if err != nil {
		return nil, err
	}
	for _, id := range rows {
		granted[id] = true
	}
	return granted, nil
}

// insertGrants writes the new grants in a single transaction. The user_honor unique
// index on (user_id, honor_id) guarantees idempotency under concurrent runs.
func (s *serviceImpl) insertGrants(ctx context.Context, honorID int64, userIDs []int64) error {
	now := time.Now()
	data := make([]do.UserHonor, 0, len(userIDs))
	for _, uid := range userIDs {
		data = append(data, do.UserHonor{
			UserId:     uid,
			HonorId:    honorID,
			UnlockedAt: &now,
		})
	}
	err := dao.UserHonor.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, txErr := dao.UserHonor.Ctx(ctx).Data(data).Insert()
		return txErr
	})
	if err != nil {
		return bizerr.WrapCode(err, CodeSettlementIssueFailed)
	}
	return nil
}

// certificateHonorIDs returns the IDs of all certificate honor definitions in one
// projected query.
func (s *serviceImpl) certificateHonorIDs(ctx context.Context) ([]int64, error) {
	return s.scanIDs(ctx, dao.HonorDef.Ctx(ctx).
		Where(dao.HonorDef.Columns().HonorType, certificateHonorType).
		Fields(dao.HonorDef.Columns().Id+" AS id"))
}

// scanIDs runs the prepared single-column "id" projection model and returns the IDs.
func (s *serviceImpl) scanIDs(ctx context.Context, model *gdb.Model) ([]int64, error) {
	var rows []*idRow
	if err := model.Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.Id)
	}
	return ids, nil
}

// thresholdOrOne returns threshold when positive, otherwise 1, so a count-based
// certificate always requires at least one qualifying action.
func thresholdOrOne(threshold int) int {
	if threshold > 0 {
		return threshold
	}
	return 1
}
