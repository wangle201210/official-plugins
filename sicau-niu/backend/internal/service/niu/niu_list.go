// niu_list.go implements the protected sample cattle list for the sicau-niu
// service. The dataset is a fixed in-memory constant, so the list size is a
// bounded constant and the keyword filter runs entirely in memory without any
// database access or per-row query.

package niu

import (
	"context"
	"strings"
)

// ListInput defines the sample cattle list query.
type ListInput struct {
	// Keyword is the optional case-insensitive fuzzy filter applied to the name.
	Keyword string
}

// CattleRecord is one sample cattle record carried by the service layer. The
// CreatedAtMs field already uses the Unix-millisecond shape required by the API
// response boundary, so the controller projects it without further conversion.
type CattleRecord struct {
	// ID is the sample cattle identifier.
	ID int64
	// Name is the sample cattle name.
	Name string
	// Breed is the sample cattle breed.
	Breed string
	// WeightKg is the sample cattle live weight in kilograms.
	WeightKg int
	// CreatedAtMs is the record creation time as a Unix timestamp in milliseconds.
	CreatedAtMs int64
}

// ListOutput defines the sample cattle list result.
type ListOutput struct {
	// Items holds the matched sample cattle records; never nil.
	Items []*CattleRecord
	// Total is the number of matched sample cattle records.
	Total int
}

// sampleCattle is the fixed in-memory dataset returned by the sample list. Its
// length is a small constant, which bounds the response size and the keyword
// filter cost regardless of the request.
var sampleCattle = []*CattleRecord{
	{ID: 1, Name: "Daisy", Breed: "Simmental", WeightKg: 520, CreatedAtMs: 1717372800000},
	{ID: 2, Name: "Bella", Breed: "Angus", WeightKg: 480, CreatedAtMs: 1717459200000},
	{ID: 3, Name: "Luna", Breed: "Hereford", WeightKg: 505, CreatedAtMs: 1717545600000},
}

// List returns the bounded sample cattle records filtered by the optional keyword.
func (s *serviceImpl) List(_ context.Context, in *ListInput) (out *ListOutput, err error) {
	keyword := ""
	if in != nil {
		keyword = strings.TrimSpace(strings.ToLower(in.Keyword))
	}

	items := make([]*CattleRecord, 0, len(sampleCattle))
	for _, record := range sampleCattle {
		if keyword != "" && !strings.Contains(strings.ToLower(record.Name), keyword) {
			continue
		}
		items = append(items, record)
	}

	return &ListOutput{Items: items, Total: len(items)}, nil
}
