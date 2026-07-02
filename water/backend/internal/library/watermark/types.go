package watermark

type WatermarkConfig struct {
	Order    string  `json:"order"`
	Font     string  `json:"font"`
	FontSize int     `json:"font_size"`
	Color    string  `json:"color"`
	Opacity  float64 `json:"opacity"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Base64   string  `json:"base64"`
	Rotate   int     `json:"rotate"`
}
