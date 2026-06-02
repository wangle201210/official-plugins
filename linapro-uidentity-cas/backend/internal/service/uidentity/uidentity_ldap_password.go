// This file restores legacy LDAP password verification for CAS/OAuth runtime
// login. The account table intentionally does not store password hashes.

package uidentity

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

const (
	ldapPasswordPrefixMD5  = "{MD5}"
	ldapPasswordPrefixSSHA = "{SSHA}"

	configKeyRuntimeLDAPAddr          = "legacy.jobs.ldap.addr"
	configKeyRuntimeLDAPBindDN        = "legacy.jobs.ldap.bindDN"
	configKeyRuntimeLDAPBindPassword  = "legacy.jobs.ldap.bindPassword"
	configKeyRuntimeLDAPBaseDN        = "legacy.jobs.ldap.baseDN"
	configKeyRuntimeLDAPSkipTLSVerify = "legacy.jobs.ldap.skipTLSVerify"
)

type runtimeLDAPConfig struct {
	addr          string
	bindDN        string
	bindPassword  string
	baseDN        string
	skipTLSVerify bool
}

func (s *serviceImpl) verifyLegacyLDAPPassword(ctx context.Context, number string, password string) (bool, error) {
	cfg, err := s.runtimeLDAPConfig(ctx)
	if err != nil {
		return false, nil
	}
	conn, err := ldap.DialURL(cfg.addr, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: cfg.skipTLSVerify}))
	if err != nil {
		return false, err
	}
	defer conn.Close()
	if cfg.bindDN != "" || cfg.bindPassword != "" {
		if err := conn.Bind(cfg.bindDN, cfg.bindPassword); err != nil {
			return false, err
		}
	}
	req := ldap.NewSearchRequest(
		cfg.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(uid=%s))", ldap.EscapeFilter(strings.TrimSpace(number))),
		[]string{"userPassword"},
		nil,
	)
	result, err := conn.Search(req)
	if err != nil {
		return false, err
	}
	if len(result.Entries) == 0 {
		return false, nil
	}
	return validateLegacyLDAPPassword(password, result.Entries[0].GetAttributeValue("userPassword")), nil
}

func (s *serviceImpl) syncLegacyLDAPPassword(ctx context.Context, number string, password string) error {
	cfg, err := s.runtimeLDAPConfig(ctx)
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
	entry, err := searchLegacyLDAPAccount(conn, cfg.baseDN, number)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("ldap user does not exist: %s", strings.TrimSpace(number))
	}
	passwordHash, err := generateLegacyLDAPSSHA(password)
	if err != nil {
		return err
	}
	modify := ldap.NewModifyRequest(entry.DN, nil)
	modify.Replace("userPassword", []string{passwordHash})
	return conn.Modify(modify)
}

func (s *serviceImpl) runtimeLDAPConfig(ctx context.Context) (runtimeLDAPConfig, error) {
	if s == nil || s.configSvc == nil {
		return runtimeLDAPConfig{}, fmt.Errorf("ldap config is unavailable")
	}
	addr, err := s.configSvc.String(ctx, configKeyRuntimeLDAPAddr, "")
	if err != nil {
		return runtimeLDAPConfig{}, err
	}
	baseDN, err := s.configSvc.String(ctx, configKeyRuntimeLDAPBaseDN, "")
	if err != nil {
		return runtimeLDAPConfig{}, err
	}
	addr = strings.TrimSpace(addr)
	baseDN = strings.TrimSpace(baseDN)
	if addr == "" || baseDN == "" {
		return runtimeLDAPConfig{}, fmt.Errorf("ldap config is incomplete")
	}
	bindDN, err := s.configSvc.String(ctx, configKeyRuntimeLDAPBindDN, "")
	if err != nil {
		return runtimeLDAPConfig{}, err
	}
	bindPassword, err := s.configSvc.String(ctx, configKeyRuntimeLDAPBindPassword, "")
	if err != nil {
		return runtimeLDAPConfig{}, err
	}
	skipTLSVerify, err := s.configSvc.Bool(ctx, configKeyRuntimeLDAPSkipTLSVerify, true)
	if err != nil {
		return runtimeLDAPConfig{}, err
	}
	return runtimeLDAPConfig{
		addr:          addr,
		bindDN:        strings.TrimSpace(bindDN),
		bindPassword:  bindPassword,
		baseDN:        baseDN,
		skipTLSVerify: skipTLSVerify,
	}, nil
}

func searchLegacyLDAPAccount(conn *ldap.Conn, baseDN string, number string) (*ldap.Entry, error) {
	req := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(uid=%s))", ldap.EscapeFilter(strings.TrimSpace(number))),
		[]string{"userPassword"},
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

func validateLegacyLDAPPassword(password string, ldapPassword string) bool {
	if strings.HasPrefix(ldapPassword, ldapPasswordPrefixMD5) {
		return validateLegacyLDAPMD5(password, ldapPassword)
	}
	if strings.HasPrefix(ldapPassword, ldapPasswordPrefixSSHA) {
		return validateLegacyLDAPSSHA(password, ldapPassword)
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(ldapPassword)) == 1
}

func validateLegacyLDAPSSHA(password string, ldapPassword string) bool {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ldapPassword, ldapPasswordPrefixSSHA))
	if err != nil || len(decoded) <= sha1.Size {
		return false
	}
	storedHash := decoded[:sha1.Size]
	salt := decoded[sha1.Size:]
	hash := sha1.New()
	_, _ = hash.Write([]byte(password))
	_, _ = hash.Write(salt)
	return subtle.ConstantTimeCompare(hash.Sum(nil), storedHash) == 1
}

func validateLegacyLDAPMD5(password string, ldapPassword string) bool {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ldapPassword, ldapPasswordPrefixMD5))
	if err != nil || len(decoded) < md5.Size {
		return false
	}
	storedHash := decoded[:md5.Size]
	salt := decoded[md5.Size:]
	hash := md5.New()
	_, _ = hash.Write([]byte(password))
	_, _ = hash.Write(salt)
	return subtle.ConstantTimeCompare(hash.Sum(nil), storedHash) == 1
}

func generateLegacyLDAPSSHA(password string) (string, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := sha1.New()
	_, _ = hash.Write([]byte(password))
	_, _ = hash.Write(salt)
	return ldapPasswordPrefixSSHA + base64.StdEncoding.EncodeToString(append(hash.Sum(nil), salt...)), nil
}
