// This file defines water service request, response, and domain types.

package water

import (
	"context"

	"lina-plugin-water/backend/internal/library/watermark"
)

// StrategyResolver defines the media strategy lookup contract consumed by water.
type StrategyResolver interface {
	// ResolveStrategy resolves the effective media strategy for one tenant/device pair.
	ResolveStrategy(ctx context.Context, in ResolveStrategyInput) (*ResolveStrategyOutput, error)
}

// ResolveStrategyInput defines one strategy lookup request.
type ResolveStrategyInput struct {
	TenantId string // TenantId is the media tenant ID.
	DeviceId string // DeviceId is the GB device ID.
}

// ResolveStrategyOutput defines one resolved media strategy projection.
type ResolveStrategyOutput struct {
	Matched      bool   // Matched reports whether a strategy matched.
	Source       string // Source is the matching strategy source.
	SourceLabel  string // SourceLabel is the source label prepared by media.
	StrategyId   int64  // StrategyId is the matched strategy ID.
	StrategyName string // StrategyName is the matched strategy name.
	Strategy     string // Strategy is the matched YAML strategy body.
}

// SubmitSnapInput defines one asynchronous snapshot request.
type SubmitSnapInput struct {
	DeviceType  string // DeviceType is the route-level device type.
	DeviceId    string // DeviceId is the route-level device ID.
	ErrorCode   string // ErrorCode is the upstream error code.
	DeviceCode  string // DeviceCode is the GB device identifier.
	ChannelCode string // ChannelCode is the channel identifier.
	DeviceIdx   string // DeviceIdx is the upstream device index.
	Image       string // Image is a base64 image or data URL.
	ImageName   string // ImageName is the upstream image name.
	ImagePath   string // ImagePath is the upstream image path.
	AccessNode  string // AccessNode is the upstream access node.
	AcceptNode  string // AcceptNode is the upstream accept node.
	UploadUrl   string // UploadUrl is the upstream upload URL.
	CallbackUrl string // CallbackUrl receives the processing result.
	Url         string // Url is the hotgo-compatible callback URL field.
	User        string // User is the business user ID.
	Tenant      string // Tenant is the media business tenant ID.
}

// SubmitSnapOutput defines the asynchronous submit response.
type SubmitSnapOutput struct {
	Success bool       // Success reports whether the task was queued.
	TaskId  string     // TaskId is the generated task identifier.
	Status  TaskStatus // Status is the initial task status.
}

// PreviewInput defines one synchronous watermark preview request.
type PreviewInput struct {
	Tenant      string // Tenant is the media business tenant ID.
	DeviceId    string // DeviceId is the media device ID.
	DeviceCode  string // DeviceCode is the GB device identifier.
	ChannelCode string // ChannelCode is the channel identifier.
	Image       string // Image is a base64 image or data URL.
}

// ProcessOutput defines a processed watermark result.
type ProcessOutput struct {
	Success      bool           // Success reports whether processing completed.
	Status       TaskStatus     // Status is the processing status.
	Message      string         // Message is a Chinese status description.
	Error        string         // Error contains a user-visible error detail.
	Image        string         // Image is the output PNG data URL.
	StrategyId   int64          // StrategyId is the matched media strategy ID.
	StrategyName string         // StrategyName is the matched media strategy name.
	Source       StrategySource // Source is the matched strategy source.
	SourceLabel  string         // SourceLabel is the Chinese source label.
	DurationMs   int64          // DurationMs is the processing duration in milliseconds.
}

// TaskSnapshot defines one recent task status snapshot.
type TaskSnapshot struct {
	TaskId       string         // TaskId is the generated task identifier.
	Status       TaskStatus     // Status is the processing status.
	Success      bool           // Success reports whether processing completed successfully.
	Message      string         // Message is a Chinese status description.
	Error        string         // Error contains a user-visible error detail.
	Tenant       string         // Tenant is the media business tenant ID.
	DeviceId     string         // DeviceId is the route-level device ID.
	StrategyId   int64          // StrategyId is the matched media strategy ID.
	StrategyName string         // StrategyName is the matched media strategy name.
	Source       StrategySource // Source is the matched strategy source.
	SourceLabel  string         // SourceLabel is the Chinese source label.
	Image        string         // Image stays empty for async status snapshots so cache values remain bounded.
	CreatedAt    int64          // CreatedAt is the Unix timestamp in milliseconds.
	UpdatedAt    int64          // UpdatedAt is the Unix timestamp in milliseconds.
	DurationMs   int64          // DurationMs is the processing duration in milliseconds.
}

// snapPayload is the callback-compatible snapshot payload.
type snapPayload struct {
	ErrorCode   string `json:"error_code"`
	DeviceCode  string `json:"deviceCode"`
	ChannelCode string `json:"channelCode"`
	DeviceIdx   string `json:"deviceIdx"`
	Image       string `json:"image"`
	ImageName   string `json:"imageName"`
	ImagePath   string `json:"imagePath"`
	AccessNode  string `json:"accessNode"`
	AcceptNode  string `json:"acceptNode"`
	UploadUrl   string `json:"uploadUrl"`
}

// resolvedStrategy contains the effective media strategy and source metadata.
type resolvedStrategy struct {
	Matched      bool           // Matched reports whether a strategy matched.
	Source       StrategySource // Source is the matching strategy source.
	SourceLabel  string         // SourceLabel is the Chinese matching source label.
	StrategyId   int64          // StrategyId is the matched media strategy ID.
	StrategyName string         // StrategyName is the matched media strategy name.
	Strategy     string         // Strategy is the YAML strategy content.
}

// watermarkConfig defines the normalized snapshot watermark rendering configuration.
type watermarkConfig struct {
	Text     string             `json:"text" yaml:"text"`         // Text is the watermark text.
	Font     string             `json:"font" yaml:"font"`         // Font is an optional font file path.
	FontSize int                `json:"fontSize" yaml:"fontSize"` // FontSize is the text size in pixels.
	Color    string             `json:"color" yaml:"color"`       // Color is a hex color such as #ffffff.
	PosX     int                `json:"posX" yaml:"posX"`         // PosX is an optional absolute x coordinate.
	PosY     int                `json:"posY" yaml:"posY"`         // PosY is an optional absolute y coordinate.
	Align    watermarkAlignment `json:"align" yaml:"align"`       // Align is the named or numeric alignment value.
	Image    string             `json:"image" yaml:"image"`       // Image is an optional watermark image path.
	Opacity  float64            `json:"opacity" yaml:"opacity"`   // Opacity controls text and image alpha.
	Base64   string             `json:"base64" yaml:"base64"`     // Base64 is an optional watermark image data URL.
}

// ToWatermarkConfig converts parsed strategy YAML into the migrated HotGo watermark library config.
func (c *watermarkConfig) ToWatermarkConfig() watermark.WatermarkConfig {
	if c == nil {
		return watermark.WatermarkConfig{}
	}
	return watermark.WatermarkConfig{
		Order:    c.Text,
		Font:     c.Font,
		FontSize: c.FontSize,
		Color:    c.Color,
		Opacity:  c.Opacity,
		Base64:   normalizeWatermarkBase64(c.Base64),
	}
}
