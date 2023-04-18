package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	tiktoken_go "github.com/j178/tiktoken-go"
)

type Message struct {
	ID        string    // message id
	Role      string    // user | 'system' | 'assistant
	Content   string    // message content
	Tokens    int       //amount of tokens from the content message
	Model     *Model    // chatgpt model currently being used
	CreatedAt time.Time // moment when the message was created
}

// function to instance a new message
func NewMessage(role, content string, model *Model) (*Message, error) {
	// count how many tokens exists in current content
	totalTokens := tiktoken_go.CountTokens(model.GetName(), content)

	msg := &Message{
		ID:        uuid.New().String(),
		Role:      role,
		Content:   content,
		Tokens:    totalTokens,
		Model:     model,
		CreatedAt: time.Now(),
	}

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	return msg, nil
}

// function to validate the attribute values of the message
func (m *Message) Validate() error {
	// role must be: user, system or assistant otherwise is an invalid value
	if m.Role != "user" && m.Role != "system" && m.Role != "assistant" {
		return errors.New("invalid role")
	}

	// content cannot be empty
	if m.Content == "" {
		return errors.New("content is empty")
	}

	// check if CreatedAt is empty
	if m.CreatedAt.IsZero() {
		return errors.New("invalid created at")
	}

	return nil
}

// return current value of Tokens attribute
func (m *Message) GetQtdTokens() int {
	return m.Tokens
}
