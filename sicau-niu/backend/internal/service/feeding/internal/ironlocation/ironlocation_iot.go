// ironlocation_iot.go implements the IOT-platform refresh path for physical
// iron-cow locators. A plugin cron job calls Refresher.Refresh every configured
// interval; the refresher reads registered iron-cow codes in one query, pulls
// locator positions from the external platform, and writes matched coordinates
// back to the plugin iron table. Player-facing feeding requests never call this
// file directly; they keep using the stored-location gateway.

package ironlocation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

const (
	// defaultIOTBaseURL is the public base URL documented by the positioning platform.
	defaultIOTBaseURL = "https://xudongli.amorcloud.cn/exportIotPlus/iotPlus"
	// locatorSign is the fixed sign required by the locator endpoints.
	locatorSign = "LOCATOR"
	// defaultPageSize bounds each external locator-list page.
	defaultPageSize = 100
	// defaultMaxPages bounds one refresh cycle even if the external platform
	// reports a very large locator set.
	defaultMaxPages = 20
	// defaultTokenTTL reflects the documented default token validity of three
	// days, with a refresh buffer applied by normalizeConfig.
	defaultTokenTTL = 72 * time.Hour
	// tokenRefreshBuffer refreshes the token before its documented expiry.
	tokenRefreshBuffer = time.Hour
	// maxIOTResponseBytes bounds response reads from the external platform.
	maxIOTResponseBytes = 4 << 20
)

// HTTPClient is the explicit HTTP dependency used by the IOT refresher. A nil
// client passed to NewIOTRefresher falls back to http.DefaultClient.
type HTTPClient interface {
	// Do executes one HTTP request.
	Do(req *http.Request) (*http.Response, error)
}

// Config carries pure-value settings for the IOT-platform refresh path.
type Config struct {
	// BaseURL is the platform base URL, without endpoint path suffixes.
	BaseURL string
	// Key is the fixed public key used by getToken.
	Key string
	// Secret is the fixed private secret used by getToken.
	Secret string
	// PageSize is the locator-list page size.
	PageSize int
	// MaxPages is the maximum number of locator-list pages read per refresh.
	MaxPages int
	// TokenTTL is the documented token validity used for local refresh.
	TokenTTL time.Duration
}

// RefreshResult summarizes one refresh cycle for logs and tests.
type RefreshResult struct {
	// Skipped is true when another refresh cycle is still running.
	Skipped bool
	// Registered is the number of active local iron-cow registrations inspected.
	Registered int
	// Fetched is the number of locator records fetched from the external platform.
	Fetched int
	// Matched is the number of local registrations matched by locatorNo.
	Matched int
	// Updated is the number of local registrations updated successfully.
	Updated int
}

// Refresher updates local iron-cow positions from the external IOT platform.
type Refresher interface {
	// Refresh fetches the latest external locator coordinates and writes them to
	// the plugin iron table. It performs one local registration query and bounded
	// external pagination. The caller controls scheduling and primary-node gating.
	Refresh(ctx context.Context) (*RefreshResult, error)
}

// iotRefresher is the default Refresher implementation. It caches the external
// token in-process because the platform token is documented as valid for days
// while the refresh job runs periodically.
type iotRefresher struct {
	cfg    Config
	client HTTPClient

	refreshMu   sync.Mutex
	mu          sync.Mutex
	cachedToken string
	tokenExpiry time.Time
}

// NewIOTRefresher creates a Refresher for the documented IOT positioning
// platform. It validates only static configuration; missing credentials are a
// startup configuration error for callers that choose to enable this refresher.
func NewIOTRefresher(cfg Config, client HTTPClient) (Refresher, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &iotRefresher{cfg: normalized, client: client}, nil
}

// Refresh updates registered iron-cow rows from the external locator platform.
func (r *iotRefresher) Refresh(ctx context.Context) (*RefreshResult, error) {
	if !r.refreshMu.TryLock() {
		return &RefreshResult{Skipped: true}, nil
	}
	defer r.refreshMu.Unlock()

	rows, err := r.loadRegisteredIrons(ctx)
	if err != nil {
		return nil, err
	}
	result := &RefreshResult{Registered: len(rows)}
	if len(rows) == 0 {
		return result, nil
	}

	wanted := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		code := strings.TrimSpace(row.Code)
		if code != "" {
			wanted[code] = struct{}{}
		}
	}
	if len(wanted) == 0 {
		return result, nil
	}

	locations, fetched, err := r.fetchLocations(ctx, wanted)
	if err != nil {
		return nil, err
	}
	result.Fetched = fetched

	now := time.Now()
	updates := make([]ironLocationUpdate, 0, len(rows))
	for _, row := range rows {
		location, ok := locations[strings.TrimSpace(row.Code)]
		if !ok {
			continue
		}
		result.Matched++
		updates = append(updates, ironLocationUpdate{
			ID:  row.Id,
			Lat: location.Lat,
			Lng: location.Lng,
		})
	}
	if len(updates) == 0 {
		return result, nil
	}
	updated, err := r.updateLocations(ctx, updates, now)
	if err != nil {
		return nil, err
	}
	result.Updated = updated
	return result, nil
}

// updateLocations writes all matched coordinates in one UPDATE statement. The
// CASE expressions are built from parsed numeric values and DAO column names so
// the query stays bounded to one database round trip without interpolating user
// text.
func (r *iotRefresher) updateLocations(ctx context.Context, updates []ironLocationUpdate, locatedAt time.Time) (int, error) {
	cols := dao.Iron.Columns()
	ids := make([]int64, 0, len(updates))
	latCase := newCaseExpr(cols.Id, cols.LastLat)
	lngCase := newCaseExpr(cols.Id, cols.LastLng)
	for _, update := range updates {
		ids = append(ids, update.ID)
		latCase.Add(update.ID, update.Lat)
		lngCase.Add(update.ID, update.Lng)
	}
	result, err := dao.Iron.Ctx(ctx).
		WhereIn(cols.Id, ids).
		Data(do.Iron{
			LastLat:   gdb.Raw(latCase.String()),
			LastLng:   gdb.Raw(lngCase.String()),
			LocatedAt: &locatedAt,
		}).
		Update()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeQueryFailed)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return int(rowsAffected), nil
}

// loadRegisteredIrons reads all active iron-cow registrations in one projected query.
func (r *iotRefresher) loadRegisteredIrons(ctx context.Context) ([]*entitymodel.Iron, error) {
	rows := make([]*entitymodel.Iron, 0)
	err := dao.Iron.Ctx(ctx).
		Fields(dao.Iron.Columns().Id, dao.Iron.Columns().Code).
		OrderAsc(dao.Iron.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return rows, nil
}

// fetchLocations reads locator pages until all requested codes are found, the
// external count is exhausted, or MaxPages is reached.
func (r *iotRefresher) fetchLocations(
	ctx context.Context,
	wanted map[string]struct{},
) (map[string]locatorLocation, int, error) {
	token, err := r.token(ctx)
	if err != nil {
		return nil, 0, err
	}

	locations := make(map[string]locatorLocation, len(wanted))
	fetched := 0
	for page := 1; page <= r.cfg.MaxPages; page++ {
		body := locatorListRequest{
			Page:     page,
			PageSize: r.cfg.PageSize,
			Sign:     locatorSign,
		}
		var response locatorListResponse
		if err = r.postJSON(ctx, "/open/device/getInfoList", token, body, &response); err != nil {
			return nil, fetched, err
		}
		if response.Code != 200 {
			return nil, fetched, bizerr.WrapCode(
				gerror.Newf("iot locator list failed code=%d msg=%s", response.Code, response.Msg),
				CodeIOTRequestFailed,
			)
		}
		items := response.Data.Data
		fetched += len(items)
		for _, item := range items {
			location, ok := item.location()
			if !ok {
				continue
			}
			if _, wantedCode := wanted[location.Code]; !wantedCode {
				continue
			}
			locations[location.Code] = location
		}
		if len(locations) >= len(wanted) || len(items) == 0 {
			break
		}
		count := response.Data.Count
		if count > 0 && page*r.cfg.PageSize >= count {
			break
		}
	}
	return locations, fetched, nil
}

// token returns a cached platform token or refreshes it with getToken.
func (r *iotRefresher) token(ctx context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cachedToken != "" && time.Now().Before(r.tokenExpiry) {
		return r.cachedToken, nil
	}

	var response tokenResponse
	if err := r.postJSON(ctx, "/open/device/getToken", "", tokenRequest{
		Key:    r.cfg.Key,
		Secret: r.cfg.Secret,
	}, &response); err != nil {
		return "", err
	}
	if response.Code != 200 {
		return "", bizerr.WrapCode(
			gerror.Newf("iot token failed code=%d msg=%s", response.Code, response.Msg),
			CodeIOTRequestFailed,
		)
	}
	token := strings.TrimSpace(response.Data.Data)
	if token == "" {
		return "", bizerr.NewCode(CodeIOTResponseInvalid)
	}

	r.cachedToken = token
	r.tokenExpiry = time.Now().Add(r.cfg.TokenTTL - tokenRefreshBuffer)
	return token, nil
}

// postJSON posts a JSON body to one platform path and decodes the JSON response.
func (r *iotRefresher) postJSON(
	ctx context.Context,
	path string,
	token string,
	payload any,
	target any,
) (err error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return bizerr.WrapCode(err, CodeIOTResponseInvalid)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint(path), bytes.NewReader(data))
	if err != nil {
		return bizerr.WrapCode(err, CodeIOTRequestFailed)
	}
	request.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(token) != "" {
		request.Header.Set("Authorization", strings.TrimSpace(token))
	}

	response, err := r.client.Do(request)
	if err != nil {
		return bizerr.WrapCode(err, CodeIOTRequestFailed)
	}
	defer closeutil.Close(ctx, response.Body, &err, "close iot platform response body")

	body, err := io.ReadAll(io.LimitReader(response.Body, maxIOTResponseBytes))
	if err != nil {
		return bizerr.WrapCode(err, CodeIOTRequestFailed)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return bizerr.WrapCode(
			gerror.Newf("iot platform http status %d", response.StatusCode),
			CodeIOTRequestFailed,
		)
	}
	if err = json.Unmarshal(body, target); err != nil {
		return bizerr.WrapCode(err, CodeIOTResponseInvalid)
	}
	return nil
}

// endpoint joins the configured base URL with one documented endpoint path.
func (r *iotRefresher) endpoint(path string) string {
	base := strings.TrimRight(r.cfg.BaseURL, "/")
	suffix := "/" + strings.TrimLeft(path, "/")
	return base + suffix
}

// normalizeConfig applies defaults and validates required IOT settings.
func normalizeConfig(cfg Config) (Config, error) {
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultIOTBaseURL
	}
	parsedBaseURL, err := url.ParseRequestURI(cfg.BaseURL)
	if err != nil {
		return Config{}, gerror.Wrap(err, "sicau-niu niu.baseUrl is invalid")
	}
	if parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" {
		return Config{}, gerror.New("sicau-niu niu.baseUrl is invalid")
	}
	cfg.Key = strings.TrimSpace(cfg.Key)
	cfg.Secret = strings.TrimSpace(cfg.Secret)
	if cfg.Key == "" || cfg.Secret == "" {
		return Config{}, gerror.New("sicau-niu IOT locator key and secret are required")
	}
	if cfg.PageSize <= 0 {
		cfg.PageSize = defaultPageSize
	}
	if cfg.MaxPages <= 0 {
		cfg.MaxPages = defaultMaxPages
	}
	if cfg.TokenTTL <= tokenRefreshBuffer {
		cfg.TokenTTL = defaultTokenTTL
	}
	return cfg, nil
}

// tokenRequest is the documented getToken request body.
type tokenRequest struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

// tokenResponse is the documented getToken response body.
type tokenResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Data string `json:"data"`
	} `json:"data"`
}

// locatorListRequest is the documented locator list request body.
type locatorListRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Sign     string `json:"sign"`
}

// locatorListResponse is the documented locator list response body.
type locatorListResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Count int           `json:"count"`
		Data  []locatorItem `json:"data"`
	} `json:"data"`
}

// locatorItem is one IOT locator projection returned by getInfoList. The
// platform also returns raw GPS lat/lon fields, but they are WGS-84 and are
// deliberately not decoded: only the Gaode latitude/longitude (GCJ-02) fields
// may enter the iron table, matching the plugin-wide GCJ-02 coordinate system.
type locatorItem struct {
	LocatorNo      string `json:"locatorNo"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	LastLocateTime string `json:"lastLocateTime"`
}

// locatorLocation is the normalized local coordinate projection.
type locatorLocation struct {
	Code string
	Lat  float64
	Lng  float64
}

// location normalizes a locator item into a usable coordinate. Only the
// documented Gaode latitude/longitude fields (GCJ-02) are accepted; when they
// are missing or invalid the item is skipped and the iron cow keeps its last
// synced position, so a WGS-84 raw GPS value can never mix into the GCJ-02
// iron table and silently break the ~12m proximity-bonus threshold.
func (i locatorItem) location() (locatorLocation, bool) {
	code := strings.TrimSpace(i.LocatorNo)
	if code == "" {
		return locatorLocation{}, false
	}

	lat, latOK := parseCoordinate(i.Latitude, -90, 90)
	lng, lngOK := parseCoordinate(i.Longitude, -180, 180)
	if !latOK || !lngOK {
		return locatorLocation{}, false
	}

	return locatorLocation{
		Code: code,
		Lat:  lat,
		Lng:  lng,
	}, true
}

// parseCoordinate parses and bounds one decimal-degree coordinate.
func parseCoordinate(raw string, min float64, max float64) (float64, bool) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < min || value > max {
		return 0, false
	}
	return value, true
}

type ironLocationUpdate struct {
	ID  int64
	Lat float64
	Lng float64
}

type caseExpr struct {
	matchColumn string
	elseColumn  string
	clauses     []string
}

func newCaseExpr(matchColumn string, elseColumn string) *caseExpr {
	return &caseExpr{
		matchColumn: matchColumn,
		elseColumn:  elseColumn,
		clauses:     make([]string, 0),
	}
}

func (e *caseExpr) Add(id int64, value float64) {
	e.clauses = append(e.clauses, fmt.Sprintf(
		"WHEN %d THEN %s",
		id,
		strconv.FormatFloat(value, 'f', -1, 64),
	))
}

func (e *caseExpr) String() string {
	if len(e.clauses) == 0 {
		return e.elseColumn
	}
	return "CASE " + e.matchColumn + " " + strings.Join(e.clauses, " ") + " ELSE " + e.elseColumn + " END"
}
