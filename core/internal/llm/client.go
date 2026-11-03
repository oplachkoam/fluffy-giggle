// Package llm talks to any OpenAI-compatible chat endpoint. The same client serves
// all three roles; which side of the boundary a call is on is the pipeline's concern.
package llm

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"

	"fluffy-giggle/core/internal/domain/port"
)

type Config struct {
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	MaxRetries int
	// DisableThinking sends chat_template_kwargs.enable_thinking=false, the switch
	// Qwen-style servers understand.
	DisableThinking bool
	// Temperature is sent only when it is not negative; some models reject any value
	// but their default.
	Temperature float64
}

type Client struct {
	api         openai.Client
	model       string
	temperature float64
	opts        []option.RequestOption
}

var _ port.Model = (*Client)(nil)

func New(cfg Config) *Client {
	return &Client{
		api: openai.NewClient(
			option.WithBaseURL(cfg.BaseURL),
			option.WithAPIKey(cfg.APIKey),
			option.WithRequestTimeout(cfg.Timeout),
			option.WithMaxRetries(cfg.MaxRetries),
		),
		model:       cfg.Model,
		temperature: cfg.Temperature,
		opts:        requestOptions(cfg),
	}
}

func requestOptions(cfg Config) []option.RequestOption {
	if !cfg.DisableThinking {
		return nil
	}
	return []option.RequestOption{
		option.WithJSONSet("chat_template_kwargs", map[string]any{"enable_thinking": false}),
	}
}

func (c *Client) Complete(ctx context.Context, req port.ChatRequest) (*port.ChatResponse, error) {
	var messages []openai.ChatCompletionMessageParamUnion
	if req.System != "" {
		messages = append(messages, openai.SystemMessage(req.System))
	}
	if req.Image != nil {
		url := "data:" + req.Image.MIME + ";base64," + base64.StdEncoding.EncodeToString(req.Image.Data)
		messages = append(messages, openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
			openai.TextContentPart(req.User),
			openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{URL: url}),
		}))
	} else {
		messages = append(messages, openai.UserMessage(req.User))
	}

	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(c.model),
		Messages: messages,
	}
	if c.temperature >= 0 {
		params.Temperature = openai.Float(c.temperature)
	}
	if req.JSON {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
		}
	}

	completion, err := c.api.Chat.Completions.New(ctx, params, c.opts...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", c.model, err)
	}
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("%s: %w", c.model, errors.New("empty completion"))
	}

	return &port.ChatResponse{
		Content:          completion.Choices[0].Message.Content,
		Model:            c.model,
		PromptTokens:     int(completion.Usage.PromptTokens),
		CompletionTokens: int(completion.Usage.CompletionTokens),
	}, nil
}
