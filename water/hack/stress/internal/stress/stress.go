// Package stress implements the water plugin HTTP concurrency stress runner.
package stress

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// ModePreview posts to the synchronous watermark preview API.
	ModePreview = "preview"
	// ModeSubmit posts to the asynchronous watermark submit API.
	ModeSubmit = "submit"
)

// Config defines one water stress test run.
type Config struct {
	URL             string        // URL is the target endpoint.
	BearerToken     string        // BearerToken is the optional authorization token without the Bearer prefix.
	Data            string        // Data is the JSON request body.
	Concurrency     int           // Concurrency is the number of concurrent workers.
	TotalRequests   int           // TotalRequests is the total number of requests to send.
	ImagePath       string        // ImagePath optionally overrides the image field with file content.
	Tenant          string        // Tenant is used when building a request body from ImagePath.
	DeviceType      string        // DeviceType is used by submit mode URL defaults.
	DeviceID        string        // DeviceID identifies the media device.
	DeviceCode      string        // DeviceCode is the GB device code sent in the request body.
	ChannelCode     string        // ChannelCode is sent in the request body.
	Mode            string        // Mode selects preview or submit body defaults.
	RequestTimeout  time.Duration // RequestTimeout is the per-request timeout.
	ErrorLogPath    string        // ErrorLogPath stores failed request details when non-empty.
	ResponsePreview int           // ResponsePreview limits failed response body snippets in logs.
}

// Result contains aggregate stress test metrics.
type Result struct {
	TotalRequests   int
	SuccessRequests int
	FailedRequests  int
	TotalDuration   time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	AvgResponseTime time.Duration
	RPS             float64
}

// HTTPDoer is the minimal HTTP client contract used by Run.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// FailureLogger writes failed request diagnostics.
type FailureLogger interface {
	Printf(format string, args ...any)
}

// Run executes one water stress test.
func Run(ctx context.Context, config Config, client HTTPDoer, failureLogger FailureLogger) (*Result, error) {
	normalized, err := NormalizeConfig(config)
	if err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{Timeout: normalized.RequestTimeout}
	}

	var (
		wg                sync.WaitGroup
		mutex             sync.Mutex
		totalResponseTime time.Duration
		result            = &Result{MinResponseTime: time.Hour}
	)
	taskChan := make(chan int, normalized.TotalRequests)
	startedAt := time.Now()

	for workerID := 0; workerID < normalized.Concurrency; workerID++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range taskChan {
				responseTime, success := MakeRequest(ctx, normalized, client, failureLogger)
				mutex.Lock()
				if responseTime < result.MinResponseTime {
					result.MinResponseTime = responseTime
				}
				if responseTime > result.MaxResponseTime {
					result.MaxResponseTime = responseTime
				}
				totalResponseTime += responseTime
				if success {
					result.SuccessRequests++
				} else {
					result.FailedRequests++
				}
				mutex.Unlock()
			}
		}()
	}

	for i := 0; i < normalized.TotalRequests; i++ {
		select {
		case <-ctx.Done():
			close(taskChan)
			wg.Wait()
			return nil, ctx.Err()
		case taskChan <- i:
		}
	}
	close(taskChan)
	wg.Wait()

	result.TotalRequests = normalized.TotalRequests
	result.TotalDuration = time.Since(startedAt)
	if result.TotalRequests == 0 {
		result.MinResponseTime = 0
	}
	if result.TotalRequests > 0 {
		result.AvgResponseTime = time.Duration(int64(totalResponseTime) / int64(result.TotalRequests))
	}
	if result.TotalDuration > 0 {
		result.RPS = float64(result.TotalRequests) / result.TotalDuration.Seconds()
	}
	return result, nil
}

// NormalizeConfig validates and completes one stress config.
func NormalizeConfig(config Config) (Config, error) {
	config.URL = strings.TrimSpace(config.URL)
	config.BearerToken = strings.TrimSpace(config.BearerToken)
	config.Data = strings.TrimSpace(config.Data)
	config.ImagePath = strings.TrimSpace(config.ImagePath)
	config.Tenant = defaultString(config.Tenant, "tenant-a")
	config.DeviceType = defaultString(config.DeviceType, "gb")
	config.DeviceID = defaultString(config.DeviceID, "34020000001320000001")
	config.DeviceCode = defaultString(config.DeviceCode, config.DeviceID)
	config.ChannelCode = defaultString(config.ChannelCode, config.DeviceID)
	config.Mode = defaultString(config.Mode, ModePreview)
	if config.Mode != ModePreview && config.Mode != ModeSubmit {
		return config, fmt.Errorf("mode must be preview or submit")
	}
	if config.Concurrency <= 0 {
		config.Concurrency = 1
	}
	if config.TotalRequests <= 0 {
		config.TotalRequests = 1
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = 30 * time.Second
	}
	if config.ResponsePreview <= 0 {
		config.ResponsePreview = 512
	}
	if config.URL == "" {
		return config, fmt.Errorf("url is required")
	}
	if config.ImagePath != "" {
		body, err := BuildBodyFromImage(config)
		if err != nil {
			return config, err
		}
		config.Data = body
	}
	if config.Data == "" {
		return config, fmt.Errorf("data or image path is required")
	}
	if !json.Valid([]byte(config.Data)) {
		return config, fmt.Errorf("data must be valid JSON")
	}
	return config, nil
}

// BuildBodyFromImage builds a LinaPro water request body from an image file.
func BuildBodyFromImage(config Config) (string, error) {
	content, err := os.ReadFile(config.ImagePath)
	if err != nil {
		return "", fmt.Errorf("read image file: %w", err)
	}
	body := map[string]string{
		"tenant":      defaultString(config.Tenant, "tenant-a"),
		"deviceId":    defaultString(config.DeviceID, "34020000001320000001"),
		"deviceCode":  defaultString(config.DeviceCode, defaultString(config.DeviceID, "34020000001320000001")),
		"channelCode": defaultString(config.ChannelCode, defaultString(config.DeviceID, "34020000001320000001")),
		"imageName":   imageName(config.ImagePath),
		"image":       "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(content),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal image request body: %w", err)
	}
	return string(payload), nil
}

// MakeRequest sends one HTTP request and returns its response time and success state.
func MakeRequest(ctx context.Context, config Config, client HTTPDoer, failureLogger FailureLogger) (time.Duration, bool) {
	startedAt := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.URL, bytes.NewBufferString(config.Data))
	if err != nil {
		logFailure(failureLogger, "创建请求失败: %v, URL: %s", err, config.URL)
		return time.Since(startedAt), false
	}
	req.Header.Set("Content-Type", "application/json")
	if config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.BearerToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		logFailure(failureLogger, "发送请求失败: %v, URL: %s", err, config.URL)
		return time.Since(startedAt), false
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logFailure(failureLogger, "关闭响应体失败: %v, URL: %s", closeErr, config.URL)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logFailure(failureLogger, "读取响应失败: %v, URL: %s, StatusCode: %d", err, config.URL, resp.StatusCode)
		return time.Since(startedAt), false
	}
	success := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
	if !success {
		logFailure(
			failureLogger,
			"请求失败: URL: %s, StatusCode: %d, ResponseTime: %v, Body: %s",
			config.URL,
			resp.StatusCode,
			time.Since(startedAt),
			previewBody(body, config.ResponsePreview),
		)
	}
	return time.Since(startedAt), success
}

// SuccessRate returns the successful request percentage.
func (r *Result) SuccessRate() float64 {
	if r == nil || r.TotalRequests == 0 {
		return 0
	}
	return float64(r.SuccessRequests) / float64(r.TotalRequests) * 100
}

// PrintReport writes one human-readable stress test report.
func PrintReport(writer io.Writer, result *Result) {
	if result == nil {
		fmt.Fprintln(writer, "No stress test result.")
		return
	}
	fmt.Fprintln(writer, "\n=== Water Stress Test Report ===")
	fmt.Fprintf(writer, "Total Requests:     %d\n", result.TotalRequests)
	fmt.Fprintf(writer, "Successful:         %d\n", result.SuccessRequests)
	fmt.Fprintf(writer, "Failed:             %d\n", result.FailedRequests)
	fmt.Fprintf(writer, "Success Rate:       %.2f%%\n", result.SuccessRate())
	fmt.Fprintf(writer, "Total Duration:     %v\n", result.TotalDuration)
	fmt.Fprintf(writer, "Requests/Second:    %.2f\n", result.RPS)
	fmt.Fprintf(writer, "Min Response Time:  %v\n", result.MinResponseTime)
	fmt.Fprintf(writer, "Max Response Time:  %v\n", result.MaxResponseTime)
	fmt.Fprintf(writer, "Avg Response Time:  %v\n", result.AvgResponseTime)
	fmt.Fprintln(writer, "===============================")
}

// DefaultURL returns the default LinaPro water API URL for the selected mode.
func DefaultURL(baseURL string, mode string, deviceType string, deviceID string) string {
	baseURL = strings.TrimRight(defaultString(baseURL, "http://127.0.0.1:8080/api/v1"), "/")
	switch defaultString(mode, ModePreview) {
	case ModeSubmit:
		return fmt.Sprintf("%s/water/snaps/%s/%s", baseURL, url.PathEscape(deviceType), url.PathEscape(deviceID))
	default:
		return fmt.Sprintf("%s/water/preview", baseURL)
	}
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

// imageName extracts a request image name from a local path.
func imageName(path string) string {
	name := strings.TrimSpace(path)
	if name == "" {
		return "snap.jpg"
	}
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	if len(parts) == 0 {
		return "snap.jpg"
	}
	return parts[len(parts)-1]
}

// previewBody returns a bounded response-body snippet for failure logs.
func previewBody(body []byte, limit int) string {
	if limit <= 0 || len(body) <= limit {
		return string(body)
	}
	return string(body[:limit]) + "...(truncated)"
}

// logFailure writes a formatted failure message when a logger is configured.
func logFailure(logger FailureLogger, format string, args ...any) {
	if logger == nil {
		return
	}
	logger.Printf(format, args...)
}
