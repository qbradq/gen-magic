package llm

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/revrost/go-openrouter"
)

func openRouterChatCompletion(def *Definition, system, prompt *Message, chatContext []*Turn) (chan *Message, func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan *Message, 1024)
	client := openrouter.NewClient(def.APIKey)
	messages := []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(system.Content),
	}
	for _, turn := range chatContext {
		for _, msg := range turn.Response {
			switch msg.Role {
			case "user":
				messages = append(messages, openrouter.UserMessage(msg.Content))
			case "assistant":
				messages = append(messages, openrouter.AssistantMessage(msg.Content))
			default:
				log.Printf("error in openRouterChatCompletion unsupported message role in chat context %s\n", msg.Role)
			}
		}
	}
	messages = append(messages, openrouter.UserMessage(prompt.Content))
	go func() {
		defer close(out)
		stream, err := client.CreateChatCompletionStream(ctx, openrouter.ChatCompletionRequest{
			Model: def.Model,
			Messages: messages,
			Stream: true,
			Usage: &openrouter.IncludeUsage{
				Include: true,
			},
		})
		if err != nil {
			log.Printf("error requesting streaming response: %v\n", err)
		}
		defer stream.Close()
		first := true
		for {
			response, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					log.Printf("error streaming response: %v\n", err)
				}
				break
			}
			for _, choice := range response.Choices {
				// log.Println(choice.Delta.Content)
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
