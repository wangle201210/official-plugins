// This file implements the old SyncMysql2Ldap job as a LinaPro task-management
// handler. It reads plugin-owned account data in pages and mirrors account
// attributes into a configured LDAP directory without owning scheduling state.

package jobs

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

type ldapJobConfig struct {
	addr            string
	bindDN          string
	bindPassword    string
	baseDN          string
	objectClass     []string
	defaultPassword string
	pageSize        int
	skipTLSVerify   bool
}

func (s *serviceImpl) syncMysql2LDAP(ctx context.Context) error {
	cfg, err := s.ldapJobConfig(ctx)
	if err != nil {
		return err
	}
	conn, err := openLDAPConn(cfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	tenantID := s.tenantID(ctx)
	stats := jobRunStats{}
	for page := 0; ; page++ {
		accounts, err := ldapAccountPage(ctx, tenantID, page, cfg.pageSize)
		if err != nil {
			return err
		}
		if len(accounts) == 0 {
			break
		}
		details, containers, err := ldapAccountRelations(ctx, tenantID, accounts)
		if err != nil {
			return err
		}
		for _, account := range accounts {
			detail := details[account.Id]
			container := containers[account.ContainerId]
			changed, added, err := syncOneLDAPAccount(conn, cfg, account, detail, container)
			if err != nil {
				stats.errNum++
				logger.Warningf(ctx, "uidentity ldap sync failed number=%s err=%v", account.Number, err)
				continue
			}
			if added {
				stats.createNum++
			}
			if changed {
				stats.updateNum++
			}
		}
		if len(accounts) < cfg.pageSize {
			break
		}
	}
	logger.Infof(ctx, "uidentity ldap sync finished tenant=%d stats=%v", tenantID, sqlLogFields(stats))
	return nil
}

func (s *serviceImpl) ldapJobConfig(ctx context.Context) (ldapJobConfig, error) {
	addr, err := s.requireConfigString(ctx, configKeyLDAPAddr)
	if err != nil {
		return ldapJobConfig{}, err
	}
	baseDN, err := s.requireConfigString(ctx, configKeyLDAPBaseDN)
	if err != nil {
		return ldapJobConfig{}, err
	}
	bindDN, err := s.configString(ctx, configKeyLDAPBindDN, "")
	if err != nil {
		return ldapJobConfig{}, err
	}
	bindPassword, err := s.configString(ctx, configKeyLDAPBindPassword, "")
	if err != nil {
		return ldapJobConfig{}, err
	}
	defaultPassword, err := s.configString(ctx, configKeyLDAPDefaultPass, defaultLDAPPassword)
	if err != nil {
		return ldapJobConfig{}, err
	}
	pageSize, err := s.configInt(ctx, configKeyLDAPSyncPageSize, defaultLDAPSearchPageSize)
	if err != nil {
		return ldapJobConfig{}, err
	}
	skipTLSVerify := true
	if s != nil && s.configSvc != nil {
		skipTLSVerify, err = s.configSvc.Bool(ctx, configKeyLDAPSkipTLSVerify, true)
		if err != nil {
			return ldapJobConfig{}, err
		}
	}
	objectClass, err := s.ldapObjectClass(ctx)
	if err != nil {
		return ldapJobConfig{}, err
	}
	return ldapJobConfig{
		addr:            addr,
		bindDN:          bindDN,
		bindPassword:    bindPassword,
		baseDN:          baseDN,
		objectClass:     objectClass,
		defaultPassword: defaultPassword,
		pageSize:        pageSize,
		skipTLSVerify:   skipTLSVerify,
	}, nil
}

func openLDAPConn(cfg ldapJobConfig) (*ldap.Conn, error) {
	conn, err := ldap.DialURL(cfg.addr, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: cfg.skipTLSVerify}))
	if err != nil {
		return nil, err
	}
	if cfg.bindDN != "" || cfg.bindPassword != "" {
		if err := conn.Bind(cfg.bindDN, cfg.bindPassword); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return conn, nil
}

func (s *serviceImpl) ldapObjectClass(ctx context.Context) ([]string, error) {
	defaults := []string{"top", "person", "inetOrgPerson", "organizationalPerson", "xUserObjectClass"}
	if s == nil || s.configSvc == nil {
		return defaults, nil
	}
	value, err := s.configSvc.Get(ctx, configKeyLDAPObjectClass)
	if err != nil {
		return nil, err
	}
	if value == nil || value.IsNil() {
		return defaults, nil
	}
	values := value.Strings()
	result := make([]string, 0, len(values))
	for _, item := range values {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return defaults, nil
	}
	return result, nil
}

func ldapAccountPage(ctx context.Context, tenantID int, page int, pageSize int) ([]*entity.Account, error) {
	var accounts []*entity.Account
	err := dao.Account.Ctx(ctx).
		Where(dao.Account.Columns().TenantId, tenantID).
		OrderAsc(dao.Account.Columns().Id).
		Offset(page * pageSize).
		Limit(pageSize).
		Scan(&accounts)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func ldapAccountRelations(ctx context.Context, tenantID int, accounts []*entity.Account) (map[int64]*entity.AccountDetail, map[int64]*entity.Container, error) {
	accountIDs := make(map[int64]struct{}, len(accounts))
	containerIDs := make(map[int64]struct{}, len(accounts))
	for _, account := range accounts {
		if account == nil {
			continue
		}
		accountIDs[account.Id] = struct{}{}
		containerIDs[account.ContainerId] = struct{}{}
	}
	details := make(map[int64]*entity.AccountDetail, len(accounts))
	if ids := int64sFromSet(accountIDs); len(ids) > 0 {
		var rows []*entity.AccountDetail
		if err := dao.AccountDetail.Ctx(ctx).
			Where(dao.AccountDetail.Columns().TenantId, tenantID).
			WhereIn(dao.AccountDetail.Columns().AccountId, ids).
			Scan(&rows); err != nil {
			return nil, nil, err
		}
		for _, row := range rows {
			if row != nil {
				details[row.AccountId] = row
			}
		}
	}
	containers := make(map[int64]*entity.Container, len(containerIDs))
	if ids := int64sFromSet(containerIDs); len(ids) > 0 {
		var rows []*entity.Container
		if err := dao.Container.Ctx(ctx).
			Where(dao.Container.Columns().TenantId, tenantID).
			WhereIn(dao.Container.Columns().Id, ids).
			Scan(&rows); err != nil {
			return nil, nil, err
		}
		for _, row := range rows {
			if row != nil {
				containers[row.Id] = row
			}
		}
	}
	return details, containers, nil
}

func syncOneLDAPAccount(conn *ldap.Conn, cfg ldapJobConfig, account *entity.Account, detail *entity.AccountDetail, container *entity.Container) (changed bool, added bool, err error) {
	if account == nil || strings.TrimSpace(account.Number) == "" {
		return false, false, pluginJobUnsupported()
	}
	if container == nil || strings.TrimSpace(container.Name) == "" {
		return false, false, pluginJobUnsupported()
	}
	entry, err := searchLDAPAccount(conn, cfg.baseDN, account.Number)
	if err != nil {
		return false, false, err
	}
	targetDN := ldapAccountDN(cfg.baseDN, container.Name, account.Number)
	if entry == nil {
		return false, true, addLDAPAccount(conn, cfg, account, detail, container)
	}
	if !strings.EqualFold(entry.DN, targetDN) {
		if err := conn.ModifyDN(ldap.NewModifyDNRequest(entry.DN, fmt.Sprintf("uid=%s", ldap.EscapeFilter(account.Number)), true, ldapContainerDN(cfg.baseDN, container.Name))); err != nil {
			return false, false, err
		}
		entry.DN = targetDN
		changed = true
	}
	modify := ldap.NewModifyRequest(entry.DN, nil)
	if buildLDAPModify(modify, entry, account, detail, container) {
		if err := conn.Modify(modify); err != nil {
			return false, false, err
		}
		changed = true
	}
	return changed, false, nil
}

func moveLDAPAccountToContainer(conn *ldap.Conn, cfg ldapJobConfig, account *entity.Account, target *entity.Container) error {
	if account == nil || strings.TrimSpace(account.Number) == "" {
		return pluginJobUnsupported()
	}
	if target == nil || strings.TrimSpace(target.Name) == "" {
		return pluginJobUnsupported()
	}
	entry, err := searchLDAPAccount(conn, cfg.baseDN, account.Number)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("ldap user does not exist: %s", account.Number)
	}
	targetDN := ldapAccountDN(cfg.baseDN, target.Name, account.Number)
	if !strings.EqualFold(entry.DN, targetDN) {
		if err := conn.ModifyDN(ldap.NewModifyDNRequest(entry.DN, fmt.Sprintf("uid=%s", account.Number), true, ldapContainerDN(cfg.baseDN, target.Name))); err != nil {
			return err
		}
	}
	modify := ldap.NewModifyRequest(targetDN, nil)
	modify.Replace("usertype", []string{target.Name})
	return conn.Modify(modify)
}

func searchLDAPAccount(conn *ldap.Conn, baseDN string, number string) (*ldap.Entry, error) {
	req := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(uid=%s))", ldap.EscapeFilter(number)),
		[]string{"*"},
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

func addLDAPAccount(conn *ldap.Conn, cfg ldapJobConfig, account *entity.Account, detail *entity.AccountDetail, container *entity.Container) error {
	req, err := buildLDAPAddRequest(cfg, account, detail, container)
	if err != nil {
		return err
	}
	return conn.Add(req)
}

func buildLDAPAddRequest(cfg ldapJobConfig, account *entity.Account, detail *entity.AccountDetail, container *entity.Container) (*ldap.AddRequest, error) {
	password, err := generateLDAPSSHA(cfg.defaultPassword)
	if err != nil {
		return nil, err
	}
	req := ldap.NewAddRequest(ldapAccountDN(cfg.baseDN, container.Name, account.Number), nil)
	req.Attribute("objectClass", cfg.objectClass)
	req.Attribute("uid", []string{account.Number})
	req.Attribute("sn", []string{account.Name})
	req.Attribute("cn", []string{account.Name})
	req.Attribute("userPassword", []string{password})
	if account.Phone != "" {
		req.Attribute("telephoneNumber", []string{account.Phone})
	}
	if detail != nil {
		if detail.Email != "" {
			req.Attribute("securityEmail", []string{detail.Email})
		}
	}
	if account.Status == 1 {
		req.Attribute("inetUserStatus", []string{"Active"})
	}
	req.Attribute("usertype", []string{container.Name})
	return req, nil
}

func buildLDAPModify(req *ldap.ModifyRequest, entry *ldap.Entry, account *entity.Account, detail *entity.AccountDetail, container *entity.Container) bool {
	changed := false
	replace := func(name string, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		if entry.GetAttributeValue(name) != value {
			req.Replace(name, []string{value})
			changed = true
		}
	}
	replace("cn", account.Name)
	replace("sn", account.Name)
	replace("telephoneNumber", account.Phone)
	if account.Status == 1 {
		replace("inetUserStatus", "Active")
	}
	if detail != nil {
		replace("securityEmail", detail.Email)
		replace("openUID", detail.Wechat)
	}
	return changed
}

func ldapAccountDN(baseDN string, container string, number string) string {
	return fmt.Sprintf("uid=%s,%s", number, ldapContainerDN(baseDN, container))
}

func ldapContainerDN(baseDN string, container string) string {
	return fmt.Sprintf("ou=%s,%s,%s", container, "ou="+defaultLDAPPeopleOU, baseDN)
}
