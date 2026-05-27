// This file defines typed lifecycle callback DTOs for the dynamic plugin sample.

package v1

import "lina-core/pkg/plugin/pluginbridge/protocol"

// LifecycleDecisionRes is returned by dynamic lifecycle callbacks.
type LifecycleDecisionRes = protocol.LifecycleDecision

// BeforeInstallReq is the typed dynamic lifecycle request for BeforeInstall.
type BeforeInstallReq = protocol.LifecycleRequest

// AfterInstallReq is the typed dynamic lifecycle request for AfterInstall.
type AfterInstallReq = protocol.LifecycleRequest

// BeforeUpgradeReq is the typed dynamic lifecycle request for BeforeUpgrade.
type BeforeUpgradeReq = protocol.LifecycleRequest

// UpgradeReq is the typed dynamic lifecycle request for Upgrade.
type UpgradeReq = protocol.LifecycleRequest

// AfterUpgradeReq is the typed dynamic lifecycle request for AfterUpgrade.
type AfterUpgradeReq = protocol.LifecycleRequest

// BeforeDisableReq is the typed dynamic lifecycle request for BeforeDisable.
type BeforeDisableReq = protocol.LifecycleRequest

// AfterDisableReq is the typed dynamic lifecycle request for AfterDisable.
type AfterDisableReq = protocol.LifecycleRequest

// BeforeUninstallReq is the typed dynamic lifecycle request for BeforeUninstall.
type BeforeUninstallReq = protocol.LifecycleRequest

// UninstallReq is the typed dynamic lifecycle request for Uninstall.
type UninstallReq = protocol.LifecycleRequest

// AfterUninstallReq is the typed dynamic lifecycle request for AfterUninstall.
type AfterUninstallReq = protocol.LifecycleRequest

// BeforeTenantDisableReq is the typed dynamic lifecycle request for BeforeTenantDisable.
type BeforeTenantDisableReq = protocol.LifecycleRequest

// AfterTenantDisableReq is the typed dynamic lifecycle request for AfterTenantDisable.
type AfterTenantDisableReq = protocol.LifecycleRequest

// BeforeTenantDeleteReq is the typed dynamic lifecycle request for BeforeTenantDelete.
type BeforeTenantDeleteReq = protocol.LifecycleRequest

// AfterTenantDeleteReq is the typed dynamic lifecycle request for AfterTenantDelete.
type AfterTenantDeleteReq = protocol.LifecycleRequest

// BeforeInstallModeChangeReq is the typed dynamic lifecycle request for BeforeInstallModeChange.
type BeforeInstallModeChangeReq = protocol.LifecycleRequest

// AfterInstallModeChangeReq is the typed dynamic lifecycle request for AfterInstallModeChange.
type AfterInstallModeChangeReq = protocol.LifecycleRequest
