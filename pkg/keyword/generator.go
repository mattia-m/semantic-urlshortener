package keyword

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Generator struct {
	client *openai.Client
}

func NewGenerator(apiKey string) *Generator {
	return &Generator{
		client: openai.NewClient(apiKey),
	}
}

func (g *Generator) GenerateKeyword(ctx context.Context, title, description, keywords string) (string, error) {
	// Prepare the prompt
	prompt := fmt.Sprintf(
		"Generate a single, simple English word (noun) that best describes this website.\n\n" +
		"Title: %s\n" +
		"Description: %s\n" +
		"Keywords: %s\n\n" +
		"Rules:\n" +
		"1. Return ONLY the word, nothing else\n" +
		"2. Word must be a simple noun\n" +
		"3. Word must be lowercase\n" +
		"4. No special characters or spaces\n" +
		"5. Maximum 15 characters\n",
		title, description, keywords,
	)

	// Create completion request
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a URL keyword generator. You generate single-word keywords that describe websites.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   10,
		Temperature: 0.3,
	}

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get completion
	resp, err := g.client.CreateChatCompletion(ctxWithTimeout, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate keyword: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no keyword generated")
	}

	// Clean and validate the response
	keyword := strings.TrimSpace(resp.Choices[0].Message.Content)
	keyword = strings.ToLower(keyword)

	// Basic validation
	if len(keyword) > 15 || strings.Contains(keyword, " ") || strings.Contains(keyword, "\n") {
		return "", fmt.Errorf("invalid keyword generated: %s", keyword)
	}

	return keyword, nil
}
