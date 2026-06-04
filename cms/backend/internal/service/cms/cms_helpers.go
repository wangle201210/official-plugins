// This file contains shared CMS service normalization and actor helpers.

package cms

import "context"

// normalizePageNum applies the default first page when pagination input is invalid.
func normalizePageNum(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

// normalizePageSize applies default and maximum page-size limits.
func normalizePageSize(value int) int {
	if value < 1 {
		return 10
	}
	if value > 100 {
		return 100
	}
	return value
}

// currentUserID resolves the acting user ID from the host business context.
func (s *serviceImpl) currentUserID(ctx context.Context) int64 {
	return int64(s.bizCtxSvc.Current(ctx).UserID)
}
