// cattle_pagination.go provides the shared pagination defaults and normalization
// for the cattle and iron-cow operator list queries.

package cattle

// Listing paging defaults bound the operator console page size for both the
// cattle and iron-cow lists.
const (
	defaultPageNum  = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// normalizePagination applies the paging defaults and the max page-size cap to a
// requested page number and size.
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
