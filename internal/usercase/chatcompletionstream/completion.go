package chatcompletionstream

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/aureanemoraes/chatservice/internal/domain/entity"
	"github.com/aureanemoraes/chatservice/internal/domain/gateway"
	"github.com/sashabaranov/go-openai"
)

// data type
type ChatCompletionConfigInputDTO struct {
	Model                string
	ModelMaxTokens       int
	Temperature          float32
	TopP                 float32
	N                    int
	Stop                 []string
	MaxTokens            int
	PresencePenalty      float32
	FrequencyPenalty     float32
	InitialSystemMessage string
}

// data type to initialize an completion
type ChatCompletionInputDTO struct {
	ChatID      string
	UserID      string
	UserMessage string
	Config      ChatCompletionConfigInputDTO
}

// data type of completion's output
type ChatCompletionOutputDTO struct {
	ChatID  string
	UserID  string
	Content string
}

// data type of main struct of this file
type ChatCompletionUseCase struct {
	ChatGateway  gateway.ChatGateway
	OpenAiClient *openai.Client
	Stream       chan ChatCompletionOutputDTO
}

// function to instantiate a completion
func NewChatCompletionUseCase(chatGateway gateway.ChatGateway, openAiClient *openai.Client, stream chan ChatCompletionOutputDTO) *ChatCompletionUseCase {
	return &ChatCompletionUseCase{
		ChatGateway:  chatGateway,
		OpenAiClient: openAiClient,
		Stream:       stream,
	}
}

func (uc *ChatCompletionUseCase) Execute(ctx context.Context, input ChatCompletionInputDTO) (*ChatCompletionOutputDTO, error) {
	chat, err := uc.ChatGateway.FindChatByID(ctx, input.ChatID)

	// Validations
	if err != nil {
		if err.Error() == "chat not found" {
			// create a new entity
			chat, err = CreateNewChat(input)

			if err != nil {
				return nil, errors.New("error creating new chat: " + err.Error())
			}

			// save on database
			err = uc.ChatGateway.CreateChat(ctx, chat)

			if err != nil {
				return nil, errors.New("error persisting new chat: " + err.Error())
			}
		} else {
			return nil, errors.New("error fetching existing chat: " + err.Error())
		}
	}

	// creating user message
	userMessage, err := entity.NewMessage("user", input.UserMessage, chat.Config.Model)

	if err != nil {
		return nil, errors.New("error creating user message: " + err.Error())
	}

	// adding message to chat
	err = chat.AddMessage(userMessage)

	if err != nil {
		return nil, errors.New("error adding new message: " + err.Error())
	}

	// instantiating messages to send to chat gpt
	messages := []openai.ChatCompletionMessage{}

	for _, msg := range chat.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// sending message in format of stream to chat gpt
	resp, err := uc.OpenAiClient.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:            chat.Config.Model.Name,
			Messages:         messages,
			MaxTokens:        chat.Config.MaxTokens,
			Temperature:      chat.Config.Temperature,
			TopP:             chat.Config.TopP,
			PresencePenalty:  chat.Config.PresencePenalty,
			FrequencyPenalty: chat.Config.FrequencyPenalty,
			Stop:             chat.Config.Stop,
			Stream:           true,
		},
	)

	if err != nil {
		return nil, errors.New("error creating chat completion: " + err.Error())
	}

	var fullResponse strings.Builder

	// building response from stream
	for {
		// getting response from openia api
		response, err := resp.Recv()

		// checking if response message is ended
		if errors.Is(err, io.EOF) {
			break
		}

		// checking if is another error not mapped
		if err != nil {
			return nil, errors.New("error streaming response: " + err.Error())
		}

		fullResponse.WriteString(response.Choices[0].Delta.Content)

		r := ChatCompletionOutputDTO{
			ChatID:  chat.ID,
			UserID:  input.UserID,
			Content: fullResponse.String(),
		}

		// sending every buffer to the channel to work concurrently
		uc.Stream <- r
	}

	// when all buffers from stream response is received, instantiate a new message from user "assistant"
	assistant, err := entity.NewMessage("assistant", fullResponse.String(), chat.Config.Model)

	if err != nil {
		return nil, errors.New("error creating assistant message: " + err.Error())
	}

	// saving the created assistant message entity
	err = chat.AddMessage(assistant)

	if err != nil {
		return nil, errors.New("error adding new message: " + err.Error())
	}

	// saving on database chat with messages updated
	err = uc.ChatGateway.SaveChat(ctx, chat)

	if err != nil {
		return nil, errors.New("error saving chat: " + err.Error())
	}

	return &ChatCompletionOutputDTO{
		ChatID:  chat.ID,
		UserID:  input.UserID,
		Content: fullResponse.String(),
	}, nil
}

func CreateNewChat(input ChatCompletionInputDTO) (*entity.Chat, error) {
	model := entity.NewModel(input.Config.Model, input.Config.ModelMaxTokens)

	chatConfig := &entity.ChatConfig{
		Temperature:      input.Config.Temperature,
		TopP:             input.Config.TopP,
		N:                input.Config.N,
		Stop:             input.Config.Stop,
		MaxTokens:        input.Config.MaxTokens,
		PresencePenalty:  input.Config.PresencePenalty,
		FrequencyPenalty: input.Config.FrequencyPenalty,
		Model:            model,
	}

	initialMessage, err := entity.NewMessage("system", input.Config.InitialSystemMessage, model)

	if err != nil {
		return nil, errors.New("error creating initial message: " + err.Error())
	}

	chat, err := entity.NewChat(input.UserID, initialMessage, chatConfig)

	if err != nil {
		return nil, errors.New("error creating new chat: " + err.Error())
	}

	return chat, nil
}
