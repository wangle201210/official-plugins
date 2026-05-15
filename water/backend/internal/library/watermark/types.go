package watermark

type Alignment int

const (
	AlignmentNothing Alignment = iota
	AlignmentLeft
	AlignmentCenter
	AlignmentRight
	AlignmentTop
	AlignmentBottom
	AlignmentTopLeft
	AlignmentTopRight
	AlignmentBottomLeft
	AlignmentBottomRight
)

type TextSetting struct {
	Text     string
	Font     string
	FontSize int
	Color    string
	PosX     int
	PosY     int
	Align    Alignment
}

type ImageSetting struct {
	Image   string
	Opacity float64
	Base64  string
}

type WatermarkConfig struct {
	TextSetting  TextSetting
	ImageSetting ImageSetting
}
