package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fluffy-giggle/core/internal/domain/port"
)

func completion(content string) map[string]any {
	return map[string]any{
		"id": "chatcmpl-1", "object": "chat.completion", "model": "test-model",
		"choices": []map[string]any{{
			"index": 0, "finish_reason": "stop",
			"message": map[string]any{"role": "assistant", "content": content},
		}},
		"usage": map[string]any{"prompt_tokens": 12, "completion_tokens": 3, "total_tokens": 15},
	}
}

func capture(t *testing.T, got *map[string]any, content string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(got))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(completion(content))
	}))
}

func TestComplete(t *testing.T) {
	t.Parallel()

	t.Run("sends system, user, image, json mode and temperature", func(t *testing.T) {
		t.Parallel()

		var got map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/chat/completions", r.URL.Path)
			assert.Equal(t, "Bearer secret", r.Header.Get("Authorization"))
			require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(completion(`{"ok":true}`))
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:     server.URL + "/v1",
			APIKey:      "secret",
			Model:       "weak",
			Timeout:     time.Second,
			Temperature: 0.2,
		})
		resp, err := client.Complete(t.Context(), port.ChatRequest{
			System: "be brief", User: "read this", JSON: true,
			Image: &port.ImageData{Data: []byte{1, 2, 3}, MIME: "image/png"},
		})
		require.NoError(t, err)

		assert.Equal(t, `{"ok":true}`, resp.Content)
		assert.Equal(t, "weak", resp.Model)
		assert.Equal(t, 12, resp.PromptTokens)
		assert.Equal(t, 3, resp.CompletionTokens)

		assert.Equal(t, "weak", got["model"])
		assert.Equal(t, 0.2, got["temperature"])
		assert.Equal(t, map[string]any{"type": "json_object"}, got["response_format"])
		_, thinking := got["chat_template_kwargs"]
		assert.False(t, thinking)
		messages := got["messages"].([]any)
		require.Len(t, messages, 2)
		assert.Equal(t, "system", messages[0].(map[string]any)["role"])
		parts := messages[1].(map[string]any)["content"].([]any)
		require.Len(t, parts, 2)
		assert.Equal(t, "read this", parts[0].(map[string]any)["text"])
		image := parts[1].(map[string]any)["image_url"].(map[string]any)
		assert.Equal(t, "data:image/png;base64,AQID", image["url"])
	})

	t.Run("switches thinking off and leaves temperature to the model when asked", func(t *testing.T) {
		t.Parallel()

		var got map[string]any
		server := capture(t, &got, "ok")
		defer server.Close()

		client := New(Config{
			BaseURL:         server.URL,
			APIKey:          "k",
			Model:           "m",
			Timeout:         time.Second,
			DisableThinking: true,
			Temperature:     -1,
		})
		_, err := client.Complete(t.Context(), port.ChatRequest{User: "hi"})
		require.NoError(t, err)

		assert.Equal(t, map[string]any{"enable_thinking": false}, got["chat_template_kwargs"])
		_, sent := got["temperature"]
		assert.False(t, sent, "a negative temperature is not sent")
	})

	t.Run("retries a server error", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if calls.Add(1) == 1 {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(completion("fine"))
		}))
		defer server.Close()

		client := New(
			Config{BaseURL: server.URL, APIKey: "k", Model: "m", Timeout: 5 * time.Second, MaxRetries: 1},
		)
		resp, err := client.Complete(t.Context(), port.ChatRequest{User: "hi"})
		require.NoError(t, err)
		assert.Equal(t, "fine", resp.Content)
		assert.Equal(t, int32(2), calls.Load())
	})

	t.Run("fails after retries are spent", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := New(
			Config{BaseURL: server.URL, APIKey: "k", Model: "m", Timeout: 5 * time.Second, MaxRetries: 0},
		)
		_, err := client.Complete(t.Context(), port.ChatRequest{User: "hi"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "m:")
	})
}
