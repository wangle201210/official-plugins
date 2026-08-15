// activation_poster.go implements the activation poster composition data. It
// verifies the player has activated the target cattle (otherwise rejects), then
// assembles the poster fields: the player's nickname and identity type from the
// identity service, the cattle serial code, the player's arrival order for the
// cattle, a random enabled school-history quote and the configured campus badge.
// The composed fields are returned as data; the share poster itself is drawn by
// the mini-program on canvas, so no image is produced here.

package activation

import (
	"context"

	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
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
	// NiuName is the activated cattle display name; empty for unnamed common cattle.
	NiuName string
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

	niuCode, niuName, err := s.loadNiuLabels(ctx, niuID)
	if err != nil {
		return nil, err
	}

	quote, err := s.randomEnabledQuote(ctx)
	if err != nil {
		return nil, err
	}
	campusBadge, err := s.posterCampusBadge(ctx)
	if err != nil {
		return nil, err
	}

	output := &PosterOutput{
		Nickname:     profile.Nickname,
		IdentityType: profile.IdentityType,
		NiuCode:      niuCode,
		NiuName:      niuName,
		OrderNo:      record.OrderNo,
		Quote:        quote,
		CampusBadge:  campusBadge,
	}

	return output, nil
}

// loadNiuLabels loads the cattle serial code and display name by ID in one query.
// It returns CodeNiuNotFound when the cattle is missing; the name is empty for
// unnamed common cattle, and the poster then falls back to the code.
func (s *serviceImpl) loadNiuLabels(ctx context.Context, niuID int64) (code string, name string, err error) {
	var niuRow *entitymodel.Niu
	if err = dao.Niu.Ctx(ctx).
		Fields(dao.Niu.Columns().Code, dao.Niu.Columns().Name).
		Where(do.Niu{Id: niuID}).
		Scan(&niuRow); err != nil {
		return "", "", bizerr.WrapCode(err, CodeQueryFailed)
	}
	if niuRow == nil {
		return "", "", bizerr.NewCode(CodeNiuNotFound)
	}
	return niuRow.Code, niuRow.Name, nil
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

// posterCampusBadge returns the operator-maintained poster badge when rules are
// injected, otherwise the constructor fallback.
func (s *serviceImpl) posterCampusBadge(ctx context.Context) (string, error) {
	if s.rulesSvc == nil {
		return s.campusBadge, nil
	}
	return s.rulesSvc.PosterCampusBadge(ctx)
}
