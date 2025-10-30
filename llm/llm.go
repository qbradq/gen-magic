package llm

import (
	"fmt"
	"strings"
)

// ChatCompletion executes a chat completion with the defined LLM and given
// inputs.
func ChatCompletion(def *LanguageModel, system, prompt *Message, context []*Turn) (chan *Message, func(), error) {
	switch strings.ToLower(def.API) {
	case "openrouter":
		return openRouterChatCompletion(def, system, prompt, context)
	default:
		return nil, nil, fmt.Errorf("unknown API \"%s\"", def.API)
	}
}
