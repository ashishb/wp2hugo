package llmhelper

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func CallLLM(ctx context.Context, model openai.ChatModel, systemPrompt string, userPrompt string) (*string, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}

	client := openai.NewClient(option.WithAPIKey(openAIKey))
	param := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Seed:  openai.Int(1),
		Model: model,
	}

	completion, err := client.Chat.Completions.New(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("error getting completion: %w", err)
	}

	response := completion.Choices[0].Message.Content
	return &response, nil
}
