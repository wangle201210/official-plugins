// identity_players.go implements the operator-facing read-only player query with
// DB-side filtering, ordering and pagination. It returns a bounded current page
// plus the total matched count and never loads the full player set into memory.

package identity

import (
	"context"
	"strings"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ListPlayersInput defines the operator player list query.
type ListPlayersInput struct {
	// Keyword is the optional fuzzy match applied to the player nickname.
	Keyword string
	// IdentityType optionally filters players by identity tag; empty lists all.
	IdentityType string
	// PageNum is the requested page number; defaults to 1 when non-positive.
	PageNum int
	// PageSize is the requested page size; defaults to 10 and is capped at 100.
	PageSize int
}

// ListPlayersOutput defines the operator player list result.
type ListPlayersOutput struct {
	// List holds the current page of players.
	List []*PlayerItem
	// Total is the total matched player count.
	Total int
}

// PlayerItem defines one player row projected for the operator console.
type PlayerItem struct {
	// Id is the player ID.
	Id int64
	// Nickname is the player nickname.
	Nickname string
	// Phone is the bound phone number, empty when not bound.
	Phone string
	// IdentityType is the identity tag string.
	IdentityType string
	// CollegeId is the selected college ID, 0 when none.
	CollegeId int64
	// Grade is the grade number, 0 when unset.
	Grade int
	// GraduationYear is the graduation year, 0 when unset.
	GraduationYear int
	// CreatedAt is the creation time as a Unix timestamp in milliseconds.
	CreatedAt *int64
	// UpdatedAt is the update time as a Unix timestamp in milliseconds.
	UpdatedAt *int64
}

// ListPlayers returns one DB-side paged, read-only player page.
func (s *serviceImpl) ListPlayers(ctx context.Context, in *ListPlayersInput) (*ListPlayersOutput, error) {
	pageNum, pageSize := normalizePlayerPagination(in)
	model := dao.User.Ctx(ctx)
	if in != nil {
		if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
			model = model.WhereLike(dao.User.Columns().Nickname, "%"+keyword+"%")
		}
		if identityType := strings.TrimSpace(in.IdentityType); identityType != "" {
			model = model.Where(do.User{IdentityType: identityType})
		}
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodePlayerQueryFailed)
	}

	rows := make([]*entitymodel.User, 0)
	err = model.
		Fields(
			dao.User.Columns().Id,
			dao.User.Columns().Nickname,
			dao.User.Columns().Phone,
			dao.User.Columns().IdentityType,
			dao.User.Columns().CollegeId,
			dao.User.Columns().Grade,
			dao.User.Columns().GraduationYear,
			dao.User.Columns().CreatedAt,
			dao.User.Columns().UpdatedAt,
		).
		OrderDesc(dao.User.Columns().Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodePlayerQueryFailed)
	}

	list := make([]*PlayerItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, &PlayerItem{
			Id:             row.Id,
			Nickname:       row.Nickname,
			Phone:          row.Phone,
			IdentityType:   row.IdentityType,
			CollegeId:      row.CollegeId,
			Grade:          row.Grade,
			GraduationYear: row.GraduationYear,
			CreatedAt:      apitime.Milli(row.CreatedAt),
			UpdatedAt:      apitime.Milli(row.UpdatedAt),
		})
	}
	return &ListPlayersOutput{List: list, Total: total}, nil
}

// Player listing paging defaults bound the operator console page size.
const (
	defaultPlayerPageNum  = 1
	defaultPlayerPageSize = 10
	maxPlayerPageSize     = 100
)

// normalizePlayerPagination applies paging defaults and the max page-size cap.
func normalizePlayerPagination(in *ListPlayersInput) (int, int) {
	if in == nil {
		return defaultPlayerPageNum, defaultPlayerPageSize
	}
	pageNum := in.PageNum
	if pageNum <= 0 {
		pageNum = defaultPlayerPageNum
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPlayerPageSize
	}
	if pageSize > maxPlayerPageSize {
		pageSize = maxPlayerPageSize
	}
	return pageNum, pageSize
}
