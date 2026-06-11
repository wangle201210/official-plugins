// This file decorates legacy resource rows with the nested relation objects
// the old gorm Preload responses carried, keyed exactly like the old model
// JSON (untagged relation fields marshaled under their Go field names).

package uidentity

import (
	"context"

	"github.com/gogf/gf/v2/util/gconv"
)

// legacyResourceRelation describes one preloaded relation of the old list and
// detail responses: the related rows are attached under responseKey using the
// related resource projection.
type legacyResourceRelation struct {
	responseKey string
	resource    string
	idField     string
}

func legacyResourceRelations(resource string) []legacyResourceRelation {
	switch resource {
	case "account-app-blacklists":
		return []legacyResourceRelation{
			{responseKey: "App", resource: "applications", idField: "appId"},
			{responseKey: "Account", resource: "accounts", idField: "accountId"},
		}
	case "group-app-blacklists":
		return []legacyResourceRelation{
			{responseKey: "App", resource: "applications", idField: "appId"},
			{responseKey: "Group", resource: "groups", idField: "groupId"},
		}
	case "account-app-roles":
		return []legacyResourceRelation{
			{responseKey: "App", resource: "applications", idField: "appId"},
			{responseKey: "EmpoweredAccount", resource: "accounts", idField: "empoweredAccountId"},
		}
	case "cas-login-logs":
		return []legacyResourceRelation{
			{responseKey: "Account", resource: "accounts", idField: "accountId"},
			{responseKey: "ChoiceAccount", resource: "accounts", idField: "choiceAccountId"},
			{responseKey: "Application", resource: "applications", idField: "appId"},
		}
	case "oauth-logs":
		return []legacyResourceRelation{
			{responseKey: "Account", resource: "accounts", idField: "userId"},
			{responseKey: "Application", resource: "applications", idField: "appId"},
		}
	default:
		return nil
	}
}

// decorateResourceRecords attaches the nested relation objects the old
// responses preloaded, plus the account-specific flat decorations.
func (s *serviceImpl) decorateResourceRecords(ctx context.Context, def *resourceDefinition, records []Record) error {
	if def == nil || len(records) == 0 {
		return nil
	}
	if def.name == "accounts" {
		return s.decorateAccountRecords(ctx, records)
	}
	for _, relation := range legacyResourceRelations(def.name) {
		if err := s.attachRelatedRecords(ctx, relation, records); err != nil {
			return err
		}
	}
	return nil
}

func (s *serviceImpl) attachRelatedRecords(ctx context.Context, relation legacyResourceRelation, records []Record) error {
	ids := make([]int64, 0, len(records))
	seen := make(map[int64]struct{}, len(records))
	for _, record := range records {
		id := gconv.Int64(record[relation.idField])
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}
	related, err := s.relatedRecordsByID(ctx, relation.resource, ids)
	if err != nil {
		return err
	}
	for _, record := range records {
		if value, ok := related[gconv.Int64(record[relation.idField])]; ok {
			record[relation.responseKey] = value
		}
	}
	return nil
}

func (s *serviceImpl) relatedRecordsByID(ctx context.Context, resource string, ids []int64) (map[int64]Record, error) {
	def, err := s.resourceDefinition(resource)
	if err != nil {
		return nil, err
	}
	rows, err := def.model(ctx).
		Fields(projectionFields(def)...).
		WhereIn(def.idColumn, ids).
		All()
	if err != nil {
		return nil, err
	}
	result := make(map[int64]Record, len(rows))
	for _, row := range rows {
		record := projectRecord(row, def)
		if id := gconv.Int64(record["id"]); id > 0 {
			result[id] = record
		}
	}
	return result, nil
}
