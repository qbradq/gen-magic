package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type ollamaMessage struct {
	Role string `json:"role"`
	Content string `json:"content"`
}

type ollamaRequest struct {
	Model string `json:"model"`
	Messages []ollamaMessage `json:"messages"`
}

type ollamaResponseMessage struct {
	Role string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Error string `json:"error"`
	Message ollamaResponseMessage `json:"message"`
	Done bool `json:"done"`
	DoneReason string `json:"done_reason"`
}

func ollamaChatCompletion(def *LanguageModel, system, prompt *Message, chatContext []*Turn) (chan *Message, func(), error) {
	if def.APIEndpoint == "" {
		return nil, nil, errors.New("an API Endpoint is require for LLM provider Ollama")
	}
	if !strings.Contains(def.APIEndpoint, "localhost") && def.APIKey == "" {
		return nil, nil, errors.New("an API Key is required for remove Ollama endpoints")
	}
	if def.Model == "" {
		return nil, nil, errors.New("a model is require for LLM provider Ollama")
	}
	oReq := ollamaRequest{
		Model: def.Model,
		Messages: []ollamaMessage{
			{
				Role: "system",
				Content: system.Content,
			},
		},
	}
	for _, t := range chatContext {
		oReq.Messages = append(oReq.Messages, ollamaMessage{
			Role: "user",
			Content: t.Prompt.Content,
		})
		for _, m := range t.Response {
			oReq.Messages = append(oReq.Messages, ollamaMessage{
				Role: m.Role,
				Content: m.Content,
			})
		}
	}
	oReq.Messages = append(oReq.Messages, ollamaMessage{
		Role: "user",
		Content: prompt.Content,
	})
	payload, err := json.Marshal(&oReq)
	if err != nil {
		return nil, nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan *Message, 1024)
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		def.APIEndpoint + "/api/chat",
		bytes.NewReader(payload),
	)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	go func(){
		defer close(out)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			slog.Error("error requesting streaming response", "error", err)
			return
		}
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		first := true
		for {
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					out <- &Message{
						Error: fmt.Errorf("error scanning Ollama API output: %w", err),
					}
				}
				break
			}
			line := scanner.Text()
			slog.Debug("Ollama API output", "message", line)
			response := ollamaResponse{}
			err := json.Unmarshal([]byte(line), &response)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					out <- &Message{
						Error: fmt.Errorf("error decoding Ollama API response: %s", response.Error),
					}
				}
				break
			}
			if response.Error != "" {
				out <- &Message{
					Error: fmt.Errorf("error response from Ollama API: %s", response.Error),
				}
				break
			}
			if response.Message.Content != "" {
				out <- &Message{
					Role: response.Message.Role,
					Content: response.Message.Content,
					Delta: !first,
				}
				first = false
			}
			if response.Done && response.DoneReason != "load" {
				break
			}
		}
	}()
	return out, cancel, nil
}
