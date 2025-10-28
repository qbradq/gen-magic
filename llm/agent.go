package llm

// Agent holds the data needed to maintain and automate a chat.
type Agent struct {
	ID int
	Name string
	Definition *Definition
	System *Message
}
