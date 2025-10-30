package llm

// Agent holds the data needed to maintain and automate a chat.
type Agent struct {
	ID int64
	Name string
	LLM *LanguageModel
	System Message
}
