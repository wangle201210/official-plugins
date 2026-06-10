// This file implements the old uidentity/admin ULink SMS gateway client. The
// gateway activates only when plugin config provides the gateway address and
// account; without it SMS codes stay plugin-local records, which keeps tests
// and gateway-less deployments working.

package uidentity

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
)

const (
	configKeyLegacySMSGatewayAddr = "legacy.sms.gatewayAddr"
	configKeyLegacySMSLoginName   = "legacy.sms.loginName"
	configKeyLegacySMSPassword    = "legacy.sms.password"
	configKeyLegacySMSFeeType     = "legacy.sms.feeType"
	configKeyLegacySMSSignName    = "legacy.sms.signName"
	configKeyLegacySMSTemplate    = "legacy.sms.contentTemplate"

	// Defaults match the old common/sms/ulink.go request contract.
	defaultLegacySMSFeeType  = "2"
	defaultLegacySMSTemplate = "你的验证码为{code}"

	// legacySMSOKPrefix matches the old ULink success body prefix.
	legacySMSOKPrefix = "OK"
)

// legacySMSGatewayConfig carries one resolved ULink gateway configuration.
type legacySMSGatewayConfig struct {
	Addr      string
	LoginName string
	Password  string
	FeeType   string
	SignName  string
	Template  string
}

// enabled reports whether the gateway has the minimum send configuration.
func (c *legacySMSGatewayConfig) enabled() bool {
	return c != nil && strings.TrimSpace(c.Addr) != "" && strings.TrimSpace(c.LoginName) != ""
}

// content renders the old verification message for one code.
func (c *legacySMSGatewayConfig) content(code string) string {
	template := c.Template
	if strings.TrimSpace(template) == "" {
		template = defaultLegacySMSTemplate
	}
	return strings.ReplaceAll(template, "{code}", code)
}

// legacySMSGatewayConfig reads the ULink gateway settings from plugin config.
func (s *serviceImpl) legacySMSGatewayConfig(ctx context.Context) (*legacySMSGatewayConfig, error) {
	addr, err := s.configSvc.String(ctx, configKeyLegacySMSGatewayAddr, "")
	if err != nil {
		return nil, err
	}
	loginName, err := s.configSvc.String(ctx, configKeyLegacySMSLoginName, "")
	if err != nil {
		return nil, err
	}
	password, err := s.configSvc.String(ctx, configKeyLegacySMSPassword, "")
	if err != nil {
		return nil, err
	}
	feeType, err := s.configSvc.String(ctx, configKeyLegacySMSFeeType, defaultLegacySMSFeeType)
	if err != nil {
		return nil, err
	}
	signName, err := s.configSvc.String(ctx, configKeyLegacySMSSignName, "")
	if err != nil {
		return nil, err
	}
	template, err := s.configSvc.String(ctx, configKeyLegacySMSTemplate, defaultLegacySMSTemplate)
	if err != nil {
		return nil, err
	}
	return &legacySMSGatewayConfig{
		Addr:      addr,
		LoginName: loginName,
		Password:  password,
		FeeType:   feeType,
		SignName:  signName,
		Template:  template,
	}, nil
}

// sendLegacySMSGateway delivers one message through the old ULink GET API and
// returns the raw gateway response body.
func sendLegacySMSGateway(ctx context.Context, cfg *legacySMSGatewayConfig, phone string, content string) (string, error) {
	response, err := g.Client().
		SetTimeout(15*time.Second).
		Get(ctx, cfg.Addr, map[string]string{
			"LoginName":  cfg.LoginName,
			"Pwd":        cfg.Password,
			"FeeType":    cfg.FeeType,
			"Mobile":     phone,
			"Content":    content,
			"SignName":   cfg.SignName,
			"TimingDate": "",
			"ExtCode":    "",
			"TimeStamp":  "",
			"Signature":  "",
		})
	if err != nil {
		return "", bizerr.WrapCode(err, CodeSMSGatewayFailed)
	}
	defer response.Close()
	body := response.ReadAllString()
	if !strings.HasPrefix(body, legacySMSOKPrefix) {
		return body, bizerr.NewCode(CodeSMSGatewayFailed)
	}
	return body, nil
}
