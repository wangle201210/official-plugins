// This file restores the old AddOrModifyLdapUser write-path behavior: account
// create, update, and import upsert one LDAP entry immediately instead of
// waiting for the nightly mysql->ldap sync job. It reuses the runtime LDAP
// config and connection helpers and mirrors the job's DN, objectClass, and
// attribute mapping. LDAP failures are best-effort and never fail the write,
// matching the old service which logged and continued.

package uidentity

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"

	"lina-core/pkg/logger"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	configKeyRuntimeLDAPObjectClass = "legacy.jobs.ldap.objectClass"
	configKeyRuntimeLDAPDefaultPass = "legacy.jobs.ldap.defaultPassword"

	defaultRuntimeLDAPPassword = "Isicau@2024"
	defaultRuntimeLDAPPeopleOU = "People"
)

var defaultRuntimeLDAPObjectClass = []string{"top", "person", "inetOrgPerson", "organizationalPerson", "xUserObjectClass"}

// syncAccountLDAPByID upserts one account into LDAP on the write path. It is a
// best-effort operation: when LDAP is unconfigured or unreachable it logs and
// returns nil so account writes keep working without LDAP, and the nightly
// mysql->ldap job remains the backstop.
func (s *serviceImpl) syncAccountLDAPByID(ctx context.Context, accountID int64) {
	if err := s.upsertAccountLDAPByID(ctx, accountID); err != nil {
		logger.Warningf(ctx, "legacy account ldap sync failed accountID=%d err=%v", accountID, err)
	}
}

func (s *serviceImpl) upsertAccountLDAPByID(ctx context.Context, accountID int64) error {
	cfg, err := s.runtimeLDAPConfig(ctx)
	if err != nil {
		// LDAP is not configured; skip immediate sync and rely on the job.
		return nil
	}
	account, err := s.getAccountByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account.ContainerId <= 0 {
		// Without a container the old GetLdapAddData returned nil and skipped.
		return nil
	}
	container, err := s.containerByID(ctx, account.ContainerId)
	if err != nil || container == nil {
		return err
	}
	detail, err := s.accountDetailByAccountID(ctx, accountID)
	if err != nil {
		return err
	}
	objectClass, err := s.runtimeLDAPObjectClass(ctx)
	if err != nil {
		return err
	}
	defaultPassword, err := s.configSvc.String(ctx, configKeyRuntimeLDAPDefaultPass, defaultRuntimeLDAPPassword)
	if err != nil {
		return err
	}
	conn, err := ldap.DialURL(cfg.addr, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: cfg.skipTLSVerify}))
	if err != nil {
		return err
	}
	defer conn.Close()
	if cfg.bindDN != "" || cfg.bindPassword != "" {
		if err := conn.Bind(cfg.bindDN, cfg.bindPassword); err != nil {
			return err
		}
	}
	entry, err := searchLegacyLDAPAccountFull(conn, cfg.baseDN, account.Number)
	if err != nil {
		return err
	}
	if entry == nil {
		return addLegacyLDAPAccount(conn, cfg.baseDN, objectClass, defaultPassword, account, detail, container)
	}
	return modifyLegacyLDAPAccount(conn, cfg.baseDN, entry, account, detail, container)
}

func (s *serviceImpl) containerByID(ctx context.Context, id int64) (*entity.Containers, error) {
	var container *entity.Containers
	if err := dao.Containers.Ctx(ctx).
		Where(dao.Containers.Columns().Id, id).
		Scan(&container); err != nil {
		return nil, err
	}
	return container, nil
}

func (s *serviceImpl) runtimeLDAPObjectClass(ctx context.Context) ([]string, error) {
	if s == nil || s.configSvc == nil {
		return defaultRuntimeLDAPObjectClass, nil
	}
	value, err := s.configSvc.Get(ctx, configKeyRuntimeLDAPObjectClass, nil)
	if err != nil {
		return nil, err
	}
	if value == nil || value.IsNil() {
		return defaultRuntimeLDAPObjectClass, nil
	}
	result := make([]string, 0)
	for _, item := range value.Strings() {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return defaultRuntimeLDAPObjectClass, nil
	}
	return result, nil
}

// searchLegacyLDAPAccountFull reads the attributes needed to diff one entry.
func searchLegacyLDAPAccountFull(conn *ldap.Conn, baseDN string, number string) (*ldap.Entry, error) {
	req := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(uid=%s))", ldap.EscapeFilter(strings.TrimSpace(number))),
		[]string{"cn", "sn", "telephoneNumber", "inetUserStatus", "securityEmail", "openUID", "usertype"},
		nil,
	)
	result, err := conn.Search(req)
	if err != nil {
		return nil, err
	}
	if len(result.Entries) == 0 {
		return nil, nil
	}
	return result.Entries[0], nil
}

func addLegacyLDAPAccount(
	conn *ldap.Conn,
	baseDN string,
	objectClass []string,
	defaultPassword string,
	account *entity.Account,
	detail *entity.AccountDetails,
	container *entity.Containers,
) error {
	password, err := generateLegacyLDAPSSHA(defaultPassword)
	if err != nil {
		return err
	}
	req := ldap.NewAddRequest(legacyLDAPAccountDN(baseDN, container.Name, account.Number), nil)
	req.Attribute("objectClass", objectClass)
	req.Attribute("uid", []string{account.Number})
	req.Attribute("sn", []string{account.Name})
	req.Attribute("cn", []string{account.Name})
	req.Attribute("userPassword", []string{password})
	if account.Phone != "" {
		req.Attribute("telephoneNumber", []string{account.Phone})
	}
	if detail != nil && detail.Email != "" {
		req.Attribute("securityEmail", []string{detail.Email})
	}
	if account.Status == AccountStatusNormal {
		req.Attribute("inetUserStatus", []string{"Active"})
	}
	req.Attribute("usertype", []string{container.Name})
	return conn.Add(req)
}

func modifyLegacyLDAPAccount(
	conn *ldap.Conn,
	baseDN string,
	entry *ldap.Entry,
	account *entity.Account,
	detail *entity.AccountDetails,
	container *entity.Containers,
) error {
	modify := ldap.NewModifyRequest(entry.DN, nil)
	changed := false
	replace := func(name string, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		if entry.GetAttributeValue(name) != value {
			modify.Replace(name, []string{value})
			changed = true
		}
	}
	replace("cn", account.Name)
	replace("sn", account.Name)
	replace("telephoneNumber", account.Phone)
	if account.Status == AccountStatusNormal {
		replace("inetUserStatus", "Active")
	}
	if detail != nil {
		replace("securityEmail", detail.Email)
		replace("openUID", detail.Wechat)
	}
	if changed {
		if err := conn.Modify(modify); err != nil {
			return err
		}
	}
	// Mirror the old usertype/container reconciliation: move the entry when the
	// account's container no longer matches the LDAP usertype OU.
	if usertype := entry.GetAttributeValue("usertype"); usertype != container.Name {
		move := ldap.NewModifyDNRequest(
			entry.DN,
			fmt.Sprintf("uid=%s", account.Number),
			true,
			legacyLDAPContainerDN(baseDN, container.Name),
		)
		if err := conn.ModifyDN(move); err != nil {
			return err
		}
		typeModify := ldap.NewModifyRequest(legacyLDAPAccountDN(baseDN, container.Name, account.Number), nil)
		typeModify.Replace("usertype", []string{container.Name})
		if err := conn.Modify(typeModify); err != nil {
			return err
		}
	}
	return nil
}

func legacyLDAPAccountDN(baseDN string, container string, number string) string {
	return fmt.Sprintf("uid=%s,%s", number, legacyLDAPContainerDN(baseDN, container))
}

func legacyLDAPContainerDN(baseDN string, container string) string {
	return fmt.Sprintf("ou=%s,ou=%s,%s", container, defaultRuntimeLDAPPeopleOU, baseDN)
}
