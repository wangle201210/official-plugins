// Package niu implements the sample backend services exposed by the sicau-niu
// source plugin. The service has no host or database dependencies: it returns a
// fixed in-memory dataset to demonstrate public and protected plugin routes
// without introducing storage, tenant data, or N+1 query paths.
package niu

import "context"

// Service defines the sicau-niu sample service contract.
type Service interface {
	// Ping returns the public ping payload used by route registration verification.
	// It never fails and always returns a non-nil output.
	Ping(ctx context.Context) (out *PingOutput, err error)
	// List returns the bounded sample cattle records, optionally filtered by a
	// case-insensitive keyword on the cattle name. An empty keyword returns the
	// full sample dataset. The returned slice is never nil; err is always nil
	// because the data source is a fixed in-memory constant.
	List(ctx context.Context, in *ListInput) (out *ListOutput, err error)
}

// Interface compliance assertion for the default sicau-niu service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service using a fixed in-memory sample dataset.
type serviceImpl struct{}

// New creates and returns a new sicau-niu sample service instance.
func New() Service {
	return &serviceImpl{}
}
