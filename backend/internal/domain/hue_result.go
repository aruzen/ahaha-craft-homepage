package domain

import "strings"

// HueRGB はレスポンスで返す RGB 値を保持する。
type HueRGB struct {
	r int
	g int
	b int
}

// NewHueRGB は 0-255 の範囲チェックを行う。
func NewHueRGB(r, g, b int) (HueRGB, error) {
	if !validRGB(r) || !validRGB(g) || !validRGB(b) {
		return HueRGB{}, ErrInvalidHueResult
	}

	return HueRGB{r: r, g: g, b: b}, nil
}

func validRGB(v int) bool {
	return v >= 0 && v <= 255
}

func (h HueRGB) R() int { return h.r }
func (h HueRGB) G() int { return h.g }
func (h HueRGB) B() int { return h.b }

// HueResult は結果確認画面で表示する色とメッセージをまとめる。
type HueResult struct {
	hue     HueRGB
	message string
}

// NewHueResult は hue の妥当性とメッセージの空チェックを行う。
func NewHueResult(hue HueRGB, message string) (HueResult, error) {
	m := strings.TrimSpace(message)
	if m == "" {
		return HueResult{}, ErrInvalidHueResult
	}

	return HueResult{hue: hue, message: m}, nil
}

// NewHueResultFromRaw は生の値から HueResult を組み立てる。
func NewHueResultFromRaw(r, g, b int, message string) (HueResult, error) {
	hue, err := NewHueRGB(r, g, b)
	if err != nil {
		return HueResult{}, err
	}

	return NewHueResult(hue, message)
}

func (r HueResult) Hue() HueRGB     { return r.hue }
func (r HueResult) Message() string { return r.message }
