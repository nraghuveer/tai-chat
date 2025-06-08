package openai

import (
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/sirupsen/logrus"
)

var ErrAPIKeyNotFound = fmt.Errorf("API key not found in environment variable")

type Provider string
type Model string

const (
	OpenAI    Provider = "openai"
	Anthropic Provider = "anthropic"
)

type AIModel struct {
	Provider          Provider `json:"provider"`
	Model             Model    `json:"model"`
	APIKeyEnvVariable string   `json:"api_key_env_variable"`
}

type AIClientInterface interface {
	APIKey() string
}

// should work with all openai api compatible clients
type OpenAIClient struct {
	apiKey string
	client *openai.Client
}

func (ai *OpenAIClient) APIKey() string {
	return ai.apiKey
}

// for now just return the OpenAI client
func NewAIClient(aiMode AIModel, logger *logrus.Entry) (AIClientInterface, error) {
	// get the API key from environment variable
	apiKey := os.Getenv(aiMode.APIKeyEnvVariable)
	if apiKey == "" {
		logger.Infof("API key not found for provider: %s and model: %s", aiMode.Provider, aiMode.Model)
		return nil, ErrAPIKeyNotFound
	}
	openaiClient := openai.NewClient(openai.DefaultClientOptions()...)
	return &OpenAIClient{apiKey: apiKey, client: &openaiClient}, nil
}
