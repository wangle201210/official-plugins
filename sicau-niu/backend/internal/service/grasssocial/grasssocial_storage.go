// grasssocial_storage.go provides deterministic player-row locking for
// quota-bearing grass transfers and reciprocal-transfer deadlock prevention.
package grasssocial

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// lockSocialPlayers serializes quota-bearing writes for the actor and prevents
// reciprocal transfers from locking their grass accounts in opposite orders.
func lockSocialPlayers(ctx context.Context, firstUserID, secondUserID int64) error {
	if firstUserID > secondUserID {
		firstUserID, secondUserID = secondUserID, firstUserID
	}
	for _, userID := range []int64{firstUserID, secondUserID} {
		var player *entitymodel.User
		if err := dao.User.Ctx(ctx).
			Fields(dao.User.Columns().Id).
			Where(do.User{Id: userID}).
			LockUpdate().
			Scan(&player); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if player == nil {
			return bizerr.NewCode(CodeQueryFailed)
		}
	}
	return nil
}
