// ranking_code.go defines the leaderboard business error codes.

package ranking

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// defaultTopN bounds each board when no positive Top-N is configured.
const defaultTopN = 100

var (
	// CodeRankingQueryFailed reports that a leaderboard aggregation query failed.
	CodeRankingQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_RANKING_QUERY_FAILED",
		"Failed to query leaderboard data",
		gcode.CodeInternalError,
	)
)
