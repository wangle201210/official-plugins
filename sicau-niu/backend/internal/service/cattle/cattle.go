// Package cattle implements the sicau-niu cattle content-asset capability:
// operator-facing CRUD for cattle (普通牛/特殊牛) and iron-cow identifiers. Cattle
// carry a unique serial code, a stable type/subtype enum, an optional linked
// college (reusing the C1 college dictionary), GPS anchors, a default-inactive
// status and an optional online-time visibility schedule. Deleting a cattle cascade soft-deletes
// its unique main card in one transaction so no dangling card remains. Iron-cow
// rows register a unique device code; their real-time location is written later
// by the C4 bonus flow and is never set on the operator side. All store access
// uses the generated DAO/DO objects so GoFrame manages soft-delete and timestamp
// columns automatically. List queries run DB-side filtering/sorting/pagination
// and batch-assemble related college names and card-binding flags to avoid N+1
// queries.
package cattle

import (
	"context"

	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
)

// Service defines the cattle and iron-cow content-asset contract.
type Service interface {
	// ListNiu returns one DB-side paged, ID-descending cattle page for the
	// operator console. Filtering (keyword/type), ordering and pagination
	// run in the database; the current page's college names and card-binding
	// flags are batch-assembled in two bounded queries to avoid N+1. It returns a
	// query bizerr on store failure.
	ListNiu(ctx context.Context, in *ListNiuInput) (out *ListNiuOutput, err error)
	// GetNiu returns one cattle detail with its batch-assembled college name and
	// card-binding flag. It returns CodeNiuNotFound for a missing ID and a query
	// bizerr on store failure.
	GetNiu(ctx context.Context, id int64) (out *NiuItem, err error)
	// CreateNiu inserts one cattle after code-uniqueness, type/subtype enum and
	// college-existence validation. Status defaults to inactive. It returns the
	// relevant validation bizerr or the new ID on success.
	CreateNiu(ctx context.Context, in *NiuMutateInput) (id int64, err error)
	// UpdateNiu modifies one cattle after existence, code-uniqueness, enum and
	// college-existence validation. Status is not changed here. It returns
	// CodeNiuNotFound for a missing ID, the relevant validation bizerr, or nil on
	// success.
	UpdateNiu(ctx context.Context, id int64, in *NiuMutateInput) error
	// DeleteNiu soft-deletes one cattle and cascade soft-deletes its unique main
	// card in one transaction. It returns CodeNiuNotFound for a missing ID and a
	// write bizerr on store failure.
	DeleteNiu(ctx context.Context, id int64) error
	// NiuExists reports whether an active cattle with id exists. It is consumed by
	// the card service to validate card ownership without loading the full row. A
	// non-positive id reports false without a query.
	NiuExists(ctx context.Context, id int64) (exists bool, err error)

	// ListIron returns one DB-side paged, ID-descending iron-cow page for the
	// operator console. Filtering, ordering and pagination run in the database. It
	// returns a query bizerr on store failure.
	ListIron(ctx context.Context, in *ListIronInput) (out *ListIronOutput, err error)
	// CreateIron inserts one iron-cow registration after code-uniqueness
	// validation. Real-time location columns are not written here. It returns
	// CodeIronCodeRequired/CodeIronCodeExists or the new ID on success.
	CreateIron(ctx context.Context, in *IronMutateInput) (id int64, err error)
	// UpdateIron modifies one iron-cow's code, name and remark after existence and
	// code-uniqueness validation. It returns CodeIronNotFound for a missing ID,
	// CodeIronCodeExists on a conflicting code, or nil on success.
	UpdateIron(ctx context.Context, id int64, in *IronMutateInput) error
	// DeleteIron soft-deletes one iron-cow registration. It returns
	// CodeIronNotFound for a missing ID and a write bizerr on store failure.
	DeleteIron(ctx context.Context, id int64) error
}

// Interface compliance assertion for the default cattle service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned niu and iron tables.
// Its only runtime dependency is the college service used to validate the linked
// college of college-subtype cattle; it is injected explicitly so dependency
// changes surface at compile time.
type serviceImpl struct {
	collegeSvc collegesvc.Service // collegeSvc validates linked college existence.
}

// New creates a cattle service with explicit dependencies: the college service
// used for linked-college existence validation.
func New(collegeSvc collegesvc.Service) Service {
	return &serviceImpl{collegeSvc: collegeSvc}
}
