// Package college implements the sicau-niu college dictionary capability:
// operator-facing CRUD with name uniqueness, DB-side paged listing, sort
// maintenance and delete-time reference protection, plus a bounded player-facing
// options list for identity selection. College rows are plugin-owned business
// data maintained by operators and reused by player identity selection and
// later college-ranking aggregation. All store access uses the generated DAO/DO
// objects so GoFrame manages soft-delete and timestamp columns automatically.
package college

import (
	"context"
)

// Service defines the college dictionary contract.
type Service interface {
	// List returns one DB-side paged, sort-ordered college page for the operator
	// console. Filtering, ordering and pagination run in the database; the result
	// is a bounded current page plus the total matched count. A blank keyword
	// lists all colleges. It returns a query bizerr on store failure.
	List(ctx context.Context, in *ListInput) (out *ListOutput, err error)
	// Create inserts one college after trimming and name-uniqueness validation.
	// It returns CodeCollegeNameRequired for blank names, CodeCollegeNameExists
	// when an active college already owns the name, and the new ID on success.
	Create(ctx context.Context, in *MutateInput) (id int64, err error)
	// Update modifies one college's name and sort after existence and
	// name-uniqueness validation. It returns CodeCollegeNotFound for a missing
	// ID, CodeCollegeNameExists on a conflicting name, and nil on success.
	Update(ctx context.Context, id int64, in *MutateInput) error
	// Delete soft-deletes one college after verifying it is not referenced by any
	// player. It returns CodeCollegeNotFound for a missing ID,
	// CodeCollegeReferenced when at least one player selected it, and nil on
	// success.
	Delete(ctx context.Context, id int64) error
	// Options returns the bounded, sort-ordered college list for player identity
	// selection. The full active set is returned in one query with no pagination
	// because the dictionary is small and stable.
	Options(ctx context.Context) (out []*OptionItem, err error)
	// Exists reports whether an active college with id exists. It is consumed by
	// the player identity service to validate the selected college without
	// loading the full row. A non-positive id reports false without a query.
	Exists(ctx context.Context, id int64) (exists bool, err error)
}

// Interface compliance assertion for the default college service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned college table.
type serviceImpl struct{}

// New creates and returns a new college service instance. The component depends
// only on its generated DAO and therefore takes no runtime interface
// dependencies.
func New() Service {
	return &serviceImpl{}
}
