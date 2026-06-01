// This file implements legacy upload, health, monitor, log snapshot, and
// external-boundary operations that can run inside the source plugin.

package uidentity

import (
	"bufio"
	"context"
	"encoding/base64"
	"hash/fnv"
	"io"
	"mime"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	legacyUploadTypeSingle      = "1"
	legacyUploadTypeMulti       = "2"
	legacyUploadTypeBase64      = "3"
	legacyExternalTypeJobStart  = "job_start"
	legacyExternalTypeJobRemove = "job_remove"

	configKeyLegacyUploadPath       = "legacy.uploadPath"
	configKeyLegacyUploadPublicBase = "legacy.uploadPublicBase"
	configKeyLegacyUploadMaxSizeMB  = "legacy.uploadMaxSizeMB"
	configKeyLegacyLogDir           = "legacy.logDir"
	configKeyLegacyMonitorLocation  = "legacy.monitorLocation"
	configKeyLegacyExternalEnabled  = "legacy.externalActionEnabled"

	defaultLegacyUploadPath      = "temp/uidentity-upload"
	defaultLegacyUploadMaxSizeMB = 10
	defaultLegacyLogDir          = "temp/logs"
	defaultLegacySnapshotLines   = 200
	maxLegacySnapshotLines       = 1000
)

const legacyJobStatusEnabled = 2

// UploadLegacyFiles stores legacy files in plugin-owned local storage.
func (s *serviceImpl) UploadLegacyFiles(ctx context.Context, in LegacyUploadInput) (*LegacyUploadOutput, error) {
	uploadType := strings.TrimSpace(in.Type)
	if uploadType == "" {
		uploadType = legacyUploadTypeSingle
	}
	if uploadType == legacyUploadTypeBase64 {
		file, err := s.saveLegacyBase64Upload(ctx, in.Base64File)
		if err != nil {
			return nil, err
		}
		return &LegacyUploadOutput{Files: []*LegacyUploadFile{file}}, nil
	}
	files := in.UploadFiles
	if len(files) == 0 {
		return nil, bizerr.NewCode(CodeLegacyUploadRequired)
	}
	if uploadType != legacyUploadTypeMulti && len(files) > 1 {
		files = files[:1]
	}
	result := make([]*LegacyUploadFile, 0, len(files))
	for _, file := range files {
		stored, err := s.saveLegacyMultipartUpload(ctx, file)
		if err != nil {
			return nil, err
		}
		result = append(result, stored)
	}
	return &LegacyUploadOutput{Files: result}, nil
}

// Health returns a local health status.
func (s *serviceImpl) Health(ctx context.Context) (*LegacyHealthOutput, error) {
	return &LegacyHealthOutput{Status: "ok"}, nil
}

// ServerMonitor returns runtime and OS information.
func (s *serviceImpl) ServerMonitor(ctx context.Context) (*LegacyServerMonitorOutput, error) {
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		logger.Warningf(ctx, "legacy server monitor host info failed err=%v", err)
	}
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		logger.Warningf(ctx, "legacy server monitor memory info failed err=%v", err)
	}
	swapInfo, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		logger.Warningf(ctx, "legacy server monitor swap info failed err=%v", err)
	}
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		logger.Warningf(ctx, "legacy server monitor cpu info failed err=%v", err)
	}
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		logger.Warningf(ctx, "legacy server monitor cpu percent failed err=%v", err)
	}
	rootDisk, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		logger.Warningf(ctx, "legacy server monitor disk usage failed err=%v", err)
	}
	bootTime, err := host.BootTimeWithContext(ctx)
	if err != nil {
		logger.Warningf(ctx, "legacy server monitor boot time failed err=%v", err)
	}
	location, err := s.configSvc.String(ctx, configKeyLegacyMonitorLocation, "local")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(location) == "" {
		location = "local"
	}
	osInfo := map[string]any{
		"goOs":         runtime.GOOS,
		"arch":         runtime.GOARCH,
		"compiler":     runtime.Compiler,
		"version":      runtime.Version(),
		"numGoroutine": runtime.NumGoroutine(),
		"ip":           firstLocalIP(),
		"projectDir":   gfile.Pwd(),
		"hostName":     "",
		"time":         time.Now().Format("2006-01-02 15:04:05"),
	}
	if hostInfo != nil {
		osInfo["hostName"] = hostInfo.Hostname
	}
	memData := map[string]any{"used": uint64(0), "total": uint64(0), "percent": float64(0)}
	if memInfo != nil {
		memData["used"] = memInfo.Used / 1024 / 1024
		memData["total"] = memInfo.Total / 1024 / 1024
		memData["percent"] = roundFloat(memInfo.UsedPercent, 2)
	}
	swapData := map[string]any{"used": uint64(0), "total": uint64(0)}
	if swapInfo != nil {
		swapData["used"] = swapInfo.Used
		swapData["total"] = swapInfo.Total
	}
	cpuData := map[string]any{"cpuInfo": cpuInfo, "percent": float64(0), "cpuNum": 0}
	if len(cpuPercent) > 0 {
		cpuData["percent"] = roundFloat(cpuPercent[0], 2)
	}
	if cpuNum, err := cpu.CountsWithContext(ctx, false); err == nil {
		cpuData["cpuNum"] = cpuNum
	} else {
		logger.Warningf(ctx, "legacy server monitor cpu count failed err=%v", err)
	}
	diskData := map[string]any{"total": float64(0), "used": float64(0), "percent": float64(0)}
	if rootDisk != nil {
		diskData["total"] = float64(rootDisk.Total / 1024 / 1024 / 1024)
		diskData["used"] = float64(rootDisk.Used / 1024 / 1024 / 1024)
		diskData["percent"] = roundFloat(rootDisk.UsedPercent, 2)
	}
	bootHours := int64(0)
	if bootTime > 0 {
		bootHours = int64(time.Since(time.Unix(int64(bootTime), 0)).Hours())
	}
	return &LegacyServerMonitorOutput{
		Code:     200,
		OS:       osInfo,
		Mem:      memData,
		CPU:      cpuData,
		Disk:     diskData,
		Net:      map[string]any{"in": float64(0), "out": float64(0)},
		Swap:     swapData,
		Location: location,
		BootTime: bootHours,
	}, nil
}

// LogSnapshot returns a bounded tail of one configured log file.
func (s *serviceImpl) LogSnapshot(ctx context.Context, in LegacyLogSnapshotInput) (*LegacyLogSnapshotOutput, error) {
	date := strings.TrimSpace(in.Date)
	if date == "" {
		date = time.Now().Format(time.DateOnly)
	}
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return nil, bizerr.NewCode(CodeLegacyLogInvalid)
	}
	lines := in.Lines
	if lines <= 0 {
		lines = defaultLegacySnapshotLines
	}
	if lines > maxLegacySnapshotLines {
		lines = maxLegacySnapshotLines
	}
	logDir, err := s.configSvc.String(ctx, configKeyLegacyLogDir, defaultLegacyLogDir)
	if err != nil {
		return nil, err
	}
	path := filepath.Clean(gfile.Join(logDir, date+".log"))
	items, exists, truncated, err := readTailLines(ctx, path, lines)
	if err != nil {
		return nil, err
	}
	return &LegacyLogSnapshotOutput{
		Date:      date,
		Path:      path,
		Lines:     items,
		Exists:    exists,
		Truncated: truncated,
	}, nil
}

// RunExternalAction reports configured support for external legacy actions.
func (s *serviceImpl) RunExternalAction(ctx context.Context, in LegacyExternalActionInput) (*LegacyExternalActionOutput, error) {
	switch strings.TrimSpace(in.Type) {
	case legacyExternalTypeJobStart:
		return s.startLegacyJob(ctx, in.Target)
	case legacyExternalTypeJobRemove:
		return s.removeLegacyJob(ctx, in.Target)
	}
	enabled, err := s.configSvc.Bool(ctx, configKeyLegacyExternalEnabled, false)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, bizerr.NewCode(CodeUnsupportedExternalFlow)
	}
	return nil, bizerr.NewCode(CodeUnsupportedExternalFlow)
}

func (s *serviceImpl) startLegacyJob(ctx context.Context, target string) (*LegacyExternalActionOutput, error) {
	job, err := s.legacyJobByTarget(ctx, target)
	if err != nil {
		return nil, err
	}
	if job.Status != legacyJobStatusEnabled {
		return nil, bizerr.NewCode(CodeLegacyJobDisabled)
	}
	entryID := legacyRuntimeEntryID(job)
	_, err = s.tenantFilter.Apply(ctx, dao.SysJob.Ctx(ctx), "").
		Where(dao.SysJob.Columns().JobId, job.JobId).
		Data(do.SysJob{EntryId: entryID, UpdatedBy: s.actorID(ctx)}).
		Update()
	if err != nil {
		return nil, err
	}
	return &LegacyExternalActionOutput{
		Type:    legacyExternalTypeJobStart,
		Target:  strings.TrimSpace(target),
		Success: true,
	}, nil
}

func (s *serviceImpl) removeLegacyJob(ctx context.Context, target string) (*LegacyExternalActionOutput, error) {
	job, err := s.legacyJobByTarget(ctx, target)
	if err != nil {
		return nil, err
	}
	_, err = s.tenantFilter.Apply(ctx, dao.SysJob.Ctx(ctx), "").
		Where(dao.SysJob.Columns().JobId, job.JobId).
		Data(do.SysJob{EntryId: 0, UpdatedBy: s.actorID(ctx)}).
		Update()
	if err != nil {
		return nil, err
	}
	return &LegacyExternalActionOutput{
		Type:    legacyExternalTypeJobRemove,
		Target:  strings.TrimSpace(target),
		Success: true,
	}, nil
}

func (s *serviceImpl) legacyJobByTarget(ctx context.Context, target string) (*entity.SysJob, error) {
	jobID := parseLegacyJobID(target)
	if jobID <= 0 {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	var job *entity.SysJob
	err := s.tenantFilter.Apply(ctx, dao.SysJob.Ctx(ctx), "").
		Where(dao.SysJob.Columns().JobId, jobID).
		Scan(&job)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	return job, nil
}

func parseLegacyJobID(target string) int64 {
	return gconv.Int64(strings.TrimSpace(target))
}

func legacyRuntimeEntryID(job *entity.SysJob) int64 {
	if job == nil {
		return 0
	}
	if job.EntryId > 0 {
		return job.EntryId
	}
	return legacyEntryID(job.JobId)
}

func legacyEntryID(jobID int64) int64 {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(gconv.String(jobID)))
	return int64(hash.Sum32())
}

func (s *serviceImpl) saveLegacyMultipartUpload(ctx context.Context, file *ghttp.UploadFile) (output *LegacyUploadFile, err error) {
	if file == nil {
		return nil, bizerr.NewCode(CodeLegacyUploadRequired)
	}
	if err := s.validateLegacyUploadSize(ctx, file.Size); err != nil {
		return nil, err
	}
	source, err := file.Open()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeLegacyUploadFailed)
	}
	defer func() {
		if closeErr := source.Close(); closeErr != nil && err == nil {
			err = bizerr.WrapCode(closeErr, CodeLegacyUploadFailed)
		}
	}()
	relativePath, fullPath, err := s.nextLegacyUploadPath(ctx, file.Filename)
	if err != nil {
		return nil, err
	}
	if err = writeLegacyUploadFile(fullPath, source); err != nil {
		return nil, err
	}
	mimeType := mime.TypeByExtension(filepath.Ext(fullPath))
	return s.legacyUploadProjection(ctx, relativePath, file.Filename, mimeType)
}

func (s *serviceImpl) saveLegacyBase64Upload(ctx context.Context, data string) (*LegacyUploadFile, error) {
	mimeType, payload, err := splitLegacyBase64Payload(data)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, bizerr.NewCode(CodeLegacyUploadInvalid)
	}
	if err := s.validateLegacyUploadSize(ctx, int64(len(decoded))); err != nil {
		return nil, err
	}
	filename := "base64.jpg"
	if extensions, err := mime.ExtensionsByType(mimeType); err == nil && len(extensions) > 0 {
		filename = "base64" + extensions[0]
	}
	relativePath, fullPath, err := s.nextLegacyUploadPath(ctx, filename)
	if err != nil {
		return nil, err
	}
	if err := gfile.Mkdir(filepath.Dir(fullPath)); err != nil {
		return nil, bizerr.WrapCode(err, CodeLegacyUploadFailed)
	}
	if err := os.WriteFile(fullPath, decoded, 0o644); err != nil {
		return nil, bizerr.WrapCode(err, CodeLegacyUploadFailed)
	}
	return s.legacyUploadProjection(ctx, relativePath, "", mimeType)
}

func (s *serviceImpl) nextLegacyUploadPath(ctx context.Context, original string) (relativePath string, fullPath string, err error) {
	root, err := s.configSvc.String(ctx, configKeyLegacyUploadPath, defaultLegacyUploadPath)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(root) == "" {
		root = defaultLegacyUploadPath
	}
	now := gtime.Now()
	dir := filepath.Join(now.Format("Y"), now.Format("m"))
	filename := legacySafeFilename(original)
	ext := filepath.Ext(filename)
	storedName := now.Format("Ymd_His") + "_" + grand.S(8)
	if ext != "" {
		storedName += strings.ToLower(ext)
	}
	relativePath = gfile.Join(dir, storedName)
	fullPath = filepath.Clean(gfile.Join(root, relativePath))
	return relativePath, fullPath, nil
}

func (s *serviceImpl) legacyUploadProjection(ctx context.Context, relativePath string, original string, mimeType string) (*LegacyUploadFile, error) {
	root, err := s.configSvc.String(ctx, configKeyLegacyUploadPath, defaultLegacyUploadPath)
	if err != nil {
		return nil, err
	}
	publicBase, err := s.configSvc.String(ctx, configKeyLegacyUploadPublicBase, "/uidentity/uploads")
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Clean(gfile.Join(root, relativePath))
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeLegacyUploadFailed)
	}
	if strings.TrimSpace(mimeType) == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(fullPath))
	}
	relativePublicPath := strings.TrimPrefix(filepath.ToSlash(relativePath), "/")
	publicBase = strings.Trim(strings.TrimSpace(publicBase), "/")
	publicPath := "/" + relativePublicPath
	if publicBase != "" {
		publicPath = "/" + publicBase + "/" + relativePublicPath
	}
	return &LegacyUploadFile{
		Size:     info.Size(),
		Path:     publicPath,
		FullPath: publicPath,
		Name:     original,
		Type:     mimeType,
	}, nil
}

func (s *serviceImpl) validateLegacyUploadSize(ctx context.Context, size int64) error {
	limit, err := s.configSvc.Int(ctx, configKeyLegacyUploadMaxSizeMB, defaultLegacyUploadMaxSizeMB)
	if err != nil {
		return err
	}
	if limit <= 0 {
		limit = defaultLegacyUploadMaxSizeMB
	}
	if size > int64(limit)*1024*1024 {
		return bizerr.NewCode(CodeLegacyUploadInvalid)
	}
	return nil
}

func writeLegacyUploadFile(fullPath string, source io.Reader) (err error) {
	if err := gfile.Mkdir(filepath.Dir(fullPath)); err != nil {
		return bizerr.WrapCode(err, CodeLegacyUploadFailed)
	}
	target, err := os.Create(fullPath)
	if err != nil {
		return bizerr.WrapCode(err, CodeLegacyUploadFailed)
	}
	defer func() {
		if closeErr := target.Close(); closeErr != nil && err == nil {
			err = bizerr.WrapCode(closeErr, CodeLegacyUploadFailed)
		}
	}()
	if _, err = io.Copy(target, source); err != nil {
		if removeErr := os.Remove(fullPath); removeErr != nil && !os.IsNotExist(removeErr) {
			return bizerr.WrapCode(removeErr, CodeLegacyUploadFailed)
		}
		return bizerr.WrapCode(err, CodeLegacyUploadFailed)
	}
	return nil
}

func splitLegacyBase64Payload(data string) (mimeType string, payload string, err error) {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return "", "", bizerr.NewCode(CodeLegacyUploadRequired)
	}
	parts := strings.SplitN(trimmed, ",", 2)
	if len(parts) != 2 {
		return "image/jpeg", trimmed, nil
	}
	prefix := strings.TrimSpace(parts[0])
	mimeType = strings.TrimSuffix(strings.TrimPrefix(prefix, "data:"), ";base64")
	if mimeType == "" {
		mimeType = "image/jpeg"
	}
	return mimeType, strings.TrimSpace(parts[1]), nil
}

func legacySafeFilename(filename string) string {
	name := filepath.Base(strings.TrimSpace(filename))
	if name == "" || name == "." {
		name = "upload"
	}
	name = strings.ReplaceAll(name, "\x00", "")
	for _, value := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"} {
		name = strings.ReplaceAll(name, value, "_")
	}
	if len(name) > 255 {
		ext := filepath.Ext(name)
		name = name[:255-len(ext)] + ext
	}
	return name
}

func readTailLines(ctx context.Context, path string, limit int) (items []string, exists bool, truncated bool, err error) {
	if !gfile.Exists(path) {
		return []string{}, false, false, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, true, false, bizerr.WrapCode(err, CodeLegacyLogInvalid)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = bizerr.WrapCode(closeErr, CodeLegacyLogInvalid)
		}
	}()
	items = make([]string, 0, limit)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(items) == limit {
			copy(items, items[1:])
			items[len(items)-1] = scanner.Text()
			truncated = true
			continue
		}
		items = append(items, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, true, truncated, bizerr.WrapCode(err, CodeLegacyLogInvalid)
	}
	return items, true, truncated, nil
}

func firstLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		if ipv4 := ipNet.IP.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return ""
}

func roundFloat(value float64, precision int) float64 {
	scale := 1.0
	for i := 0; i < precision; i++ {
		scale *= 10
	}
	return float64(int(value*scale+0.5)) / scale
}
