// settlement_risk.go implements the shared-device risk view. The clusters are
// found on the database side by grouping players on a non-empty device fingerprint
// and keeping groups with more than one member; the member nicknames for the whole
// page are then batch-assembled in one projected WHERE IN query and grouped in
// memory, so no per-cluster lookup is issued.

package settlement

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// RiskClusters is the shared-device risk view result.
type RiskClusters struct {
	// List is the shared-device clusters ordered by member count descending,
	// bounded.
	List []*RiskCluster
}

// RiskCluster is one shared-device cluster.
type RiskCluster struct {
	// Fingerprint is the shared device fingerprint.
	Fingerprint string
	// Count is the number of players sharing the fingerprint.
	Count int64
	// Members are the players in the cluster.
	Members []*RiskMember
}

// RiskMember is one player in a shared-device cluster.
type RiskMember struct {
	// UserId is the player ID.
	UserId int64
	// Nickname is the player nickname; empty when unset.
	Nickname string
}

// fingerprintRow is the temporary projection for one grouped fingerprint/count
// aggregate.
type fingerprintRow struct {
	Fingerprint string `json:"fingerprint"`
	Cnt         int64  `json:"cnt"`
}

// RiskDeviceClusters returns the bounded shared-device clusters with member
// nicknames.
func (s *serviceImpl) RiskDeviceClusters(ctx context.Context) (*RiskClusters, error) {
	var groups []*fingerprintRow
	err := dao.User.Ctx(ctx).
		Where(dao.User.Columns().DeviceFingerprint+" <> ?", "").
		Fields(dao.User.Columns().DeviceFingerprint+" AS fingerprint", "COUNT(*) AS cnt").
		Group(dao.User.Columns().DeviceFingerprint).
		Having("COUNT(*) > 1").
		Order("cnt DESC").
		Limit(riskClusterLimit).
		Scan(&groups)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	if len(groups) == 0 {
		return &RiskClusters{List: []*RiskCluster{}}, nil
	}

	fingerprints := make([]string, 0, len(groups))
	for _, group := range groups {
		fingerprints = append(fingerprints, group.Fingerprint)
	}
	membersByFingerprint, err := s.batchClusterMembers(ctx, fingerprints)
	if err != nil {
		return nil, err
	}

	list := make([]*RiskCluster, 0, len(groups))
	for _, group := range groups {
		list = append(list, &RiskCluster{
			Fingerprint: group.Fingerprint,
			Count:       group.Cnt,
			Members:     membersByFingerprint[group.Fingerprint],
		})
	}
	return &RiskClusters{List: list}, nil
}

// batchClusterMembers loads all members of the given fingerprints in one projected
// query and groups them by fingerprint in memory to avoid a per-cluster lookup.
func (s *serviceImpl) batchClusterMembers(ctx context.Context, fingerprints []string) (map[string][]*RiskMember, error) {
	members := make(map[string][]*RiskMember, len(fingerprints))
	if len(fingerprints) == 0 {
		return members, nil
	}
	rows := make([]*entitymodel.User, 0, len(fingerprints)*2)
	err := dao.User.Ctx(ctx).
		Fields(
			dao.User.Columns().Id,
			dao.User.Columns().Nickname,
			dao.User.Columns().DeviceFingerprint,
		).
		WhereIn(dao.User.Columns().DeviceFingerprint, fingerprints).
		Order(dao.User.Columns().Id + " ASC").
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	for _, row := range rows {
		members[row.DeviceFingerprint] = append(members[row.DeviceFingerprint], &RiskMember{
			UserId:   row.Id,
			Nickname: row.Nickname,
		})
	}
	return members, nil
}
