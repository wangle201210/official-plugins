// This file declares the managed scheduled-job entries contributed by the
// UIdentity CAS source plugin to LinaPro's unified task management.

package jobs

import (
	"context"

	"lina-core/pkg/plugin/pluginhost"
)

const (
	jobSyncMysql2LDAP      = "SyncMysql2Ldap"
	jobSyncStudent         = "SyncStudent"
	jobSyncStudentYJS      = "SyncStudentYJS"
	jobSyncStudentWJ       = "SyncStudentWJ"
	jobSyncDept            = "SyncDept"
	jobSyncJzg             = "SyncJzg"
	jobChangeContainer     = "ChangeContainer"
	jobNewContainerAccount = "NewContainerAccount"
	jobWannaT              = "WannaT"

	patternDailyAt0100 = "# 0 1 * * *"
	patternDailyAt0115 = "# 15 1 * * *"
	patternDailyAt0130 = "# 30 1 * * *"
	patternDailyAt0145 = "# 45 1 * * *"
	patternDailyAt0200 = "# 0 2 * * *"
	patternDailyAt0215 = "# 15 2 * * *"
	patternDailyAt0230 = "# 30 2 * * *"
	patternDailyAt0245 = "# 45 2 * * *"
	patternEveryHour   = "# 0 * * * *"
)

type managedJobDeclaration struct {
	name        string
	displayName string
	description string
	pattern     string
	handler     pluginhost.CronJobHandler
}

// Register contributes all migrated uidentity/admin app/jobs handlers to the host.
func (s *serviceImpl) Register(ctx context.Context, registrar pluginhost.CronRegistrar) error {
	for _, declaration := range s.declarations() {
		if err := registrar.AddWithMetadata(
			ctx,
			declaration.pattern,
			declaration.name,
			declaration.displayName,
			declaration.description,
			declaration.handler,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *serviceImpl) declarations() []managedJobDeclaration {
	return []managedJobDeclaration{
		{
			name:        jobSyncDept,
			displayName: "UIdentity Sync Departments",
			description: "Synchronizes Oracle organization departments into UIdentity units.",
			pattern:     patternDailyAt0100,
			handler:     s.syncDept,
		},
		{
			name:        jobSyncJzg,
			displayName: "UIdentity Sync Staff",
			description: "Synchronizes Oracle staff records into UIdentity accounts.",
			pattern:     patternDailyAt0115,
			handler:     s.syncJzg,
		},
		{
			name:        jobSyncStudent,
			displayName: "UIdentity Sync Undergraduate Students",
			description: "Synchronizes Oracle undergraduate student records into UIdentity accounts.",
			pattern:     patternDailyAt0130,
			handler:     s.syncStudent,
		},
		{
			name:        jobSyncStudentYJS,
			displayName: "UIdentity Sync Graduate Students",
			description: "Synchronizes Oracle graduate student records into UIdentity accounts.",
			pattern:     patternDailyAt0145,
			handler:     s.syncStudentYJS,
		},
		{
			name:        jobSyncStudentWJ,
			displayName: "UIdentity Sync Online-Education Students",
			description: "Synchronizes Oracle online-education student records into UIdentity accounts.",
			pattern:     patternDailyAt0200,
			handler:     s.syncStudentWJ,
		},
		{
			name:        jobChangeContainer,
			displayName: "UIdentity Change Graduation Container",
			description: "Moves accounts graduating in the current year into the xy container.",
			pattern:     patternDailyAt0215,
			handler:     s.changeContainer,
		},
		{
			name:        jobNewContainerAccount,
			displayName: "UIdentity Refresh Container Account Counts",
			description: "Refreshes UIdentity container account counters with set-based database updates.",
			pattern:     patternDailyAt0230,
			handler:     s.newContainerAccount,
		},
		{
			name:        jobSyncMysql2LDAP,
			displayName: "UIdentity Sync Accounts To LDAP",
			description: "Synchronizes UIdentity accounts into the configured LDAP directory.",
			pattern:     patternDailyAt0245,
			handler:     s.syncMysql2LDAP,
		},
		{
			name:        jobWannaT,
			displayName: "UIdentity WannaT Probe",
			description: "Runs the legacy WannaT lightweight probe task.",
			pattern:     patternEveryHour,
			handler:     s.wannaT,
		},
	}
}
