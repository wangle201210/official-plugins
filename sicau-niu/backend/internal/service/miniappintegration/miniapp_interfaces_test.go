package miniappintegration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/dialect"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	activationsvc "lina-plugin-sicau-niu/backend/internal/service/activation"
	activationphotosvc "lina-plugin-sicau-niu/backend/internal/service/activationphoto"
	miniappconfigsvc "lina-plugin-sicau-niu/backend/internal/service/miniappconfig"
)

var (
	dbOnce           sync.Once
	dbLink           string
	dbPrepErr        error
	dbOriginalConfig gdb.Config
)

var schemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"002-sicau-niu-catalog.sql",
	"003-sicau-niu-activation.sql",
	"004-sicau-niu-grass.sql",
	"007-sicau-niu-rule-config.sql",
	"010-sicau-niu-miniapp-interfaces.sql",
}

var tables = []string{
	"plugin_sicau_niu_transport_report",
	"plugin_sicau_niu_transport_member",
	"plugin_sicau_niu_transport_team",
	"plugin_sicau_niu_miniapp_config",
	"plugin_sicau_niu_activation_photo",
	"plugin_sicau_niu_activation_attempt",
	"plugin_sicau_niu_activation",
	"plugin_sicau_niu_checkin",
	"plugin_sicau_niu_grass_txn",
	"plugin_sicau_niu_card",
	"plugin_sicau_niu_niu",
	"plugin_sicau_niu_iron",
	"plugin_sicau_niu_user",
}

func setupDB(t *testing.T, ctx context.Context) {
	t.Helper()
	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run mini-program interface integration tests")
	}
	dbOnce.Do(func() { dbPrepErr = provisionDB(ctx, baseLink) })
	if dbPrepErr != nil {
		t.Fatalf("provision mini-program interface database failed: %v", dbPrepErr)
	}
	for _, table := range tables {
		if _, err := g.DB().Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

func provisionDB(ctx context.Context, baseLink string) error {
	link, err := uniqueDBLink(baseLink)
	if err != nil {
		return err
	}
	dbDialect, err := dialect.From(link)
	if err != nil {
		return err
	}
	if err = dbDialect.PrepareDatabase(ctx, link, true); err != nil {
		return err
	}
	original := gdb.GetAllConfig()
	if err = gdb.SetConfig(gdb.Config{gdb.DefaultGroupName: gdb.ConfigGroup{{Link: link}}}); err != nil {
		return err
	}
	for _, name := range schemaFiles {
		path := filepath.Join("..", "..", "..", "..", "manifest", "sql", name)
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		translated, translateErr := dbDialect.TranslateDDL(ctx, path, string(content))
		if translateErr != nil {
			return translateErr
		}
		for _, statement := range dialect.SplitSQLStatements(translated) {
			if _, execErr := g.DB().Exec(ctx, statement); execErr != nil {
				return fmt.Errorf("execute schema SQL failed: %w\nSQL:\n%s", execErr, statement)
			}
		}
	}
	dbLink = link
	dbOriginalConfig = original
	return nil
}

func uniqueDBLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", errors.New("PostgreSQL base link configuration is empty")
	}
	if err = db.Close(context.Background()); err != nil {
		return "", err
	}
	return fmt.Sprintf("pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_miniapp_%d%s", config.User, config.Pass, config.Protocol, config.Host, config.Port, time.Now().UnixNano(), normalizeExtra(config.Extra)), nil
}

func teardownDB() {
	if dbLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close mini-program interface database failed: %v\n", err)
	}
	if err := gdb.SetConfig(dbOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropDB(ctx, dbLink); err != nil {
		fmt.Printf("drop mini-program interface database failed: %v\n", err)
	}
}

func dropDB(ctx context.Context, targetLink string) (err error) {
	target, err := gdb.New(gdb.ConfigNode{Link: targetLink})
	if err != nil {
		return err
	}
	config := target.GetConfig()
	if config == nil {
		return target.Close(ctx)
	}
	name := strings.TrimSpace(config.Name)
	if err = target.Close(ctx); err != nil || name == "" {
		return err
	}
	system, err := gdb.New(gdb.ConfigNode{Link: fmt.Sprintf("pgsql:%s:%s@%s(%s:%s)/postgres%s", config.User, config.Pass, config.Protocol, config.Host, config.Port, normalizeExtra(config.Extra))})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := system.Close(ctx); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if _, err = system.Exec(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()", name); err != nil {
		return err
	}
	quoted := `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	_, err = system.Exec(ctx, "DROP DATABASE IF EXISTS "+quoted)
	return err
}

func normalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}

func TestMain(m *testing.M) {
	code := m.Run()
	teardownDB()
	os.Exit(code)
}

type memoryStorage struct {
	storagecap.Service
	mu      sync.Mutex
	objects map[string][]byte
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{objects: make(map[string][]byte)}
}

func (s *memoryStorage) Put(_ context.Context, in storagecap.PutInput) (*storagecap.PutOutput, error) {
	content, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[in.Path] = append([]byte(nil), content...)
	return &storagecap.PutOutput{Object: &storagecap.Object{Path: in.Path, Size: int64(len(content)), ContentType: in.ContentType}}, nil
}

func (s *memoryStorage) Get(_ context.Context, in storagecap.GetInput) (*storagecap.GetOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	content, ok := s.objects[in.Path]
	if !ok {
		return &storagecap.GetOutput{Found: false}, nil
	}
	copyBytes := append([]byte(nil), content...)
	return &storagecap.GetOutput{Found: true, Object: &storagecap.Object{Path: in.Path, Size: int64(len(copyBytes))}, Body: io.NopCloser(bytes.NewReader(copyBytes))}, nil
}

func (s *memoryStorage) Delete(_ context.Context, in storagecap.DeleteInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, in.Path)
	return nil
}

func TestActivationPhotoOwnerAndOneTimeConsumption(t *testing.T) {
	ctx := context.Background()
	setupDB(t, ctx)
	owner := insertUser(t, ctx, "openid-photo-owner")
	other := insertUser(t, ctx, "openid-photo-other")
	svc, err := activationphotosvc.New(newMemoryStorage())
	if err != nil {
		t.Fatalf("create photo service failed: %v", err)
	}
	pngBytes := testPNG(t)
	uploaded, err := svc.Upload(ctx, owner, &activationphotosvc.UploadInput{RequestID: "photo-owner-1", Filename: "checkin.png", ContentType: "image/png", SizeBytes: int64(len(pngBytes)), Reader: bytes.NewReader(pngBytes)})
	if err != nil {
		t.Fatalf("upload photo failed: %v", err)
	}
	if uploaded.ContentType != "image/webp" || uploaded.SizeBytes <= 0 || uploaded.SizeBytes > 300*1024 {
		t.Fatalf("photo was not standardized: %+v", uploaded)
	}
	replayed, err := svc.Upload(ctx, owner, &activationphotosvc.UploadInput{RequestID: "photo-owner-1", Filename: "retry.png", ContentType: "image/png", SizeBytes: int64(len(pngBytes)), Reader: bytes.NewReader(pngBytes)})
	if err != nil || replayed.Token != uploaded.Token {
		t.Fatalf("photo idempotency replay failed: first=%+v replay=%+v err=%v", uploaded, replayed, err)
	}
	content, err := svc.Content(ctx, owner, uploaded.Token)
	if err != nil || content.ContentType != "image/webp" || len(content.Bytes) < 12 || string(content.Bytes[:4]) != "RIFF" || string(content.Bytes[8:12]) != "WEBP" {
		t.Fatalf("owner read failed: content=%+v err=%v", content, err)
	}
	_, err = svc.Content(ctx, other, uploaded.Token)
	assertCode(t, err, activationphotosvc.CodePhotoNotFound.RuntimeCode())
	if err = svc.Consume(ctx, owner, uploaded.Token, time.Now()); err != nil {
		t.Fatalf("consume photo failed: %v", err)
	}
	err = svc.Consume(ctx, owner, uploaded.Token, time.Now())
	assertCode(t, err, activationphotosvc.CodePhotoAlreadyUsed.RuntimeCode())
}

func TestActivationPhotoServiceRejectsMissingStorage(t *testing.T) {
	if _, err := activationphotosvc.New(nil); err == nil {
		t.Fatal("expected photo service construction to reject missing storage")
	}
}

func TestActivationPhotoDailyQuota(t *testing.T) {
	ctx := context.Background()
	setupDB(t, ctx)
	owner := insertUser(t, ctx, "openid-photo-quota")
	svc, err := activationphotosvc.New(newMemoryStorage())
	if err != nil {
		t.Fatalf("create photo service failed: %v", err)
	}
	pngBytes := testPNG(t)
	for i := 1; i <= 10; i++ {
		requestID := fmt.Sprintf("photo-quota-%d", i)
		if _, err := svc.Upload(ctx, owner, &activationphotosvc.UploadInput{
			RequestID: requestID, Filename: "quota.png", ContentType: "image/png",
			SizeBytes: int64(len(pngBytes)), Reader: bytes.NewReader(pngBytes),
		}); err != nil {
			t.Fatalf("upload %d within quota failed: %v", i, err)
		}
	}
	_, err = svc.Upload(ctx, owner, &activationphotosvc.UploadInput{
		RequestID: "photo-quota-11", Filename: "quota.png", ContentType: "image/png",
		SizeBytes: int64(len(pngBytes)), Reader: bytes.NewReader(pngBytes),
	})
	assertCode(t, err, activationphotosvc.CodePhotoDailyLimit.RuntimeCode())
}

func TestMiniappConfigPersistsOperatorUpdate(t *testing.T) {
	ctx := context.Background()
	setupDB(t, ctx)
	svc := miniappconfigsvc.New(nil, miniappconfigsvc.Config{DefaultCampus: "cd", AnniversaryAt: "2026-10-06"})
	config, err := svc.Snapshot(ctx)
	if err != nil {
		t.Fatalf("read default miniapp config failed: %v", err)
	}
	config.AssetsVersion = "2026.08.11"
	config.StaticAssetBaseURL = "https://assets.example.com/sicau-niu"
	config.ActivityPhase = miniappconfigsvc.ActivityPhasePreview
	config.Debug = true
	config.Campuses[0].MapImage = "/maps/cd.jpg"
	updated, err := svc.Update(ctx, config)
	if err != nil {
		t.Fatalf("update miniapp config failed: %v", err)
	}
	if updated.AssetsVersion != config.AssetsVersion || updated.ActivityPhase != config.ActivityPhase || !updated.Debug {
		t.Fatalf("unexpected updated config: %+v", updated)
	}
	reloaded, err := miniappconfigsvc.New(nil, miniappconfigsvc.Config{}).Snapshot(ctx)
	if err != nil {
		t.Fatalf("reload miniapp config failed: %v", err)
	}
	if reloaded.StaticAssetBaseURL != config.StaticAssetBaseURL || reloaded.Campuses[0].MapImage != "/maps/cd.jpg" {
		t.Fatalf("operator config was not persisted: %+v", reloaded)
	}
}

func TestActivationConsumesPhotoOnlyOnSuccess(t *testing.T) {
	ctx := context.Background()
	setupDB(t, ctx)
	owner := insertUser(t, ctx, "openid-activation-photo-owner")
	retryPlayer := insertUser(t, ctx, "openid-activation-photo-retry")
	past := time.Now().Add(-time.Hour)
	if _, err := dao.Niu.Ctx(ctx).Data(do.Niu{Code: "NIU-PHOTO", NiuType: "common", Lat: 30.7058, Lng: 103.8318, OnlineAt: &past, Status: "inactive"}).Insert(); err != nil {
		t.Fatalf("insert niu failed: %v", err)
	}
	storage := newMemoryStorage()
	photoSvc, err := activationphotosvc.New(storage)
	if err != nil {
		t.Fatalf("create photo service failed: %v", err)
	}
	pngBytes := testPNG(t)
	successPhoto, err := photoSvc.Upload(ctx, owner, &activationphotosvc.UploadInput{RequestID: "photo-activation-success", Filename: "success.png", ContentType: "image/png", SizeBytes: int64(len(pngBytes)), Reader: bytes.NewReader(pngBytes)})
	if err != nil {
		t.Fatalf("upload success photo failed: %v", err)
	}
	activationSvc := activationsvc.New(nil, activationsvc.NewBasicPosterRenderer(), nil, photoSvc, activationsvc.Config{LBSThresholdMeters: 50})
	if _, err = activationSvc.Activate(ctx, owner, &activationsvc.ActivateInput{RequestID: "activation-success", Lat: 30.7058, Lng: 103.8318, PhotoPath: successPhoto.Token}); err != nil {
		t.Fatalf("activation with owned photo failed: %v", err)
	}
	assertCode(t, photoSvc.Validate(ctx, owner, successPhoto.Token), activationphotosvc.CodePhotoAlreadyUsed.RuntimeCode())

	retryPhoto, err := photoSvc.Upload(ctx, retryPlayer, &activationphotosvc.UploadInput{RequestID: "photo-activation-retry", Filename: "retry.png", ContentType: "image/png", SizeBytes: int64(len(pngBytes)), Reader: bytes.NewReader(pngBytes)})
	if err != nil {
		t.Fatalf("upload retry photo failed: %v", err)
	}
	_, err = activationSvc.Activate(ctx, retryPlayer, &activationsvc.ActivateInput{RequestID: "activation-retry", Lat: 31.0, Lng: 104.0, PhotoPath: retryPhoto.Token})
	if err == nil {
		t.Fatal("expected out-of-range activation failure")
	}
	if err = photoSvc.Validate(ctx, retryPlayer, retryPhoto.Token); err != nil {
		t.Fatalf("failed activation consumed reusable photo: %v", err)
	}
}

func testPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 96, 96))
	for y := 0; y < 96; y++ {
		for x := 0; x < 96; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 2), G: uint8(y * 2), B: uint8(x + y), A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatalf("encode test PNG failed: %v", err)
	}
	return encoded.Bytes()
}

func insertUser(t *testing.T, ctx context.Context, openid string) int64 {
	t.Helper()
	id, err := dao.User.Ctx(ctx).Data(do.User{Openid: openid, Nickname: openid}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert user failed: %v", err)
	}
	return id
}

func assertCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected business error %s", want)
	}
	parsed, ok := bizerr.As(err)
	if !ok || parsed.RuntimeCode() != want {
		t.Fatalf("expected business error %s, got %v", want, err)
	}
}
