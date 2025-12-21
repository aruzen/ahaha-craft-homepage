package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type HueSaveService struct {
	hueRepo *repository.HueRepository
	logger  *log.Logger
}

func NewHueSaveService(hueRepo *repository.HueRepository, logger *log.Logger) *HueSaveService {
	if logger == nil {
		logger = log.Default()
	}
	return &HueSaveService{hueRepo: hueRepo, logger: logger}
}

func (s *HueSaveService) SaveResult(ctx context.Context, record domain.HueRecord) (domain.HueResult, error) {
	if err := s.hueRepo.Save(ctx, record); err != nil {
		s.logger.Printf("[HueSaveService] save hue record: %v", err)
		return domain.HueResult{}, err
	}

	const endpoint = "https://api.openai.com/v1/responses"
	const apiKey = "dummy-api-key"

	system := strings.Replace(`
あなたは心理テスト「Hue Are You」の結果生成AIです。
与えられた人名と、各ワードに対して選択された色から
心理的特徴を分析し、
最終的なrgb値（0〜255）と、
2〜4文程度の日本語メッセージを返してください。
	`, "\n", "", -1)

	user := ""
	for k, v := range record.Choices().ToMap() {
		if user == "" {
			user = fmt.Sprintf("選択 : (語彙, 色) = (%v, %v)", k, v)
		} else {
			user += fmt.Sprintf(", (%v, %v)", k, v)
		}
	}

	payload := map[string]interface{}{
		"model": "gpt-4.1",
		"input": []map[string]interface{}{
			{
				"role": "system",
				"content": []map[string]string{
					{"type": "input_text", "text": system},
				},
			},
			{
				"role": "user",
				"content": []map[string]string{
					{"type": "input_text", "text": user},
				},
			},
		},
		"text": map[string]interface{}{
			"format": map[string]interface{}{
				"type": "json_schema",
				"name": "HueAreYouResultResponse",
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"hue": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"r": map[string]interface{}{"type": "integer", "minimum": 0, "maximum": 255},
								"g": map[string]interface{}{"type": "integer", "minimum": 0, "maximum": 255},
								"b": map[string]interface{}{"type": "integer", "minimum": 0, "maximum": 255},
							},
							"required":             []string{"r", "g", "b"},
							"additionalProperties": false,
						},
						"message": map[string]interface{}{"type": "string"},
					},
					"required":             []string{"hue", "message"},
					"additionalProperties": false,
				},
				"strict": true,
			},
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return domain.HueResult{}, err
	}

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	fmt.Println("status:", res.Status)

	body, _ := io.ReadAll(res.Body)

	var raw struct {
		Output []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return domain.HueResult{}, err
	}

	if len(raw.Output) == 0 || len(raw.Output[0].Content) == 0 {
		return domain.HueResult{}, errors.New("no content returned")
	}

	jsonText := raw.Output[0].Content[0].Text

	var answer struct {
		Hue struct {
			R int `json:"r"`
			G int `json:"g"`
			B int `json:"b"`
		} `json:"hue"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(jsonText), &answer); err != nil {
		return domain.HueResult{}, err
	}

	hue, err := domain.NewHueRGB(answer.Hue.R, answer.Hue.G, answer.Hue.B)
	if err != nil {
		return domain.HueResult{}, err
	}

	result, err := domain.NewHueResult(hue, "saved")
	if err != nil {
		return domain.HueResult{}, err
	}

	return result, nil
}
