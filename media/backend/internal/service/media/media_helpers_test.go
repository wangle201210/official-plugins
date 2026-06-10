// This file verifies shared media service helper behavior.

package media

import "testing"

// TestNormalizePaginationCapsAtMaxPageSize verifies service-side page size bounds.
func TestNormalizePaginationCapsAtMaxPageSize(t *testing.T) {
	pageNum, pageSize := normalizePagination(1, maxPageSize+1)
	if pageNum != 1 {
		t.Fatalf("expected requested page number to be preserved, got %d", pageNum)
	}
	if pageSize != 10000 {
		t.Fatalf("expected page size to cap at 10000, got %d", pageSize)
	}
}
