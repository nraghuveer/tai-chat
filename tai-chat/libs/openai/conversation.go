package openai

import (
	"context"
	"iter"

	"github.com/openai/openai-go"
	"github.com/sirupsen/logrus"
)

type ConversationInterface interface {
	AddMessage(content string) error
	StreamResponse(ctx context.Context, logger *logrus.Entry) iter.Seq[*ConversationResponse]
	TokensUsed() int64
}

type ConversationResponse struct {
	Content  *string `json:"content,omitempty"`
	Err      error   `json:"err,omitempty"`
	Finished bool    `json:"finished"`
}

type OpenAIConversation struct {
	client     *OpenAIClient
	params     *openai.ChatCompletionNewParams
	tokensUsed int64
}

func NewOpenAIConversation(ctx context.Context, client *OpenAIClient, systemPrompt string, developerPrompt string) (ConversationInterface, error) {
	param := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.DeveloperMessage(developerPrompt),
			openai.SystemMessage(systemPrompt),
		},
		Seed:  openai.Int(1),
		Model: openai.ChatModelGPT4o,
	}
	return &OpenAIConversation{client: client, params: &param}, nil
}

func (c *OpenAIConversation) AddMessage(content string) error {
	c.params.Messages = append(c.params.Messages, openai.UserMessage(content))
	return nil
}

// StreamResponse
func (c *OpenAIConversation) StreamResponse(ctx context.Context, logger *logrus.Entry) iter.Seq[*ConversationResponse] {
	stream := c.client.client.Chat.Completions.NewStreaming(ctx, *c.params)
	acc := openai.ChatCompletionAccumulator{}

	f := func(yield func(*ConversationResponse) bool) {
		defer func() {
			c.tokensUsed += acc.Usage.TotalTokens
		}()

		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)
			if _, ok := acc.JustFinishedContent(); ok {
				if !yield(&ConversationResponse{Finished: true}) {
					return
				}
			}

			// stream the response
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				if !yield(&ConversationResponse{Content: &chunk.Choices[0].Delta.Content}) {
					return
				}
			}
		}

		if err := stream.Err(); err != nil {
			yield(&ConversationResponse{Err: err})
			return
		}
	}
	return f
}

func (c *OpenAIConversation) TokensUsed() int64 {
	return c.tokensUsed
}
