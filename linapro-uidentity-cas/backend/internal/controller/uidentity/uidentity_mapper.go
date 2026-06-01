// This file maps service-layer generic records and statistics into API DTOs.

package uidentity

import (
	v1 "lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

func toAPIRecords(records []uidentitysvc.Record) []v1.ResourceRecord {
	result := make([]v1.ResourceRecord, 0, len(records))
	for _, record := range records {
		result = append(result, v1.ResourceRecord(record))
	}
	return result
}

func toAPIStatItems(items []*uidentitysvc.StatItem) []*v1.StatItem {
	result := make([]*v1.StatItem, 0, len(items))
	for _, item := range items {
		result = append(result, &v1.StatItem{Name: item.Name, Total: item.Total})
	}
	return result
}
