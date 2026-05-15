// This file provides a CLI for stressing LinaPro water plugin APIs.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	waterstress "lina-plugin-water/hack/stress/internal/stress"
)

// loggerAdapter adapts the standard logger to the stress FailureLogger contract.
type loggerAdapter struct {
	logger *log.Logger
}

// Printf writes one failed request diagnostic line.
func (l loggerAdapter) Printf(format string, args ...any) {
	if l.logger == nil {
		return
	}
	l.logger.Printf(format, args...)
}

// main parses flags and executes one water API stress test.
func main() {
	config, baseURL := parseFlags()
	if config.URL == "" {
		config.URL = waterstress.DefaultURL(baseURL, config.Mode, config.DeviceType, config.DeviceID)
	}

	failureLogger, closeLog, err := createFailureLogger(config.ErrorLogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建错误日志失败: %v\n", err)
		os.Exit(1)
	}
	defer closeLog()

	normalized, err := waterstress.NormalizeConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "参数错误: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Starting water stress test: url=%s concurrency=%d requests=%d mode=%s\n",
		normalized.URL, normalized.Concurrency, normalized.TotalRequests, normalized.Mode)
	if normalized.ErrorLogPath != "" {
		fmt.Printf("错误日志: %s\n", normalized.ErrorLogPath)
	}

	ctx, stop := contextWithSignal()
	defer stop()

	result, err := waterstress.Run(
		ctx,
		normalized,
		&http.Client{Timeout: normalized.RequestTimeout},
		failureLogger,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "压测失败: %v\n", err)
		os.Exit(1)
	}
	waterstress.PrintReport(os.Stdout, result)
}

// parseFlags parses CLI flags into one stress config.
func parseFlags() (waterstress.Config, string) {
	config := waterstress.Config{}
	baseURL := ""
	timeout := 30 * time.Second

	flag.StringVar(&baseURL, "base-url", "http://127.0.0.1:8080/api/v1", "LinaPro API base URL used when -url is empty")
	flag.StringVar(&config.URL, "url", "", "Target water API URL; defaults from -base-url and -mode")
	flag.StringVar(&config.BearerToken, "token", "", "Bearer token without the Bearer prefix")
	flag.StringVar(&config.Data, "data", "", "JSON request body")
	flag.StringVar(&config.ImagePath, "image", "", "Image file used to build request body when -data is empty")
	flag.IntVar(&config.Concurrency, "concurrency", 10, "Number of concurrent workers")
	flag.IntVar(&config.TotalRequests, "requests", 100, "Total request count")
	flag.StringVar(&config.Tenant, "tenant", "tenant-a", "Media tenant ID")
	flag.StringVar(&config.DeviceType, "device-type", "gb", "Device type used by submit mode")
	flag.StringVar(&config.DeviceID, "device-id", "34020000001320000001", "Device ID")
	flag.StringVar(&config.DeviceCode, "device-code", "", "Device GB code; defaults to -device-id")
	flag.StringVar(&config.ChannelCode, "channel-code", "", "Channel code; defaults to -device-id")
	flag.StringVar(&config.Mode, "mode", waterstress.ModePreview, "Request mode: preview or submit")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "Per-request timeout")
	flag.StringVar(&config.ErrorLogPath, "error-log", "logs/water_stress_error.log", "Failed request log path")
	flag.IntVar(&config.ResponsePreview, "response-preview", 512, "Failed response body preview length")
	flag.Parse()

	config.RequestTimeout = timeout
	return config, baseURL
}

// createFailureLogger creates the failure log writer.
func createFailureLogger(path string) (waterstress.FailureLogger, func(), error) {
	if path == "" {
		return nil, func() {}, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	closeFn := func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "关闭错误日志失败: %v\n", err)
		}
	}
	return loggerAdapter{logger: log.New(file, "", log.LstdFlags)}, closeFn, nil
}

// contextWithSignal returns a root context cancelled by interrupt and terminate signals.
func contextWithSignal() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
