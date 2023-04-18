package entity

import (
	"errors"

	"github.com/google/uuid"
)

type ChatConfig struct {
	Model            *Model
	Temperature      float32  // 0.0 to 1.0 - how creative can chatgpt be
	TopP             float32  // 0.0 to 1.0 - how conservative can chatgpt be
	N                int      // number of messages to generate
	Stop             []string // list of tokens to stop on
	MaxTokens        int      // number of tokens to generate
	PresencePenalty  float32  // -2.0 to 2.0 - Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far, increasing the model's likelihood to talk about new topics.
	FrequencyPenalty float32  // -2.0 to 2.0 - Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far, increasing the model's likelihood to talk about new topics.
}

type Chat struct {
	ID                   string
	UserID               string
	InitialSystemMessage *Message
	Messages             []*Message
	ErasedMessages       []*Message
	Status               string
	TokenUsage           int
	Config               *ChatConfig
}

func NewChat(userID string, initialSystemMessage *Message, chatConfig *ChatConfig) (*Chat, error) {
	chat := &Chat{
		ID:                   uuid.New().String(),
		UserID:               userID,
		InitialSystemMessage: initialSystemMessage,
		Status:               "active",
		Config:               chatConfig,
		TokenUsage:           0,
	}

	chat.AddMessage(initialSystemMessage)

	if err := chat.Validate(); err != nil {
		return nil, err
	}

	return chat, nil
}

func (c *Chat) AddMessage(m *Message) error {
	if c.Status == "ended" {
		return errors.New("chat is ended. No more messages are allowed")
	}

	// loop will continue until the message is appended to the chat messages
	for {

		// if current message's token amount plus current token usage of this chat is smaller than max tokens of current model, than, the message can be added successfully
		if c.Config.Model.GetMaxTokens() >= m.GetQtdTokens()+c.TokenUsage {
			// adding current message to chat messages
			c.Messages = append(c.Messages, m)
			// increasing tokens usage amount
			c.RefreshTokenUsage()
			break
		}

		// if current message's tokens amount plus current token usage of this chat is greater then max tokens of current model, than, we need to erase messages to add new ones

		// add the older message to erased messages
		c.ErasedMessages = append(c.ErasedMessages, c.Messages[0])

		// c.Messages[1:] = ignore the first element of array and return all others
		c.Messages = c.Messages[1:]
		c.RefreshTokenUsage()
	}

	return nil
}

func (c *Chat) RefreshTokenUsage() {
	c.TokenUsage = 0

	// m is the index of current Message
	for m := range c.Messages {
		c.TokenUsage += c.Messages[m].GetQtdTokens()
	}
}

func (c *Chat) Validate() error {
	if c.UserID == "" {
		return errors.New("user id is empty")
	}

	if c.Status != "active" && c.Status != "ended" {
		return errors.New("invalid status")
	}

	if c.Config.Temperature < 0 || c.Config.Temperature > 2 {
		return errors.New("invalid temperature")
	}

	// ... more validations for config

	return nil
}

func (c *Chat) End() {
	c.Status = "ended"
}

// getters
func (c *Chat) GetMessages() []*Message {
	return c.Messages
}

func (c *Chat) CountMessages() int {
	return len(c.Messages)
}
