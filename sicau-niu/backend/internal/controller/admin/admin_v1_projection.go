// admin_v1_projection.go holds shared service-to-DTO projection helpers for the
// cattle and card handlers, so the list and detail handlers map service items to
// response items consistently in one place.

package admin

import (
	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// toNiuItem projects one service cattle item to its response DTO.
func toNiuItem(item *cattlesvc.NiuItem) *v1.NiuItem {
	return &v1.NiuItem{
		Id:              item.Id,
		Code:            item.Code,
		NiuType:         item.NiuType,
		SpecialSubtype:  item.SpecialSubtype,
		Name:            item.Name,
		CollegeId:       item.CollegeId,
		CollegeName:     item.CollegeName,
		Lat:             item.Lat,
		Lng:             item.Lng,
		ReleaseStage:    item.ReleaseStage,
		OnlineAt:        item.OnlineAt,
		VisibleWeekdays: item.VisibleWeekdays,
		VisibleStart:    item.VisibleStart,
		VisibleEnd:      item.VisibleEnd,
		Status:          item.Status,
		HasCard:         item.HasCard,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

// toCardItem projects one service card item to its response DTO.
func toCardItem(item *cardsvc.CardItem) *v1.CardItem {
	return &v1.CardItem{
		Id:        item.Id,
		NiuId:     item.NiuId,
		NiuCode:   item.NiuCode,
		NiuName:   item.NiuName,
		Category:  item.Category,
		Title:     item.Title,
		Content:   item.Content,
		ImagePath: item.ImagePath,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
