// ironlocation_reporting_cycle.go implements the v1.1 IOT locator command that
// updates one device's reporting cycle. It performs a fresh token request for
// each low-frequency operator action and requires an explicit per-device success
// result before reporting success to the caller.

package ironlocation

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

const (
	// minReportingCycle is the platform-documented minimum reporting cycle.
	minReportingCycle = 10 * time.Second
	// maxReportingCycle is the platform-documented maximum reporting cycle.
	maxReportingCycle = time.Hour
	// reportingCycleParam is the platform template field for cycle seconds.
	reportingCycleParam = "cycle"
	// setConfigPath is the v1.1 locator command endpoint.
	setConfigPath = "/open/device/setConfig"
)

// locatorConfigCode is one command code accepted by the IOT locator platform.
type locatorConfigCode string

const (
	// locatorConfigReportingCycle selects reporting-cycle configuration.
	locatorConfigReportingCycle locatorConfigCode = "REPORTING_CYCLE_CONFIG"
)

// locatorConfigStatus is one per-device command result status.
type locatorConfigStatus int

const (
	// locatorConfigStatusSuccess is the platform-documented successful status.
	locatorConfigStatusSuccess locatorConfigStatus = 0
)

// ReportingCycleUpdater updates one locator's reporting cycle on the IOT
// platform. The device code must be non-empty and cycle must be a whole number
// of seconds in the documented 10-to-3600-second range. Implementations return a
// stable bizerr when configuration, transport or per-device execution fails.
type ReportingCycleUpdater interface {
	// SetReportingCycle updates deviceCode to cycle and returns nil only when the
	// platform response contains that device with status 0.
	SetReportingCycle(ctx context.Context, deviceCode string, cycle time.Duration) error
}

// Interface compliance assertions for configured and deferred configurations.
var (
	_ ReportingCycleUpdater = (*iotReportingCycleUpdater)(nil)
	_ ReportingCycleUpdater = (*unconfiguredReportingCycleUpdater)(nil)
)

// iotReportingCycleUpdater owns immutable platform settings and an HTTP client.
// It intentionally keeps no token cache because the operator command is rare and
// is assembled independently from the scheduled refresher callback.
type iotReportingCycleUpdater struct {
	cfg    Config
	client HTTPClient
}

// unconfiguredReportingCycleUpdater defers missing credentials to the specific
// operator action so unrelated plugin routes can still start.
type unconfiguredReportingCycleUpdater struct{}

// NewIOTReportingCycleUpdater creates a configured reporting-cycle updater.
func NewIOTReportingCycleUpdater(cfg Config, client HTTPClient) (ReportingCycleUpdater, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, gerror.New("IOT reporting-cycle updater requires an HTTP client")
	}
	return &iotReportingCycleUpdater{cfg: normalized, client: client}, nil
}

// NewUnconfiguredReportingCycleUpdater creates a deferred-failure updater for
// deployments that intentionally omit IOT platform credentials.
func NewUnconfiguredReportingCycleUpdater() ReportingCycleUpdater {
	return &unconfiguredReportingCycleUpdater{}
}

// SetReportingCycle reports that the IOT platform is not configured.
func (*unconfiguredReportingCycleUpdater) SetReportingCycle(
	_ context.Context,
	_ string,
	_ time.Duration,
) error {
	return bizerr.NewCode(CodeIOTNotConfigured)
}

// SetReportingCycle obtains a fresh token and sends one fixed-size setConfig
// command. It does not read or write plugin database rows.
func (u *iotReportingCycleUpdater) SetReportingCycle(
	ctx context.Context,
	deviceCode string,
	cycle time.Duration,
) error {
	code := strings.TrimSpace(deviceCode)
	if code == "" ||
		cycle < minReportingCycle ||
		cycle > maxReportingCycle ||
		cycle%time.Second != 0 {
		return bizerr.NewCode(CodeIOTCommandInvalid)
	}

	token, err := requestIOTToken(ctx, u.cfg, u.client)
	if err != nil {
		return err
	}

	var response locatorConfigResponse
	err = postIOTJSON(
		ctx,
		u.cfg,
		u.client,
		setConfigPath,
		token,
		locatorConfigRequest{
			DeviceNoList: []string{code},
			Sign:         locatorSign,
			ConfigCode:   locatorConfigReportingCycle,
			Template: []locatorConfigParameter{
				{
					Params: reportingCycleParam,
					Value:  strconv.FormatInt(int64(cycle/time.Second), 10),
				},
			},
		},
		&response,
	)
	if err != nil {
		return err
	}
	if response.Code != http.StatusOK {
		return bizerr.WrapCode(
			gerror.Newf("iot set config failed code=%d msg=%s", response.Code, response.Msg),
			CodeIOTRequestFailed,
		)
	}

	for _, item := range response.Data.Data {
		if strings.TrimSpace(item.Name) != code {
			continue
		}
		if item.Status == nil {
			return bizerr.NewCode(CodeIOTResponseInvalid)
		}
		if *item.Status != locatorConfigStatusSuccess {
			return bizerr.WrapCode(
				gerror.Newf(
					"iot set config device=%s status=%d reason=%s",
					code,
					*item.Status,
					strings.TrimSpace(item.Value),
				),
				CodeIOTRequestFailed,
			)
		}
		return nil
	}
	return bizerr.NewCode(CodeIOTResponseInvalid)
}

// locatorConfigRequest is the documented setConfig request body.
type locatorConfigRequest struct {
	DeviceNoList []string                 `json:"deviceNoList"`
	Sign         string                   `json:"sign"`
	ConfigCode   locatorConfigCode        `json:"configCode"`
	Template     []locatorConfigParameter `json:"template"`
}

// locatorConfigParameter is one command template entry.
type locatorConfigParameter struct {
	Params string `json:"params"`
	Value  string `json:"value"`
}

// locatorConfigResponse is the documented setConfig response body.
type locatorConfigResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Count int                   `json:"count"`
		Data  []locatorConfigResult `json:"data"`
	} `json:"data"`
}

// locatorConfigResult is one per-device setConfig result.
type locatorConfigResult struct {
	Name   string               `json:"name"`
	Value  string               `json:"value"`
	Status *locatorConfigStatus `json:"status"`
}
