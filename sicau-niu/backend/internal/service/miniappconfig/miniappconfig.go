// Package miniappconfig owns the non-sensitive runtime projection consumed by
// the sicau-niu WeChat mini program. The plugin database is authoritative when
// an operator has saved a row; live gameplay rules are projected in one bounded
// read and startup values are used only as defaults.
package miniappconfig

import (
	"context"
	"encoding/json"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

const (
	// CampusChengdu is the stable Chengdu campus ID.
	CampusChengdu = "cd"
	// CampusDujiangyan is the stable Dujiangyan campus ID.
	CampusDujiangyan = "djy"
	// CampusYaan is the stable Yaan campus ID.
	CampusYaan = "ya"

	// ActivityPhasePreview marks a not-yet-active public activity.
	ActivityPhasePreview = "preview"
	// ActivityPhaseActive marks the live public activity.
	ActivityPhaseActive = "active"
	// ActivityPhaseClosed marks a completed public activity.
	ActivityPhaseClosed = "closed"

	// configKey identifies the single operator-maintained configuration row.
	configKey = "default"
)

var (
	// miniProgramLocation is the fixed UTC+8 activity timezone.
	miniProgramLocation = time.FixedZone("Asia/Shanghai", 8*60*60)
	// assetsVersionRE constrains safe cache-busting version strings.
	assetsVersionRE = regexp.MustCompile(`^[A-Za-z0-9._-]{1,32}$`)
)

// Config contains startup defaults that vary by deployment.
type Config struct {
	// DefaultCampus is the startup default campus ID.
	DefaultCampus string
	// Anniversary is the startup anniversary display name.
	Anniversary string
	// AnniversaryAt is the date-only anniversary in YYYY-MM-DD form.
	AnniversaryAt string
	// Debug controls development-only helpers exposed in the public projection.
	Debug bool
	// ActivateRadiusMeters is the fallback activation radius.
	ActivateRadiusMeters int
	// AssetsVersion is the static resource cache-busting version.
	AssetsVersion string
	// StaticAssetBaseURL is the optional HTTPS resource base URL.
	StaticAssetBaseURL string
	// ActivityPhase is preview, active or closed.
	ActivityPhase string
}

// Service exposes public reads and operator replacement of mini-program config.
type Service interface {
	// Snapshot returns the operator row overlaid on startup defaults and live
	// public gameplay rules. Store and config failures return config bizerrs.
	Snapshot(ctx context.Context) (*Snapshot, error)
	// Update validates and replaces the complete operator-maintained projection,
	// then returns the normalized live snapshot. Invalid input is not persisted.
	Update(ctx context.Context, in *Snapshot) (*Snapshot, error)
}

// serviceImpl implements Service with one database row and optional live rules.
type serviceImpl struct {
	// defaults is the immutable normalized startup projection.
	defaults Snapshot
	// rulesSvc supplies the live public gameplay rules when configured.
	rulesSvc rulessvc.Service
}

// Interface compliance assertion for the default mini-program config service.
var _ Service = (*serviceImpl)(nil)

// Snapshot is the complete public mini-program configuration.
type Snapshot struct {
	// Campuses contains exactly the three supported campus calibrations.
	Campuses []*Campus
	// DefaultCampus is one of cd, djy or ya.
	DefaultCampus string
	// ActivateRadiusM is the live activation radius in meters.
	ActivateRadiusM int
	// StealDailyLimit is the live per-player daily steal action cap.
	StealDailyLimit int
	// GiftDailyLimit is the live per-player daily gift action cap.
	GiftDailyLimit int
	// GiftMinAmount is the live minimum grass amount for one gift.
	GiftMinAmount int
	// CountdownDays is the non-negative whole-day anniversary countdown.
	CountdownDays int
	// Anniversary is the public anniversary display name.
	Anniversary string
	// AnniversaryAt is the date-only anniversary in YYYY-MM-DD form.
	AnniversaryAt string
	// Debug reports whether development helpers are enabled by configuration.
	Debug bool
	// AssetsVersion is the static resource cache-busting version.
	AssetsVersion string
	// StaticAssetBaseURL is an optional HTTPS asset base URL.
	StaticAssetBaseURL string
	// ActivityPhase is preview, active or closed.
	ActivityPhase string
}

// Campus describes one supported campus and its affine map calibration.
type Campus struct {
	// Id is the stable campus ID.
	Id string `json:"id"`
	// Name is the campus display name.
	Name string `json:"name"`
	// MapImage is an optional relative path or HTTPS image URL.
	MapImage string `json:"mapImage,omitempty"`
	// Calibration maps GCJ-02 coordinates to this campus image.
	Calibration Calibration `json:"calibration"`
}

// Calibration maps GCJ-02 coordinates to the mini-program map canvas.
type Calibration struct {
	// Version is the calibration schema version.
	Version int `json:"version"`
	// Origin is the GCJ-02 reference coordinate.
	Origin Origin `json:"origin"`
	// ImageSize is the source map size in pixels.
	ImageSize ImageSize `json:"imageSize"`
	// Affine is the coordinate-to-pixel affine transform.
	Affine Affine `json:"affine"`
}

// Origin is the GCJ-02 calibration reference coordinate.
type Origin struct {
	// Lat0 is the reference latitude.
	Lat0 float64 `json:"lat0"`
	// Lng0 is the reference longitude.
	Lng0 float64 `json:"lng0"`
}

// ImageSize is the source map size in pixels.
type ImageSize struct {
	// W is the source width in pixels.
	W int `json:"w"`
	// H is the source height in pixels.
	H int `json:"h"`
}

// Affine is the six-coefficient coordinate-to-pixel transform.
type Affine struct {
	// A is the longitude scale/rotation coefficient.
	A float64 `json:"a"`
	// B is the latitude cross-axis coefficient.
	B float64 `json:"b"`
	// C is the longitude cross-axis coefficient.
	C float64 `json:"c"`
	// D is the latitude scale/rotation coefficient.
	D float64 `json:"d"`
	// Tx is the horizontal pixel translation.
	Tx float64 `json:"tx"`
	// Ty is the vertical pixel translation.
	Ty float64 `json:"ty"`
}

// Area is the stable fuzzy region returned for an unactivated cattle.
type Area struct {
	// Name is the stable fuzzy-region display name.
	Name string
	// Lat is the deterministically offset GCJ-02 center latitude.
	Lat float64
	// Lng is the deterministically offset GCJ-02 center longitude.
	Lng float64
	// RadiusM is the fuzzy circle radius in meters.
	RadiusM int
}

// New creates the database-backed public configuration service.
func New(rulesSvc rulessvc.Service, config Config) Service {
	defaults := Snapshot{
		Campuses: []*Campus{
			// Chengdu campus (211 Huimin Road, Wenjiang). The origin is the GCJ-02 campus
			// center; the anchor is the centroid of the campus artwork, which is not centered
			// in its own image. The scale makes the drawn campus span roughly 920m x 430m.
			// The artwork is an oblique illustration, so this affine is an approximation:
			// replace it with a control-point solve once field survey data is available.
			anchoredCampus(CampusChengdu, "成都校区", 30.7054, 103.8632, 0.068, 16644, 8241, 5654, 2931),
			campus(CampusDujiangyan, "都江堰校区", 31.001, 103.63, 1.2, 991, 749),
			campus(CampusYaan, "雅安校区", 29.9855, 102.9988, 1.15, 1200, 900),
		},
		DefaultCampus:      normalizeCampus(config.DefaultCampus),
		ActivateRadiusM:    config.ActivateRadiusMeters,
		StealDailyLimit:    5,
		GiftDailyLimit:     12,
		GiftMinAmount:      12,
		Anniversary:        strings.TrimSpace(config.Anniversary),
		AnniversaryAt:      strings.TrimSpace(config.AnniversaryAt),
		Debug:              config.Debug,
		AssetsVersion:      strings.TrimSpace(config.AssetsVersion),
		StaticAssetBaseURL: strings.TrimSpace(config.StaticAssetBaseURL),
		ActivityPhase:      strings.TrimSpace(config.ActivityPhase),
	}
	if defaults.Anniversary == "" {
		defaults.Anniversary = "120 周年校庆"
	}
	if defaults.ActivateRadiusM <= 0 {
		defaults.ActivateRadiusM = 50
	}
	if defaults.AssetsVersion == "" {
		defaults.AssetsVersion = "1"
	}
	if !validActivityPhase(defaults.ActivityPhase) {
		defaults.ActivityPhase = ActivityPhaseActive
	}
	return &serviceImpl{defaults: defaults, rulesSvc: rulesSvc}
}

// Snapshot reads the one operator row and resolves all public live rules in one
// bounded rule query.
func (s *serviceImpl) Snapshot(ctx context.Context) (*Snapshot, error) {
	out := cloneSnapshot(&s.defaults)
	var row *entitymodel.MiniappConfig
	if err := dao.MiniappConfig.Ctx(ctx).Where(dao.MiniappConfig.Columns().ConfigKey, configKey).Scan(&row); err != nil {
		return nil, bizerr.WrapCode(err, CodeConfigQueryFailed)
	}
	if row != nil {
		campuses := make([]*Campus, 0, 3)
		if err := json.Unmarshal([]byte(row.CampusesJson), &campuses); err != nil {
			return nil, bizerr.WrapCode(err, CodeConfigInvalid)
		}
		out.Campuses = campuses
		out.DefaultCampus = row.DefaultCampus
		out.Anniversary = row.Anniversary
		out.AnniversaryAt = row.AnniversaryAt
		out.Debug = row.Debug == 1
		out.AssetsVersion = row.AssetsVersion
		out.StaticAssetBaseURL = row.StaticAssetBaseUrl
		out.ActivityPhase = row.ActivityPhase
	}
	if s.rulesSvc != nil {
		rules, err := s.rulesSvc.Rules(ctx)
		if err != nil {
			return nil, err
		}
		applyRuntimeRules(out, rules)
	}
	out.CountdownDays = countdownDays(out.AnniversaryAt, time.Now())
	return out, nil
}

// Update validates and replaces the complete operator-maintained projection.
func (s *serviceImpl) Update(ctx context.Context, in *Snapshot) (*Snapshot, error) {
	normalized, err := normalizeSnapshot(in)
	if err != nil {
		return nil, err
	}
	campusesJSON, err := json.Marshal(normalized.Campuses)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeConfigInvalid)
	}
	data := do.MiniappConfig{
		AssetsVersion:      normalized.AssetsVersion,
		StaticAssetBaseUrl: normalized.StaticAssetBaseURL,
		ActivityPhase:      normalized.ActivityPhase,
		DefaultCampus:      normalized.DefaultCampus,
		Anniversary:        normalized.Anniversary,
		AnniversaryAt:      normalized.AnniversaryAt,
		Debug:              boolInt(normalized.Debug),
		CampusesJson:       string(campusesJSON),
	}
	count, err := dao.MiniappConfig.Ctx(ctx).Where(dao.MiniappConfig.Columns().ConfigKey, configKey).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeConfigQueryFailed)
	}
	if count > 0 {
		_, err = dao.MiniappConfig.Ctx(ctx).Where(dao.MiniappConfig.Columns().ConfigKey, configKey).Data(data).Update()
	} else {
		data.ConfigKey = configKey
		_, err = dao.MiniappConfig.Ctx(ctx).Data(data).Insert()
	}
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeConfigWriteFailed)
	}
	return s.Snapshot(ctx)
}

// CampusID assigns a coordinate to its nearest supported campus center.
func CampusID(lat, lng float64) string {
	centers := []struct {
		id       string
		lat, lng float64
	}{{CampusChengdu, 30.7058, 103.8318}, {CampusDujiangyan, 31.001, 103.63}, {CampusYaan, 29.9855, 102.9988}}
	bestID := CampusChengdu
	bestDistance := math.MaxFloat64
	for _, center := range centers {
		distance := squaredDistance(lat, lng, center.lat, center.lng)
		if distance < bestDistance {
			bestID, bestDistance = center.id, distance
		}
	}
	return bestID
}

// FuzzyArea deterministically offsets an inactive cattle anchor.
func FuzzyArea(niuID int64, lat, lng float64) *Area {
	angle := float64((niuID*7919)%360) * math.Pi / 180
	offsetMeters := 80 + float64((niuID*37)%120)
	dLat := offsetMeters * math.Cos(angle) / 110540
	cosLat := math.Cos(lat * math.Pi / 180)
	if math.Abs(cosLat) < 0.01 {
		cosLat = 0.01
	}
	dLng := offsetMeters * math.Sin(angle) / (111320 * cosLat)
	return &Area{Name: campusName(CampusID(lat, lng)) + "神秘区域", Lat: lat + dLat, Lng: lng + dLng, RadiusM: 260}
}

// normalizeSnapshot validates and clones a complete operator update.
func normalizeSnapshot(in *Snapshot) (*Snapshot, error) {
	if in == nil || len(in.Campuses) != 3 || !assetsVersionRE.MatchString(strings.TrimSpace(in.AssetsVersion)) || !validActivityPhase(in.ActivityPhase) {
		return nil, bizerr.NewCode(CodeConfigInvalid)
	}
	out := cloneSnapshot(in)
	out.DefaultCampus = strings.TrimSpace(out.DefaultCampus)
	out.AssetsVersion = strings.TrimSpace(out.AssetsVersion)
	out.StaticAssetBaseURL = strings.TrimRight(strings.TrimSpace(out.StaticAssetBaseURL), "/")
	out.ActivityPhase = strings.TrimSpace(out.ActivityPhase)
	out.Anniversary = strings.TrimSpace(out.Anniversary)
	out.AnniversaryAt = strings.TrimSpace(out.AnniversaryAt)
	if out.Anniversary == "" || !validStaticURL(out.StaticAssetBaseURL) {
		return nil, bizerr.NewCode(CodeConfigInvalid)
	}
	if out.AnniversaryAt != "" {
		if _, err := time.ParseInLocation("2006-01-02", out.AnniversaryAt, miniProgramLocation); err != nil {
			return nil, bizerr.NewCode(CodeConfigInvalid)
		}
	}
	seen := make(map[string]bool, 3)
	for _, item := range out.Campuses {
		if item == nil {
			return nil, bizerr.NewCode(CodeConfigInvalid)
		}
		item.Id = strings.TrimSpace(item.Id)
		item.Name = strings.TrimSpace(item.Name)
		item.MapImage = strings.TrimSpace(item.MapImage)
		cal := item.Calibration
		if !validCampus(item.Id) || seen[item.Id] || item.Name == "" || !validMapImage(item.MapImage) || cal.Version <= 0 || !validCoordinate(cal.Origin.Lat0, cal.Origin.Lng0) || cal.ImageSize.W <= 0 || cal.ImageSize.H <= 0 || math.Abs(cal.Affine.A*cal.Affine.D-cal.Affine.B*cal.Affine.C) < 1e-12 {
			return nil, bizerr.NewCode(CodeConfigInvalid)
		}
		seen[item.Id] = true
	}
	if !seen[out.DefaultCampus] {
		return nil, bizerr.NewCode(CodeConfigInvalid)
	}
	return out, nil
}

// campus builds one default north-up affine calibration anchored at the image center.
func campus(id, name string, lat0, lng0, metersPerPixel float64, width, height int) *Campus {
	return anchoredCampus(id, name, lat0, lng0, metersPerPixel, width, height, float64(width)/2, float64(height)/2)
}

// anchoredCampus builds one default north-up affine calibration whose origin lands on an
// arbitrary image pixel. Use it when the campus artwork is not centered in its own image.
func anchoredCampus(id, name string, lat0, lng0, metersPerPixel float64, width, height int, anchorX, anchorY float64) *Campus {
	scale := 1 / metersPerPixel
	return &Campus{Id: id, Name: name, Calibration: Calibration{Version: 1, Origin: Origin{Lat0: lat0, Lng0: lng0}, ImageSize: ImageSize{W: width, H: height}, Affine: Affine{A: scale, D: -scale, Tx: anchorX, Ty: anchorY}}}
}

// cloneSnapshot deep-copies a snapshot and its campus slice.
func cloneSnapshot(in *Snapshot) *Snapshot {
	out := *in
	out.Campuses = make([]*Campus, 0, len(in.Campuses))
	for _, item := range in.Campuses {
		if item == nil {
			out.Campuses = append(out.Campuses, nil)
			continue
		}
		copyItem := *item
		out.Campuses = append(out.Campuses, &copyItem)
	}
	return &out
}

// countdownDays returns non-negative whole UTC+8 days to a date-only target.
func countdownDays(raw string, now time.Time) int {
	target, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(raw), miniProgramLocation)
	if err != nil {
		return 0
	}
	days := int(math.Ceil(target.Sub(now.In(miniProgramLocation)).Hours() / 24))
	if days < 0 {
		return 0
	}
	return days
}

// normalizeCampus returns a supported campus ID or the Chengdu default.
func normalizeCampus(value string) string {
	if validCampus(strings.TrimSpace(value)) {
		return strings.TrimSpace(value)
	}
	return CampusChengdu
}

// validCampus reports whether value is one supported campus ID.
func validCampus(value string) bool {
	return value == CampusChengdu || value == CampusDujiangyan || value == CampusYaan
}

// campusName returns the Chinese display name for one stable campus ID.
func campusName(id string) string {
	if id == CampusDujiangyan {
		return "都江堰校区"
	}
	if id == CampusYaan {
		return "雅安校区"
	}
	return "成都校区"
}

// validCoordinate reports whether latitude and longitude are finite and in range.
func validCoordinate(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180 && !(lat == 0 && lng == 0)
}

// validActivityPhase reports whether value is preview, active or closed.
func validActivityPhase(value string) bool {
	value = strings.TrimSpace(value)
	return value == ActivityPhasePreview || value == ActivityPhaseActive || value == ActivityPhaseClosed
}

// validStaticURL accepts an empty value or a safe HTTPS asset base URL.
func validStaticURL(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

// validMapImage accepts an empty value, safe relative path or HTTPS image URL.
func validMapImage(value string) bool {
	return value == "" || strings.HasPrefix(value, "/") || validStaticURL(value)
}

// squaredDistance returns a comparison-only planar coordinate distance.
func squaredDistance(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := lat1 - lat2
	dLng := (lng1 - lng2) * math.Cos(lat2*math.Pi/180)
	return dLat*dLat + dLng*dLng
}

// boolInt maps a boolean to the database's 0/1 representation.
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
