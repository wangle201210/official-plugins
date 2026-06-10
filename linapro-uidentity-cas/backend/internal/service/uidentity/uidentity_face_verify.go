// This file implements the old uidentity/admin external face verification
// client: an artemis HMAC-SHA256 signed POST comparing one submitted face
// image against the certificate photo library. Verification activates only
// when the plugin config provides artemis credentials; without credentials the
// activation face step keeps the plugin-local marker behavior.

package uidentity

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

const (
	configKeyLegacyFaceAK         = "legacy.face.ak"
	configKeyLegacyFaceSK         = "legacy.face.sk"
	configKeyLegacyFaceBaseURL    = "legacy.face.baseUrl"
	configKeyLegacyFacePath       = "legacy.face.path"
	configKeyLegacyFaceSimilarity = "legacy.face.similarity"

	// Defaults match the old common/face/face.go constants and the old
	// settings.yml similarity threshold.
	defaultLegacyFaceBaseURL    = "https://facelib.sicau.edu.cn:443"
	defaultLegacyFacePath       = "/artemis/api/cflms/v1/face/verification"
	defaultLegacyFaceSimilarity = 80

	legacyFaceFailureMsg = "人脸核验不通过，请重传照片或持个人有效证件前往三校区信教中心前台人工核验"

	// legacyFaceInlineBase64MinLength matches the old heuristic: shorter
	// faceUrl values are storage paths, longer values are inline base64 data.
	legacyFaceInlineBase64MinLength = 1024
)

// legacyFaceRequest carries the old artemis face verification payload.
type legacyFaceRequest struct {
	CertNo               string `json:"certNo"`
	SrcFacePicBase64Data string `json:"srcFacePicBase64Data"`
	JobNo                string `json:"jobNo"`
}

// legacyFaceResponse carries the old artemis face verification result.
type legacyFaceResponse struct {
	Code string `json:"code"`
	Data struct {
		Similarity string `json:"similarity"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// similarityScore returns the parsed similarity or zero on failure payloads.
func (x *legacyFaceResponse) similarityScore() int64 {
	if x.Code != "0" || x.Msg != "success" {
		return 0
	}
	score, err := strconv.Atoi(strings.TrimSpace(x.Data.Similarity))
	if err != nil {
		return 0
	}
	return int64(score)
}

// verifyActivationFace compares one submitted face image against the external
// certificate photo library. It returns pass=true without an external call
// when artemis credentials are not configured, preserving the plugin-local
// marker behavior for deployments without the face service.
func (s *serviceImpl) verifyActivationFace(ctx context.Context, number string, certNo string, faceURL string) (bool, string, error) {
	ak, err := s.configSvc.String(ctx, configKeyLegacyFaceAK, "")
	if err != nil {
		return false, "", err
	}
	sk, err := s.configSvc.String(ctx, configKeyLegacyFaceSK, "")
	if err != nil {
		return false, "", err
	}
	if strings.TrimSpace(ak) == "" || strings.TrimSpace(sk) == "" {
		return true, "", nil
	}
	baseURL, err := s.configSvc.String(ctx, configKeyLegacyFaceBaseURL, defaultLegacyFaceBaseURL)
	if err != nil {
		return false, "", err
	}
	requestPath, err := s.configSvc.String(ctx, configKeyLegacyFacePath, defaultLegacyFacePath)
	if err != nil {
		return false, "", err
	}
	threshold, err := s.configSvc.Int(ctx, configKeyLegacyFaceSimilarity, defaultLegacyFaceSimilarity)
	if err != nil {
		return false, "", err
	}
	image, err := s.resolveFaceImageBase64(ctx, faceURL)
	if err != nil {
		return false, "", err
	}
	response, err := sendLegacyFaceVerification(ctx, baseURL, requestPath, ak, sk, &legacyFaceRequest{
		CertNo:               strings.TrimSpace(certNo),
		SrcFacePicBase64Data: image,
		JobNo:                strings.TrimSpace(number),
	})
	if err != nil {
		return false, "", err
	}
	if response.Msg != "success" {
		return false, response.Msg + "，" + legacyFaceFailureMsg, nil
	}
	if response.similarityScore() < int64(threshold) {
		logger.Infof(ctx, "legacy face verification rejected number=%s similarity=%s msg=%s",
			number, response.Data.Similarity, response.Msg)
		return false, legacyFaceFailureMsg, nil
	}
	return true, "", nil
}

// resolveFaceImageBase64 keeps the old faceUrl heuristic: long values are
// inline base64 image data, short values are plugin-local upload paths.
func (s *serviceImpl) resolveFaceImageBase64(ctx context.Context, faceURL string) (string, error) {
	value := strings.TrimSpace(faceURL)
	if value == "" {
		return "", bizerr.NewCode(CodeFaceVerifyFailed)
	}
	if len(value) >= legacyFaceInlineBase64MinLength {
		return value, nil
	}
	root, err := s.configSvc.String(ctx, configKeyLegacyUploadPath, defaultLegacyUploadPath)
	if err != nil {
		return "", err
	}
	candidate := filepath.Clean(strings.TrimPrefix(strings.TrimPrefix(value, "./"), "/"))
	if strings.HasPrefix(candidate, "..") {
		return "", bizerr.NewCode(CodeFaceVerifyFailed)
	}
	for _, path := range []string{candidate, filepath.Join(root, filepath.Base(candidate))} {
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		return base64.StdEncoding.EncodeToString(content), nil
	}
	return "", bizerr.NewCode(CodeFaceVerifyFailed)
}

// sendLegacyFaceVerification posts one artemis face verification request with
// the old x-ca HMAC-SHA256 signature headers.
func sendLegacyFaceVerification(
	ctx context.Context,
	baseURL string,
	requestPath string,
	ak string,
	sk string,
	payload *legacyFaceRequest,
) (*legacyFaceResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	nonce := legacyFaceNonce(32)
	response, err := g.Client().
		SetTimeout(15*time.Second).
		Header(map[string]string{
			"Accept":                 "*/*",
			"Content-Type":           "application/json;charset=UTF-8",
			"x-ca-timestamp":         timestamp,
			"x-ca-key":               ak,
			"x-ca-signature":         legacyFaceSignature(ak, sk, nonce, timestamp, requestPath),
			"x-ca-signature-headers": "x-ca-key,x-ca-nonce,x-ca-timestamp",
			"x-ca-nonce":             nonce,
		}).
		Post(ctx, strings.TrimRight(baseURL, "/")+requestPath, string(body))
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFaceVerifyFailed)
	}
	defer response.Close()
	result := &legacyFaceResponse{}
	if err := json.Unmarshal(response.ReadAll(), result); err != nil {
		return nil, bizerr.WrapCode(err, CodeFaceVerifyFailed)
	}
	return result, nil
}

// legacyFaceSignature builds the old artemis x-ca-signature digest.
func legacyFaceSignature(ak string, sk string, nonce string, timestamp string, requestPath string) string {
	content := strings.Join([]string{
		"POST",
		"*/*",
		"application/json;charset=UTF-8",
		"x-ca-key:" + ak,
		"x-ca-nonce:" + nonce,
		"x-ca-timestamp:" + timestamp,
		requestPath,
	}, "\n")
	mac := hmac.New(sha256.New, []byte(sk))
	_, _ = mac.Write([]byte(content))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// legacyFaceNonce returns one random base64 nonce for the artemis signature.
func legacyFaceNonce(length int) string {
	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(buffer)
}
