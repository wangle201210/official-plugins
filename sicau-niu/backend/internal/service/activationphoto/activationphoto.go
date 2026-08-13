// Package activationphoto validates and standardizes private activation photos,
// stores them through plugin-scoped object storage, and enforces owner, daily
// quota and one-time-consumption constraints in plugin tables.
package activationphoto

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/heic"
	"github.com/gen2brain/webp"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/google/uuid"
	"golang.org/x/image/draw"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	"lina-plugin-sicau-niu/backend/internal/requestid"
)

const (
	// maxInputPhotoBytes bounds the uploaded source image to 5 MiB.
	maxInputPhotoBytes int64 = 5 * 1024 * 1024
	// maxPhotoBytes bounds the standardized WebP object to 300 KiB.
	maxPhotoBytes int64 = 300 * 1024
	// maxDailyPhotos is the per-player Beijing-day successful upload quota.
	maxDailyPhotos = 10
	// maxDecodedPixels rejects decompression-bomb-sized source images.
	maxDecodedPixels = 40_000_000
)

// Service owns photo validation, private storage, owner reads and one-time use.
type Service interface {
	// Upload validates and standardizes one player-owned image, enforces the
	// Beijing-day quota and request idempotency, and returns the stable private
	// photo token. It returns a photo bizerr for invalid input, quota exhaustion,
	// storage failure or database failure.
	Upload(ctx context.Context, playerID int64, in *UploadInput) (*Photo, error)
	// Content returns the private standardized image only when playerID owns the
	// token. Foreign and unknown tokens share CodePhotoNotFound.
	Content(ctx context.Context, playerID int64, token string) (*Content, error)
	// ContentForAudit returns one private standardized image to an already
	// authorized operator path. Unknown tokens return CodePhotoNotFound.
	ContentForAudit(ctx context.Context, token string) (*Content, error)
	// ListAudit returns one bounded operator audit page with batch-assembled player
	// and cattle fields. Invalid pagination is normalized by the service.
	ListAudit(ctx context.Context, in *AuditListInput) (*AuditListOutput, error)
	// Validate checks that token belongs to playerID and is unused without
	// consuming it. Foreign, unknown and consumed tokens return photo bizerrs.
	Validate(ctx context.Context, playerID int64, token string) error
	// Consume atomically marks one validated player-owned token as used at usedAt.
	// Reuse and ownership failures return photo bizerrs.
	Consume(ctx context.Context, playerID int64, token string, usedAt time.Time) error
}

// UploadInput carries one multipart image and its idempotency metadata.
type UploadInput struct {
	// RequestID is the required player-scoped idempotency key.
	RequestID string
	// Filename is the original client filename used only for audit metadata.
	Filename string
	// ContentType is the client-declared media type validated with decoded input.
	ContentType string
	// SizeBytes is the declared input size and must not exceed 5 MiB.
	SizeBytes int64
	// Reader supplies the image bytes and is consumed once.
	Reader io.Reader
}

// Photo is the stable private token and standardized image metadata.
type Photo struct {
	// Token is the unguessable player-facing photo identifier.
	Token string
	// ContentType is always image/webp for a successful upload.
	ContentType string
	// SizeBytes is the standardized object size and is at most 300 KiB.
	SizeBytes int64
}

// Content is the bounded standardized image returned from private storage.
type Content struct {
	// ContentType is the stored media type.
	ContentType string
	// SizeBytes is the returned byte count.
	SizeBytes int64
	// Bytes contains the standardized image payload.
	Bytes []byte
}

// AuditListInput defines the protected operator photo-audit filters.
type AuditListInput struct {
	// UserID optionally filters by photo owner.
	UserID int64
	// NiuID optionally filters by activated cattle.
	NiuID int64
	// PageNum is one-based and defaults when non-positive.
	PageNum int
	// PageSize defaults when non-positive and is capped by the service.
	PageSize int
}

// AuditListOutput is one bounded operator photo-audit page.
type AuditListOutput struct {
	// List contains the current page with batch-assembled display fields.
	List []*AuditItem
	// Total is the database-side matched row count.
	Total int
}

// AuditItem is one protected activation-photo audit row.
type AuditItem struct {
	// ActivationID is the associated activation record ID.
	ActivationID int64
	// PhotoID is the private stable photo token used by protected content reads.
	PhotoID string
	// UserID is the photo owner ID.
	UserID int64
	// Nickname is the batch-assembled owner nickname.
	Nickname string
	// NiuID is the associated cattle ID.
	NiuID int64
	// NiuName is the batch-assembled cattle display name.
	NiuName string
	// NiuCode is the batch-assembled cattle serial code.
	NiuCode string
	// SizeBytes is the standardized object size.
	SizeBytes int64
	// ContentType is the standardized object media type.
	ContentType string
	// ActivatedAt is the absolute activation time when available.
	ActivatedAt *time.Time
}

// serviceImpl implements Service with plugin-scoped private object storage.
type serviceImpl struct {
	// storage is the required plugin-scoped object storage capability.
	storage storagecap.Service
}

// Compile-time photo service contract assertion.
var _ Service = (*serviceImpl)(nil)

// New creates the photo service with required plugin-scoped private storage.
// A missing storage dependency is rejected during route initialization.
func New(storage storagecap.Service) (Service, error) {
	if storage == nil {
		return nil, gerror.New("activation photo service requires plugin-scoped storage")
	}
	return &serviceImpl{storage: storage}, nil
}

// Upload transcodes JPEG, PNG or HEIC input into a bounded WebP and serializes
// each player's daily slot allocation by locking the player row.
func (s *serviceImpl) Upload(ctx context.Context, playerID int64, in *UploadInput) (*Photo, error) {
	if playerID <= 0 || in == nil || in.Reader == nil {
		return nil, bizerr.NewCode(CodePhotoRequired)
	}
	requestID, requestOK := requestid.Normalize(in.RequestID)
	if !requestOK {
		return nil, bizerr.NewCode(CodePhotoRequired)
	}
	if s.storage == nil || in.SizeBytes <= 0 || in.SizeBytes > maxInputPhotoBytes {
		return nil, bizerr.NewCode(CodePhotoInvalid)
	}
	if existing, err := s.photoByRequest(ctx, playerID, requestID); err != nil || existing != nil {
		return photoResult(existing), err
	}

	input, err := io.ReadAll(io.LimitReader(in.Reader, maxInputPhotoBytes+1))
	if err != nil || len(input) == 0 || int64(len(input)) > maxInputPhotoBytes {
		return nil, bizerr.NewCode(CodePhotoInvalid)
	}
	standardized, err := standardizeImage(input, in.ContentType, in.Filename)
	if err != nil {
		return nil, bizerr.NewCode(CodePhotoInvalid)
	}

	var storedPath string
	var out *Photo
	err = dao.ActivationPhoto.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		var owner *entitymodel.User
		if lockErr := dao.User.Ctx(ctx).Where(dao.User.Columns().Id, playerID).LockUpdate().Scan(&owner); lockErr != nil {
			return bizerr.WrapCode(lockErr, CodePhotoQueryFailed)
		}
		if owner == nil {
			return bizerr.NewCode(CodePhotoInvalid)
		}
		existing, queryErr := s.photoByRequest(ctx, playerID, requestID)
		if queryErr != nil {
			return queryErr
		}
		if existing != nil {
			out = photoResult(existing)
			return nil
		}

		day := activityday.Today()
		slotRows := make([]struct {
			DailySlot int `json:"dailySlot"`
		}, 0, maxDailyPhotos)
		if queryErr = dao.ActivationPhoto.Ctx(ctx).
			Fields(dao.ActivationPhoto.Columns().DailySlot).
			Where(do.ActivationPhoto{UserId: playerID, ActivityDate: day}).
			OrderAsc(dao.ActivationPhoto.Columns().DailySlot).
			Scan(&slotRows); queryErr != nil {
			return bizerr.WrapCode(queryErr, CodePhotoQueryFailed)
		}
		slots := make([]int, 0, len(slotRows))
		for _, row := range slotRows {
			slots = append(slots, row.DailySlot)
		}
		slot := firstAvailableSlot(slots)
		if slot == 0 {
			return bizerr.NewCode(CodePhotoDailyLimit)
		}

		token := uuid.NewString()
		storedPath = "activation-photos/" + token + ".webp"
		if _, putErr := s.storage.Put(ctx, storagecap.PutInput{Path: storedPath, Body: bytes.NewReader(standardized), Size: int64(len(standardized)), ContentType: "image/webp", Overwrite: false}); putErr != nil {
			return bizerr.WrapCode(putErr, CodePhotoWriteFailed)
		}
		_, insertErr := dao.ActivationPhoto.Ctx(ctx).Data(do.ActivationPhoto{
			Token: token, UserId: playerID, RequestId: requestID, ActivityDate: day, DailySlot: slot,
			ObjectPath: storedPath, OriginalName: strings.TrimSpace(in.Filename), ContentType: "image/webp", SizeBytes: int64(len(standardized)),
		}).Insert()
		if insertErr != nil {
			return bizerr.WrapCode(insertErr, CodePhotoWriteFailed)
		}
		out = &Photo{Token: token, ContentType: "image/webp", SizeBytes: int64(len(standardized))}
		return nil
	})
	if err != nil {
		if storedPath != "" {
			cleanupErr := s.storage.Delete(ctx, storagecap.DeleteInput{Path: storedPath})
			err = errors.Join(err, cleanupErr)
		}
		return nil, err
	}
	return out, nil
}

// Content reads an owned photo without revealing whether a foreign token exists.
func (s *serviceImpl) Content(ctx context.Context, playerID int64, token string) (*Content, error) {
	photo, err := s.ownedPhoto(ctx, playerID, token)
	if err != nil {
		return nil, err
	}
	return s.loadContent(ctx, photo)
}

// ContentForAudit is used only behind the host operator permission chain.
func (s *serviceImpl) ContentForAudit(ctx context.Context, token string) (*Content, error) {
	var photo *entitymodel.ActivationPhoto
	if strings.TrimSpace(token) != "" {
		err := dao.ActivationPhoto.Ctx(ctx).Where(dao.ActivationPhoto.Columns().Token, strings.TrimSpace(token)).Scan(&photo)
		if err != nil {
			return nil, bizerr.WrapCode(err, CodePhotoQueryFailed)
		}
	}
	if photo == nil {
		return nil, bizerr.NewCode(CodePhotoNotFound)
	}
	return s.loadContent(ctx, photo)
}

// loadContent reads one known metadata row from private storage with an output
// size guard and returns CodePhotoNotFound for missing objects.
func (s *serviceImpl) loadContent(ctx context.Context, photo *entitymodel.ActivationPhoto) (out *Content, err error) {
	stored, err := s.storage.Get(ctx, storagecap.GetInput{Path: photo.ObjectPath})
	if err != nil {
		return nil, bizerr.WrapCode(err, CodePhotoQueryFailed)
	}
	if stored == nil || !stored.Found || stored.Body == nil {
		return nil, bizerr.NewCode(CodePhotoNotFound)
	}
	defer closeutil.Close(ctx, stored.Body, &err, "close activation photo content failed")
	content, err := io.ReadAll(io.LimitReader(stored.Body, maxPhotoBytes+1))
	if err != nil {
		return nil, bizerr.WrapCode(err, CodePhotoQueryFailed)
	}
	if int64(len(content)) > maxPhotoBytes {
		return nil, bizerr.NewCode(CodePhotoInvalid)
	}
	return &Content{ContentType: photo.ContentType, SizeBytes: int64(len(content)), Bytes: content}, nil
}

// Validate checks owner and unused state without consuming the photo.
func (s *serviceImpl) Validate(ctx context.Context, playerID int64, token string) error {
	photo, err := s.ownedPhoto(ctx, playerID, token)
	if err != nil {
		return err
	}
	if photo.UsedAt != nil {
		return bizerr.NewCode(CodePhotoAlreadyUsed)
	}
	return nil
}

// Consume atomically marks an owned, unused photo as used in the caller's transaction.
func (s *serviceImpl) Consume(ctx context.Context, playerID int64, token string, usedAt time.Time) error {
	if playerID <= 0 || strings.TrimSpace(token) == "" {
		return bizerr.NewCode(CodePhotoRequired)
	}
	result, err := dao.ActivationPhoto.Ctx(ctx).
		Where(dao.ActivationPhoto.Columns().Token, strings.TrimSpace(token)).
		Where(dao.ActivationPhoto.Columns().UserId, playerID).
		Where(dao.ActivationPhoto.Columns().UsedAt + " IS NULL").
		Data(do.ActivationPhoto{UsedAt: &usedAt}).Update()
	if err != nil {
		return bizerr.WrapCode(err, CodePhotoWriteFailed)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return bizerr.WrapCode(err, CodePhotoWriteFailed)
	}
	if affected > 0 {
		return nil
	}
	photo, queryErr := s.ownedPhoto(ctx, playerID, token)
	if queryErr != nil {
		return queryErr
	}
	if photo.UsedAt != nil {
		return bizerr.NewCode(CodePhotoAlreadyUsed)
	}
	return bizerr.NewCode(CodePhotoNotFound)
}

// ownedPhoto returns one unused photo owned by playerID without revealing
// whether a foreign token exists.
func (s *serviceImpl) ownedPhoto(ctx context.Context, playerID int64, token string) (*entitymodel.ActivationPhoto, error) {
	if playerID <= 0 || strings.TrimSpace(token) == "" {
		return nil, bizerr.NewCode(CodePhotoNotFound)
	}
	var photo *entitymodel.ActivationPhoto
	err := dao.ActivationPhoto.Ctx(ctx).Where(do.ActivationPhoto{Token: strings.TrimSpace(token), UserId: playerID}).Scan(&photo)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodePhotoQueryFailed)
	}
	if photo == nil {
		return nil, bizerr.NewCode(CodePhotoNotFound)
	}
	return photo, nil
}

// photoByRequest returns the stable upload row for one player's request ID.
func (s *serviceImpl) photoByRequest(ctx context.Context, playerID int64, requestID string) (*entitymodel.ActivationPhoto, error) {
	var photo *entitymodel.ActivationPhoto
	err := dao.ActivationPhoto.Ctx(ctx).Where(do.ActivationPhoto{UserId: playerID, RequestId: requestID}).Scan(&photo)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodePhotoQueryFailed)
	}
	return photo, nil
}

// photoResult maps a persisted metadata row to the player-facing photo result.
func photoResult(photo *entitymodel.ActivationPhoto) *Photo {
	if photo == nil {
		return nil
	}
	return &Photo{Token: photo.Token, ContentType: photo.ContentType, SizeBytes: photo.SizeBytes}
}

// firstAvailableSlot returns the first unused daily slot from 1 through the
// configured daily quota, or zero when the quota is exhausted.
func firstAvailableSlot(used []int) int {
	taken := make(map[int]bool, len(used))
	for _, slot := range used {
		taken[slot] = true
	}
	for slot := 1; slot <= maxDailyPhotos; slot++ {
		if !taken[slot] {
			return slot
		}
	}
	return 0
}

// standardizeImage decodes a supported image, bounds its pixels and dimensions,
// then reduces WebP quality until the output satisfies maxPhotoBytes.
func standardizeImage(content []byte, declaredType, filename string) ([]byte, error) {
	img, err := decodeImage(content, declaredType, filename)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 || int64(bounds.Dx())*int64(bounds.Dy()) > maxDecodedPixels {
		return nil, errors.New("invalid image dimensions")
	}
	current := img
	for resizeAttempt := 0; resizeAttempt < 8; resizeAttempt++ {
		for quality := 84; quality >= 28; quality -= 8 {
			var encoded bytes.Buffer
			if err = webp.Encode(&encoded, current, webp.Options{Quality: quality, Method: 6}); err != nil {
				return nil, err
			}
			if encoded.Len() <= int(maxPhotoBytes) {
				return encoded.Bytes(), nil
			}
		}
		width := current.Bounds().Dx() * 4 / 5
		height := current.Bounds().Dy() * 4 / 5
		if width < 320 || height < 320 {
			break
		}
		resized := image.NewNRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(resized, resized.Bounds(), current, current.Bounds(), draw.Over, nil)
		current = resized
	}
	return nil, errors.New("unable to compress image to limit")
}

// decodeImage decodes JPEG, PNG or HEIC based on validated content metadata.
func decodeImage(content []byte, declaredType, filename string) (image.Image, error) {
	detected := strings.ToLower(strings.TrimSpace(http.DetectContentType(content)))
	declared := strings.ToLower(strings.TrimSpace(strings.Split(declaredType, ";")[0]))
	extension := strings.ToLower(filepath.Ext(filename))
	reader := bytes.NewReader(content)
	switch {
	case detected == "image/jpeg":
		return jpeg.Decode(reader)
	case detected == "image/png":
		return png.Decode(reader)
	case declared == "image/heic" || declared == "image/heif" || extension == ".heic" || extension == ".heif" || looksLikeHEIC(content):
		return heic.Decode(reader)
	default:
		return nil, errors.New("unsupported image type")
	}
}

// looksLikeHEIC reports whether content contains a supported ISO-BMFF HEIC brand.
func looksLikeHEIC(content []byte) bool {
	if len(content) < 12 || string(content[4:8]) != "ftyp" {
		return false
	}
	brand := string(content[8:12])
	return brand == "heic" || brand == "heix" || brand == "hevc" || brand == "hevx" || brand == "mif1" || brand == "msf1"
}
