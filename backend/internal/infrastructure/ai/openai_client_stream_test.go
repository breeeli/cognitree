package ai

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestChatStreamMockSplitsAnswerIntoChunks(t *testing.T) {
	client := &OpenAIClient{}

	var chunks []string
	err := client.ChatStream(context.Background(), "tree goal context", "current question", func(delta string) error {
		chunks = append(chunks, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("ChatStream returned error: %v", err)
	}

	if len(chunks) <= 1 {
		t.Fatalf("expected multiple chunks from mock stream, got %d", len(chunks))
	}

	got := strings.Join(chunks, "")
	want := client.mockChat("tree goal context", "current question")
	if got != want {
		t.Fatalf("unexpected reconstructed stream:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestChatStreamParsesOpenAICompatibleChunks(t *testing.T) {
	var sawRequestBody string
	client := &OpenAIClient{
		baseURL:   "http://example.invalid",
		apiKey:    "test-key",
		model:     "test-model",
		maxTokens: 128,
		httpClient: &http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost {
					t.Fatalf("unexpected method: %s", req.Method)
				}
				if req.URL.Path != "/chat/completions" {
					t.Fatalf("unexpected path: %s", req.URL.Path)
				}

				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("read request body: %v", err)
				}
				sawRequestBody = string(body)

				respBody := strings.Join([]string{
					`data: {"choices":[{"delta":{"content":"Hello"}}]}`,
					``,
					`data: {"choices":[{"delta":{"content":" world"}}]}`,
					``,
					`data: [DONE]`,
					``,
				}, "\n")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
					Body:       io.NopCloser(strings.NewReader(respBody)),
					Request:    req,
				}, nil
			}),
		},
	}

	var chunks []string
	err := client.ChatStream(context.Background(), "system", "user", func(delta string) error {
		chunks = append(chunks, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("ChatStream returned error: %v", err)
	}

	if !strings.Contains(sawRequestBody, `"stream":true`) {
		t.Fatalf("expected request body to enable streaming, got %s", sawRequestBody)
	}

	got := strings.Join(chunks, "")
	if got != "Hello world" {
		t.Fatalf("unexpected reconstructed response: %q", got)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
