// This file binds the sample plugin service to the guest-side capability
// host-service clients. The framework guest SDK provides real clients for
// wasip1 builds and unsupported stubs for ordinary Go test builds.

package dynamicservice

import (
	plugindata "lina-core/pkg/plugin/capability/data"
	"lina-core/pkg/plugin/capability/guest"
)

var guestServices = guest.Default()

// dataService abstracts the governed data facade used by the sample service.
type dataService interface {
	// Table starts one single-table governed data query builder.
	Table(table string) *plugindata.Query
	// Transaction executes one governed structured mutation transaction.
	Transaction(fn func(tx *plugindata.Tx) error) error
}

// newRuntimeHostService returns the guest-side runtime host client.
func newRuntimeHostService() runtimeHostService {
	return guestServices.Runtime()
}

// newStorageHostService returns the guest-side storage host client.
func newStorageHostService() storageHostService {
	return guestServices.Storage()
}

// newNetworkHostService returns the guest-side outbound network host client.
func newNetworkHostService() networkHostService {
	return guestServices.Network()
}

// newDataService returns the guest-side governed data facade.
func newDataService() dataService {
	return guestServices.Data()
}

// newCronHostService returns the guest-side cron registration host client.
func newCronHostService() cronHostService {
	return guestServices.Cron()
}

// newConfigHostService returns the guest-side plugin config host client.
func newConfigHostService() configHostService {
	return guestServices.Config()
}

// newManifestHostService returns the guest-side plugin manifest resource
// client.
func newManifestHostService() manifestHostService {
	return guestServices.Manifest()
}

// newHostConfigHostService returns the guest-side public host config client.
func newHostConfigHostService() hostConfigHostService {
	return guestServices.HostConfig()
}

// newOrgHostService returns the guest-side organization capability client.
func newOrgHostService() orgHostService {
	return guestServices.Org()
}

// newTenantHostService returns the guest-side tenant capability client.
func newTenantHostService() tenantHostService {
	return guestServices.Tenant()
}
