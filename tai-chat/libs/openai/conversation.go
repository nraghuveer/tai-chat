package openai

import (
	"context"
	"iter"

	"github.com/openai/openai-go"
)

type ConversationInterface interface {
	AddMessage(content string) error
	StreamResponse(ctx context.Context) (iter.Seq[*ConversationResponse], error)
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

func (c *OpenAIConversation) StreamResponse(ctx context.Context) (iter.Seq[*ConversationResponse], error) {
	stream := c.client.client.Chat.Completions.NewStreaming(ctx, *c.params)
	acc := openai.ChatCompletionAccumulator{}

	f := func(yield func(*ConversationResponse) bool) {
		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)
			if _, ok := acc.JustFinishedContent(); ok {
				if !yield(&ConversationResponse{Finished: true}) {
					c.tokensUsed += acc.Usage.TotalTokens
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
			c.tokensUsed += acc.Usage.TotalTokens
			yield(&ConversationResponse{Err: err})
			return
		}

		c.tokensUsed += acc.Usage.TotalTokens
	}
	return f, nil
}

func (c *OpenAIConversation) TokensUsed() int64 {
	return c.tokensUsed
}
