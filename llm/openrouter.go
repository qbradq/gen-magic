package llm

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/revrost/go-openrouter"
)

func openRouterChatCompletion(def *LanguageModel, system, prompt *Message, chatContext []*Turn) (chan *Message, func(), error) {
	if def.APIEndpoint == "" {
		return nil, nil, errors.New("an API endpoint is required in LLM configuration for OpenRouter")
	}
	if def.APIKey == "" {
		return nil, nil, errors.New("an API key is required in LLM configuration for OpenRouter")
	}
	if def.Model == "" {
		return nil, nil, errors.New("a model is required in LLM configuration for OpenRouter")
	}
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan *Message, 1024)
	client := openrouter.NewClient(def.APIKey)
	messages := []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(system.Content),
	}
	for _, turn := range chatContext {
		messages = append(messages, openrouter.UserMessage(turn.Prompt.Content))
		for _, msg := range turn.Response {
			switch msg.Role {
			case "user":
				messages = append(messages, openrouter.UserMessage(msg.Content))
			case "assistant":
				messages = append(messages, openrouter.AssistantMessage(msg.Content))
			default:
				slog.Error("error in openRouterChatCompletion unsupported message role in chat context", "role", msg.Role)
			}
		}
	}
	messages = append(messages, openrouter.UserMessage(prompt.Content))
	go func() {
		defer close(out)
		defer cancel()
		stream, err := client.CreateChatCompletionStream(ctx, openrouter.ChatCompletionRequest{
			Model: def.Model,
			Messages: messages,
			Stream: true,
			Usage: &openrouter.IncludeUsage{
				Include: true,
			},
		})
		if err != nil {
			slog.Error("error requesting streaming response", "error", err)
			return
		}
		defer stream.Close()
		first := true
		for {
			response, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					slog.Error("error streaming response", "error", err)
				}
				break
			}
			for _, choice := range response.Choices {
				out <- &Message{
					Role: choice.Delta.Role,
					Content: choice.Delta.Content,
					Delta: !first,
				}
				first = false
			}
		}
	}()
	return out, cancel, nil
}
