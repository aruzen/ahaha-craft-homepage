package domain

import "github.com/google/uuid"

// HueRecord は参加者名と色割り当てをまとめた値オブジェクト。
type HueRecord struct {
	id      uuid.UUID
	name    Name
	choices HueChoices
}

// NewHueRecord は空の選択を拒否し、完全なレコードを構築する。
func NewHueRecord(name Name, choices HueChoices) (HueRecord, error) {
	if choices.Size() == 0 {
		return HueRecord{}, ErrInvalidChoice
	}

	return HueRecord{
		id:      uuid.New(),
		name:    name,
		choices: choices,
	}, nil
}

// NewHueRecordFromRaw は生文字列を正規化して HueRecord を組み立てる。
func NewHueRecordFromRaw(name string, raw map[string]string) (HueRecord, error) {
	n, err := NewName(name)
	if err != nil {
		return HueRecord{}, err
	}

	choices, err := NewHueChoices(raw)
	if err != nil {
		return HueRecord{}, err
	}

	return NewHueRecord(n, choices)
}

func (r HueRecord) ID() uuid.UUID {
	return r.id
}

func (r HueRecord) Name() Name {
	return r.name
}

func (r HueRecord) Choices() HueChoices {
	return r.choices
}

func (r HueRecord) ChoiceMap() map[string]string {
	return r.choices.ToMap()
}
