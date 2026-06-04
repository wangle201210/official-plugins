// Package card implements the sicau-niu card content-asset capability:
// operator-facing CRUD for school-history cards (校史卡片) and quotes (校史金句).
// Each card belongs to exactly one cattle (the one-card-per-cattle constraint
// enforced by an active-set uniqueness check on card.niu_id) and carries a stable
// category enum, a title, content and an image path obtained from the host file
// upload (this capability does not manage file storage itself). Quotes are a
// simple non-empty text pool for later random playback, with an enabled flag.
// All store access uses the generated DAO/DO objects so GoFrame manages
// soft-delete and timestamp columns automatically. List queries run DB-side
// filtering/sorting/pagination; the card list batch-assembles owning-cattle code
// and name in one bounded query to avoid N+1.
package card

import (
	"context"

	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// Service defines the card and quote content-asset contract.
type Service interface {
	// ListCard returns one DB-side paged, ID-descending card page for the operator
	// console with the owning-cattle code and name batch-assembled in one bounded
	// query. It returns a query bizerr on store failure.
	ListCard(ctx context.Context, in *ListCardInput) (out *ListCardOutput, err error)
	// GetCard returns one card detail with its batch-assembled owning-cattle code
	// and name. It returns CodeCardNotFound for a missing ID and a query bizerr on
	// store failure.
	GetCard(ctx context.Context, id int64) (out *CardItem, err error)
	// CreateCard inserts one card after category-enum, owning-cattle existence and
	// one-card-per-cattle validation. It returns CodeCardCategoryInvalid,
	// CodeCardNiuInvalid, CodeCardNiuTaken or the new ID on success.
	CreateCard(ctx context.Context, in *CardMutateInput) (id int64, err error)
	// UpdateCard modifies one card after existence, category-enum, owning-cattle
	// existence and one-card-per-cattle validation. It returns CodeCardNotFound for
	// a missing ID, the relevant validation bizerr, or nil on success.
	UpdateCard(ctx context.Context, id int64, in *CardMutateInput) error
	// DeleteCard soft-deletes one card, freeing its owning cattle to be bound
	// again. It returns CodeCardNotFound for a missing ID and a write bizerr on
	// store failure.
	DeleteCard(ctx context.Context, id int64) error

	// ListQuote returns one DB-side paged, ID-descending quote page for the
	// operator console. It returns a query bizerr on store failure.
	ListQuote(ctx context.Context, in *ListQuoteInput) (out *ListQuoteOutput, err error)
	// CreateQuote inserts one quote after non-empty content validation. Enabled
	// defaults to 1 when nil. It returns CodeQuoteContentRequired or the new ID on
	// success.
	CreateQuote(ctx context.Context, in *QuoteMutateInput) (id int64, err error)
	// UpdateQuote modifies one quote after existence and non-empty content
	// validation. A nil Enabled leaves the flag unchanged. It returns
	// CodeQuoteNotFound for a missing ID, CodeQuoteContentRequired, or nil on
	// success.
	UpdateQuote(ctx context.Context, id int64, in *QuoteMutateInput) error
	// DeleteQuote soft-deletes one quote. It returns CodeQuoteNotFound for a
	// missing ID and a write bizerr on store failure.
	DeleteQuote(ctx context.Context, id int64) error
}

// Interface compliance assertion for the default card service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned card and quote tables.
// Its only runtime dependency is the cattle service used to validate card
// ownership; it is injected explicitly so dependency changes surface at compile
// time.
type serviceImpl struct {
	cattleSvc cattlesvc.Service // cattleSvc validates owning-cattle existence.
}

// New creates a card service with explicit dependencies: the cattle service used
// for owning-cattle existence validation.
func New(cattleSvc cattlesvc.Service) Service {
	return &serviceImpl{cattleSvc: cattleSvc}
}
