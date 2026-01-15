package domain

import (
	"log/slog"
	"strings"
)

type HueWord string

type HueColor string

var allowedHueColors = map[HueColor]struct{}{
	"黒":    {},
	"灰色":   {},
	"白":    {},
	"ピンク":  {},
	"赤":    {},
	"オレンジ": {},
	"黄色":   {},
	"緑":    {},
	"青":    {},
	"紫":    {},
	"茶":    {},
}

func (c HueColor) valid() bool {
	_, ok := allowedHueColors[c]
	return ok
}

// HueChoices は単語ごとの色割り当てを保持し、空や空白キーを許可しない。
type HueChoices struct {
	values map[HueWord]HueColor
}

func NewHueChoices(raw map[string]string) (HueChoices, error) {
	if len(raw) == 0 {
		return HueChoices{}, ErrInvalidChoice
	}

	values := make(map[HueWord]HueColor, len(raw))
	for word, color := range raw {
		w := HueWord(strings.TrimSpace(word))
		c := HueColor(strings.TrimSpace(color))
		if w == "" || c == "" || !c.valid() {
			slog.Warn("Invalid hue choice value", "value", color, "word", word)
			continue
		}
		values[w] = c
	}

	return HueChoices{values: values}, nil
}

func (c HueChoices) Size() int {
	return len(c.values)
}

func (c HueChoices) ToMap() map[string]string {
	copied := make(map[string]string, len(c.values))
	for w, color := range c.values {
		copied[string(w)] = string(color)
	}

	return copied
}
