// Package record implements the sicau-niu activity-record read-only query
// capability: operator-facing paged queries over the C3/C4 behaviour tables
// (feeding, steal, gift, check-in, activation and the grass-account ledger). Every
// query filters and paginates on the database side and never loads the full set
// into memory; the player nicknames and cattle names shown on each page are
// batch-assembled in one projected WHERE IN query each to avoid N+1. The capability
// is read-only and owns no table; all store access uses the generated DAO/DO
// objects so GoFrame manages soft-delete and timestamp columns.
package record

import "context"

// Pagination bounds shared by every record query.
const (
	// defaultPageNum is the page used when the request omits a positive page number.
	defaultPageNum = 1
	// defaultPageSize is the page size used when the request omits a positive size.
	defaultPageSize = 10
	// maxPageSize caps the page size so a record query is always bounded.
	maxPageSize = 100
)

// Service defines the activity-record read-only query contract.
type Service interface {
	// ListFeedings returns one DB-side paged feeding-record page with player and
	// cattle names batch-assembled.
	ListFeedings(ctx context.Context, in *ListFeedingsInput) (out *ListFeedingsOutput, err error)
	// ListSteals returns one DB-side paged steal-record page with actor and target
	// nicknames batch-assembled.
	ListSteals(ctx context.Context, in *ListStealsInput) (out *ListStealsOutput, err error)
	// ListGifts returns one DB-side paged gift-record page with sender and receiver
	// nicknames batch-assembled.
	ListGifts(ctx context.Context, in *ListGiftsInput) (out *ListGiftsOutput, err error)
	// ListCheckins returns one DB-side paged check-in-record page with player
	// nicknames batch-assembled.
	ListCheckins(ctx context.Context, in *ListCheckinsInput) (out *ListCheckinsOutput, err error)
	// ListActivations returns one DB-side paged activation-record page with player
	// and cattle names batch-assembled.
	ListActivations(ctx context.Context, in *ListActivationsInput) (out *ListActivationsOutput, err error)
	// ListGrassTxns returns one DB-side paged grass-ledger page with player
	// nicknames batch-assembled.
	ListGrassTxns(ctx context.Context, in *ListGrassTxnsInput) (out *ListGrassTxnsOutput, err error)
}

// Interface compliance assertion for the default record service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned behaviour tables. It
// holds no runtime dependency: every query reads the plugin's own tables through
// the generated DAO and the page size is fixed-capped.
type serviceImpl struct{}

// New creates an activity-record query service. The component reads the plugin's
// own tables through the generated DAO and therefore takes no dependencies.
func New() Service {
	return &serviceImpl{}
}

// normalizePagination applies the paging defaults and the max page-size cap.
func normalizePagination(pageNum, pageSize int) (int, int) {
	if pageNum <= 0 {
		pageNum = defaultPageNum
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return pageNum, pageSize
}
