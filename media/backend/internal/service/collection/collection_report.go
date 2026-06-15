// This file normalizes net-flux metric packets into media dashboard report projections.

package collection

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/dellinger2023/net-flux/gen"
	"github.com/gogf/gf/v2/os/gtime"
)

// report status, source, and protocol values persisted by the collection writer.
const (
	reportNodeStatusHealthy reportNodeStatus = "healthy"
	reportStatusUnknown     reportTextStatus = "unknown"

	reportSourceTypeNode     reportSourceType = "node"
	reportSourceTypeInstance reportSourceType = "instance"

	reportStreamStatusUnknown   reportStreamStatus = "universal"
	reportStreamStatusRunning   reportStreamStatus = "running"
	reportStreamStatusInactive  reportStreamStatus = "inactive"
	reportStreamStatusError     reportStreamStatus = "error"
	reportStreamStatusClosed    reportStreamStatus = "closed"
	reportStreamStatusTimeout   reportStreamStatus = "timeout"
	reportStreamStatusCancelled reportStreamStatus = "cancelled"
	reportStreamStatusFailed    reportStreamStatus = "failed"

	reportStreamProtocolUnknown  reportStreamProtocol = "UNIVERSAL"
	reportStreamProtocolRTMP     reportStreamProtocol = "RTMP"
	reportStreamProtocolRTSP     reportStreamProtocol = "RTSP"
	reportStreamProtocolHLS      reportStreamProtocol = "HLS"
	reportStreamProtocolHTTPFLV  reportStreamProtocol = "HTTP_FLV"
	reportStreamProtocolWSFLV    reportStreamProtocol = "WS_FLV"
	reportStreamProtocolHTTPSFLV reportStreamProtocol = "HTTPS_FLV"
	reportStreamProtocolWSSFLV   reportStreamProtocol = "WSS_FLV"
	reportStreamProtocolGB28181  reportStreamProtocol = "GB28181"

	defaultReportParentNodeID = "0"
	defaultReportJSONMap      = "{}"
	defaultReportJSONArray    = "[]"
	bytesPerMegabyte          = 1024 * 1024
	bytesPerGigabyte          = bytesPerMegabyte * 1024
	bitsPerMegabit            = 1000 * 1000
)

type (
	reportNodeStatus     string
	reportTextStatus     string
	reportSourceType     string
	reportStreamStatus   string
	reportStreamProtocol string
)

// networkReport carries one normalized node network projection update.
type networkReport struct {
	nodeID          string
	destinationID   string
	rtt             int32
	networkOut      float64
	lastHeartbeat   *gtime.Time
	nodeLatencyMap  string
	reportTime      int64
	hasLatencyMerge bool
}

// streamReport carries one normalized stream projection update.
type streamReport struct {
	streamID              string
	sourceType            reportSourceType
	sourceID              string
	tenantID              string
	nodeID                string
	nodeName              string
	instanceID            string
	instanceName          string
	sourceURL             string
	streamName            string
	resolution            string
	fps                   float64
	bitrate               int32
	packetLoss            float64
	status                reportStreamStatus
	startTime             *gtime.Time
	duration              int32
	avgDelay              int32
	protocolCount         int
	protocolSummary       string
	totalSessionsLifetime int64
	currentSessions       int32
	lastHeartbeat         *gtime.Time
	reportTime            int64
	watermarkEnabled      bool
}

// instanceReport carries one normalized instance projection update.
type instanceReport struct {
	instanceID      string
	instanceName    string
	nodeID          string
	nodeName        string
	region          string
	nodeStatus      string
	status          string
	cpuAllocated    float64
	cpuLoad         float64
	memoryAllocated float64
	memoryUsed      float64
	diskIoRead      float64
	diskIoWrite     float64
	networkIn       float64
	networkOut      float64
	startTime       *gtime.Time
	version         string
	reportTime      int64
}

// sessionReport carries one normalized session projection update.
type sessionReport struct {
	sessionID         string
	streamID          string
	streamName        string
	tenantID          string
	clientID          string
	clientIP          string
	clientType        string
	userName          string
	protocolType      string
	startTime         *gtime.Time
	playDuration      int32
	currentFPS        float64
	currentBitrate    int32
	currentResolution string
	nodeID            string
	nodeName          string
	instanceID        string
	instanceName      string
	linkHops          string
	totalLinkLatency  int32
	reportTime        int64
}

// protocolSummaryItem is serialized into media_report_stream.protocol_summary.
type protocolSummaryItem struct {
	ProtocolType    string `json:"protocol_type"`
	TotalSessions   int    `json:"total_sessions"`
	CurrentSessions int    `json:"current_sessions"`
}

// linkHopReportItem is serialized into media_report_session.link_hops.
type linkHopReportItem struct {
	HopID    string `json:"hop_id"`
	HopName  string `json:"hop_name"`
	HopType  string `json:"hop_type"`
	NodeID   string `json:"node_id"`
	NodeName string `json:"node_name"`
	Latency  int32  `json:"latency"`
}

// normalizeMachineMetric converts one MachineMetric to an instance projection.
func normalizeMachineMetric(metric *gen.MachineMetric) (instanceReport, bool) {
	if metric == nil {
		return instanceReport{}, false
	}
	instanceID := firstNonBlank(metric.GetInstanceId(), metric.GetMachineId())
	if instanceID == "" {
		return instanceReport{}, false
	}
	reportTime := normalizeReportTime(metric.GetTimestamp())
	instanceName := strings.TrimSpace(metric.GetInstanceName())
	if instanceName == "" {
		instanceName = instanceID
	}
	nodeID := firstNonBlank(metric.GetNodeId(), metric.GetExtra()["node_id"])
	nodeName := firstNonBlank(metric.GetNodeName(), metric.GetExtra()["node_name"])
	if nodeName == "" && nodeID != "" {
		nodeName = nodeID
	}
	return instanceReport{
		instanceID:      instanceID,
		instanceName:    instanceName,
		nodeID:          nodeID,
		nodeName:        nodeName,
		region:          strings.TrimSpace(metric.GetRegion()),
		nodeStatus:      normalizeStatusText(metric.GetNodeStatus(), string(reportNodeStatusHealthy)),
		status:          normalizeStatusText(metric.GetStatus(), string(reportStatusUnknown)),
		cpuAllocated:    float64(metric.GetCpuCount()),
		cpuLoad:         metric.GetCpuUsage(),
		memoryAllocated: bytesToGigabytes(metric.GetMemTotal()),
		memoryUsed:      bytesToGigabytes(metric.GetMemUsed()),
		diskIoRead:      bytesToMegabytes(metric.GetDiskReadBytes()),
		diskIoWrite:     bytesToMegabytes(metric.GetDiskWriteBytes()),
		networkIn:       bitsToMegabits(metric.GetNetworkIn()),
		networkOut:      bitsToMegabits(metric.GetNetworkOut()),
		startTime:       metricTimeOrReportTime(metric.GetStartTime(), reportTime),
		version:         strings.TrimSpace(metric.GetVersion()),
		reportTime:      reportTime,
	}, true
}

// firstNonBlank returns the first trimmed non-empty value.
func firstNonBlank(values ...string) string {
	for _, value := range values {
		if text := strings.TrimSpace(value); text != "" {
			return text
		}
	}
	return ""
}

// normalizeStatusText returns one non-empty status value.
func normalizeStatusText(status string, fallback string) string {
	status = strings.TrimSpace(status)
	if status != "" {
		return status
	}
	return fallback
}

// metricTimeOrReportTime converts optional metric event time with report time fallback.
func metricTimeOrReportTime(metricTime int64, reportTime int64) *gtime.Time {
	if metricTime > 0 {
		return reportTimeToGTime(metricTime)
	}
	return reportTimeToGTime(reportTime)
}

// normalizeLinkHops serializes session link hops into dashboard JSON.
func normalizeLinkHops(hops []*gen.LinkHop) string {
	if len(hops) == 0 {
		return defaultReportJSONArray
	}
	items := make([]linkHopReportItem, 0, len(hops))
	for _, hop := range hops {
		if hop == nil {
			continue
		}
		items = append(items, linkHopReportItem{
			HopID:    strings.TrimSpace(hop.GetHopId()),
			HopName:  strings.TrimSpace(hop.GetHopName()),
			HopType:  strings.TrimSpace(hop.GetHopType()),
			NodeID:   strings.TrimSpace(hop.GetNodeId()),
			NodeName: strings.TrimSpace(hop.GetNodeName()),
			Latency:  hop.GetLatency(),
		})
	}
	return mustEncodeJSON(items, defaultReportJSONArray)
}

// normalizeNetworkMetric converts one NetworkMetric to a node network projection.
func normalizeNetworkMetric(metric *gen.NetworkMetric) (networkReport, bool) {
	if metric == nil {
		return networkReport{}, false
	}
	nodeID := strings.TrimSpace(metric.GetMachineId())
	if nodeID == "" {
		nodeID = strings.TrimSpace(metric.GetSourceIp())
	}
	if nodeID == "" {
		return networkReport{}, false
	}
	destinationID := strings.TrimSpace(metric.GetDestinationIp())
	latencyMap := defaultReportJSONMap
	hasLatencyMerge := false
	if destinationID != "" {
		latencyMap = mustEncodeJSON(map[string]int32{destinationID: metric.GetRtt()}, defaultReportJSONMap)
		hasLatencyMerge = true
	}
	reportTime := normalizeReportTime(metric.GetTimestamp())
	return networkReport{
		nodeID:          nodeID,
		destinationID:   destinationID,
		rtt:             metric.GetRtt(),
		networkOut:      bitsToMegabits(metric.GetThroughput()),
		lastHeartbeat:   reportTimeToGTime(reportTime),
		nodeLatencyMap:  latencyMap,
		reportTime:      reportTime,
		hasLatencyMerge: hasLatencyMerge,
	}, true
}

// normalizeStreamMetric converts one StreamMetric to a stream projection.
func normalizeStreamMetric(metric *gen.StreamMetric) (streamReport, bool) {
	if metric == nil {
		return streamReport{}, false
	}
	streamID := strings.TrimSpace(metric.GetStreamId())
	if streamID == "" {
		return streamReport{}, false
	}
	nodeID := firstNonBlank(metric.GetNodeId(), metric.GetExtra()["node_id"])
	instanceID := firstNonBlank(metric.GetInstanceId(), metric.GetMachineId())
	sourceType := reportSourceTypeNode
	sourceID := nodeID
	if instanceID != "" {
		sourceType = reportSourceTypeInstance
		sourceID = instanceID
	}
	protocol := normalizeStreamProtocol(metric.GetProtocol())
	protocolCount := 0
	currentSessions := metric.GetCurrentActiveSessions()
	protocolSummary := defaultReportJSONArray
	if protocol != reportStreamProtocolUnknown {
		protocolCount = 1
		if currentSessions <= 0 && metric.GetStatus() == gen.StreamStatus_SS_RUNNING {
			currentSessions = 1
		}
		protocolSummary = mustEncodeJSON([]protocolSummaryItem{{
			ProtocolType:    string(protocol),
			TotalSessions:   int(metric.GetTotalSessionsLifetime()),
			CurrentSessions: int(currentSessions),
		}}, defaultReportJSONArray)
	}
	reportTime := normalizeReportTime(metric.GetTimestamp())
	startTime := reportTimeToGTime(reportTime)
	if metric.GetStartTime() > 0 {
		startTime = reportTimeToGTime(metric.GetStartTime())
	}
	return streamReport{
		streamID:              streamID,
		sourceType:            sourceType,
		sourceID:              sourceID,
		tenantID:              strings.TrimSpace(metric.GetTenantId()),
		nodeID:                nodeID,
		nodeName:              firstNonBlank(metric.GetNodeName(), metric.GetExtra()["node_name"]),
		instanceID:            instanceID,
		instanceName:          strings.TrimSpace(metric.GetInstanceName()),
		sourceURL:             normalizeStreamSourceURL(metric),
		streamName:            normalizeStreamName(metric),
		resolution:            normalizeResolution(metric.GetWidth(), metric.GetHeight()),
		fps:                   metric.GetFps(),
		bitrate:               metric.GetBitrate(),
		packetLoss:            metric.GetPacketLoss(),
		status:                normalizeStreamStatus(metric.GetStatus()),
		startTime:             startTime,
		duration:              metric.GetDuration(),
		avgDelay:              metric.GetAvgDelay(),
		protocolCount:         protocolCount,
		protocolSummary:       protocolSummary,
		totalSessionsLifetime: metric.GetTotalSessionsLifetime(),
		currentSessions:       currentSessions,
		lastHeartbeat:         reportTimeToGTime(reportTime),
		reportTime:            reportTime,
		watermarkEnabled:      metric.GetWatermarkEnabled() || parseBoolExtra(metric.GetExtra(), "watermark_enabled"),
	}, true
}

// normalizeSessionMetric converts one SessionMetric to a session projection.
func normalizeSessionMetric(metric *gen.SessionMetric) (sessionReport, bool) {
	if metric == nil {
		return sessionReport{}, false
	}
	sessionID := strings.TrimSpace(metric.GetSessionId())
	if sessionID == "" {
		return sessionReport{}, false
	}
	reportTime := normalizeReportTime(metric.GetTimestamp())
	streamName := strings.TrimSpace(metric.GetStreamName())
	if streamName == "" {
		streamName = strings.TrimSpace(metric.GetStreamId())
	}
	return sessionReport{
		sessionID:         sessionID,
		streamID:          strings.TrimSpace(metric.GetStreamId()),
		streamName:        streamName,
		tenantID:          strings.TrimSpace(metric.GetTenantId()),
		clientID:          strings.TrimSpace(metric.GetClientId()),
		clientIP:          strings.TrimSpace(metric.GetClientIp()),
		clientType:        strings.TrimSpace(metric.GetClientType()),
		userName:          strings.TrimSpace(metric.GetUserName()),
		protocolType:      string(normalizeStreamProtocol(metric.GetProtocol())),
		startTime:         metricTimeOrReportTime(metric.GetStartTime(), reportTime),
		playDuration:      metric.GetPlayDuration(),
		currentFPS:        metric.GetCurrentFps(),
		currentBitrate:    metric.GetCurrentBitrate(),
		currentResolution: normalizeResolution(metric.GetCurrentWidth(), metric.GetCurrentHeight()),
		nodeID:            firstNonBlank(metric.GetNodeId(), metric.GetExtra()["node_id"]),
		nodeName:          firstNonBlank(metric.GetNodeName(), metric.GetExtra()["node_name"]),
		instanceID:        firstNonBlank(metric.GetInstanceId(), metric.GetMachineId()),
		instanceName:      strings.TrimSpace(metric.GetInstanceName()),
		linkHops:          normalizeLinkHops(metric.GetLinkHops()),
		totalLinkLatency:  metric.GetTotalLinkLatency(),
		reportTime:        reportTime,
	}, true
}

// normalizeReportTime returns a non-zero report timestamp for storage.
func normalizeReportTime(timestamp int64) int64 {
	if timestamp > 0 {
		return timestamp
	}
	return time.Now().UnixMilli()
}

// reportTimeToGTime converts second or millisecond report timestamps to GoFrame time.
func reportTimeToGTime(timestamp int64) *gtime.Time {
	if timestamp <= 0 {
		return gtime.Now()
	}
	if timestamp >= 1_000_000_000_000 {
		return gtime.NewFromTime(time.UnixMilli(timestamp))
	}
	return gtime.NewFromTime(time.Unix(timestamp, 0))
}

// bytesToGigabytes converts byte counters into GiB-style memory report values.
func bytesToGigabytes(value int64) float64 {
	if value <= 0 {
		return 0
	}
	return float64(value) / bytesPerGigabyte
}

// bytesToMegabytes converts byte counters into MiB-style disk report values.
func bytesToMegabytes(value int64) float64 {
	if value <= 0 {
		return 0
	}
	return float64(value) / bytesPerMegabyte
}

// bitsToMegabits converts bps counters into Mbps report values.
func bitsToMegabits(value int32) float64 {
	if value <= 0 {
		return 0
	}
	return float64(value) / bitsPerMegabit
}

// normalizeStreamStatus maps net-flux stream statuses into stable report values.
func normalizeStreamStatus(status gen.StreamStatus) reportStreamStatus {
	switch status {
	case gen.StreamStatus_SS_RUNNING:
		return reportStreamStatusRunning
	case gen.StreamStatus_SS_INACTIVE:
		return reportStreamStatusInactive
	case gen.StreamStatus_SS_ERROR:
		return reportStreamStatusError
	case gen.StreamStatus_SS_CLOSED:
		return reportStreamStatusClosed
	case gen.StreamStatus_SS_TIMEOUT:
		return reportStreamStatusTimeout
	case gen.StreamStatus_SS_CANCELLED:
		return reportStreamStatusCancelled
	case gen.StreamStatus_SS_FAILED:
		return reportStreamStatusFailed
	default:
		return reportStreamStatusUnknown
	}
}

// normalizeStreamProtocol maps net-flux stream protocols into stable report values.
func normalizeStreamProtocol(protocol gen.StreamProtocol) reportStreamProtocol {
	switch protocol {
	case gen.StreamProtocol_SP_RTMP:
		return reportStreamProtocolRTMP
	case gen.StreamProtocol_SP_RTSP:
		return reportStreamProtocolRTSP
	case gen.StreamProtocol_SP_HLS:
		return reportStreamProtocolHLS
	case gen.StreamProtocol_SP_HTTP_FLV:
		return reportStreamProtocolHTTPFLV
	case gen.StreamProtocol_SP_WS_FLV:
		return reportStreamProtocolWSFLV
	case gen.StreamProtocol_SP_HTTPS_FLV:
		return reportStreamProtocolHTTPSFLV
	case gen.StreamProtocol_SP_WSS_FLV:
		return reportStreamProtocolWSSFLV
	case gen.StreamProtocol_SP_GB28181:
		return reportStreamProtocolGB28181
	default:
		return reportStreamProtocolUnknown
	}
}

// normalizeStreamSourceURL chooses the most inspectable stream URL-like field.
func normalizeStreamSourceURL(metric *gen.StreamMetric) string {
	if metric == nil {
		return ""
	}
	if value := strings.TrimSpace(metric.GetStreamPath()); value != "" {
		return value
	}
	return strings.TrimSpace(metric.GetStreamAlias())
}

// normalizeStreamName chooses a stable display name for dashboard rows.
func normalizeStreamName(metric *gen.StreamMetric) string {
	if metric == nil {
		return ""
	}
	if value := strings.TrimSpace(metric.GetStreamAlias()); value != "" {
		return value
	}
	return strings.TrimSpace(metric.GetStreamId())
}

// normalizeResolution serializes positive width and height into a dashboard resolution value.
func normalizeResolution(width int32, height int32) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	return strconv.FormatInt(int64(width), 10) + "x" + strconv.FormatInt(int64(height), 10)
}

// parseBoolExtra reads optional boolean metadata from report extras.
func parseBoolExtra(extra map[string]string, key string) bool {
	if len(extra) == 0 {
		return false
	}
	value, err := strconv.ParseBool(strings.TrimSpace(extra[key]))
	return err == nil && value
}

// mustEncodeJSON serializes report JSON fields and falls back to a valid empty JSON value.
func mustEncodeJSON(value any, fallback string) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fallback
	}
	return string(data)
}
