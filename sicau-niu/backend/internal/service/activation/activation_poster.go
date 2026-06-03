// activation_poster.go implements the activation poster composition data. It
// verifies the player has activated the target cattle (otherwise rejects), then
// assembles the poster fields: the player's nickname and identity type from the
// identity service, the cattle serial code, the player's arrival order for the
// cattle, a random enabled school-history quote and the configured campus badge.
// The PNG output is produced through the replaceable PosterRenderer seam so the
// rendering implementation can change without touching this composition logic.

package activation

import (
	"context"

	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	"lina-plugin-sicau-niu/backend/internal/service/activation/internal/posterrender"
)

// quoteEnabledOn is the enabled-flag value selecting quotes that participate in
// random poster playback, matching the card capability's persisted enabled flag.
const quoteEnabledOn = 1

// PosterOutput defines the activation poster composition data.
type PosterOutput struct {
	// Nickname is the player nickname.
	Nickname string
	// IdentityType is the raw player identity type string; empty when unset.
	IdentityType string
	// NiuCode is the activated cattle serial code.
	NiuCode string
	// OrderNo is the player's arrival order for the cattle, starting at 1.
	OrderNo int
	// Quote is a random enabled school-history quote; empty when none exists.
	Quote string
	// CampusBadge is the campus anniversary badge text; empty when unset.
	CampusBadge string
}

// Poster returns the activation poster composition data for a cattle playerID has
// activated. It rejects with CodeActivationNotFound when no activation exists.
func (s *serviceImpl) Poster(ctx context.Context, playerID int64, niuID int64) (*PosterOutput, error) {
	if niuID <= 0 {
		return nil, bizerr.NewCode(CodeNiuIDRequired)
	}
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeActivationNotFound)
	}

	var record *entitymodel.Activation
	err := dao.Activation.Ctx(ctx).
		Where(dao.Activation.Columns().UserId, playerID).
		Where(dao.Activation.Columns().NiuId, niuID).
		Scan(&record)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeActivationNotFound)
	}

	profile, err := s.identitySvc.GetProfile(ctx, playerID)
	if err != nil {
		return nil, err
	}

	niuCode, err := s.loadNiuCode(ctx, niuID)
	if err != nil {
		return nil, err
	}

	quote, err := s.randomEnabledQuote(ctx)
	if err != nil {
		return nil, err
	}

	output := &PosterOutput{
		Nickname:     profile.Nickname,
		IdentityType: profile.IdentityType,
		NiuCode:      niuCode,
		OrderNo:      record.OrderNo,
		Quote:        quote,
		CampusBadge:  s.campusBadge,
	}

	// Exercise the PNG output seam so the basic renderer stays wired; the encoded
	// bytes are not part of the JSON contract and are discarded here.
	if _, renderErr := s.posterRenderer.Render(ctx, &posterrender.PosterData{
		Nickname:     output.Nickname,
		IdentityType: output.IdentityType,
		NiuCode:      output.NiuCode,
		OrderNo:      output.OrderNo,
		Quote:        output.Quote,
		CampusBadge:  output.CampusBadge,
	}); renderErr != nil {
		return nil, renderErr
	}

	return output, nil
}

// loadNiuCode loads the cattle serial code by ID. It returns CodeNiuNotFound when
// the cattle is missing.
func (s *serviceImpl) loadNiuCode(ctx context.Context, niuID int64) (string, error) {
	var niuRow *entitymodel.Niu
	err := dao.Niu.Ctx(ctx).
		Fields(dao.Niu.Columns().Code).
		Where(do.Niu{Id: niuID}).
		Scan(&niuRow)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeQueryFailed)
	}
	if niuRow == nil {
		return "", bizerr.NewCode(CodeNiuNotFound)
	}
	return niuRow.Code, nil
}

// randomEnabledQuote returns the content of one random enabled quote. It loads the
// enabled quote IDs once and picks one uniformly with grand; an empty pool yields
// an empty string so the poster still renders without a quote.
func (s *serviceImpl) randomEnabledQuote(ctx context.Context) (string, error) {
	rows := make([]*entitymodel.Quote, 0)
	err := dao.Quote.Ctx(ctx).
		Fields(dao.Quote.Columns().Content).
		Where(do.Quote{Enabled: quoteEnabledOn}).
		Scan(&rows)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeQueryFailed)
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[grand.Intn(len(rows))].Content, nil
}
