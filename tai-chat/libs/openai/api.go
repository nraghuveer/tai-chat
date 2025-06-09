package openai

import (
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/sirupsen/logrus"
)

var ErrAPIKeyNotFound = fmt.Errorf("API key not found in environment variable")

type Provider string
type Model string

const (
	Ollama    Provider = "ollama"
	OpenAI    Provider = "openai"
	Anthropic Provider = "anthropic"
)

type AIModel struct {
	Provider          Provider `json:"provider"`
	Model             string   `json:"model"`
	APIKeyEnvVariable string   `json:"api_key_env_variable"`
	BaseAPIUrl        string   `json:"base_api_url"`
}

func (m AIModel) APIKey(logger *logrus.Entry) (string, error) {
	apiKey := os.Getenv(m.APIKeyEnvVariable)
	if apiKey == "" {
		logger.Warnf("api key not found for provider: %s and model: %s", m.Provider, m.Model)
		return "", ErrAPIKeyNotFound
	}
	return apiKey, nil
}

func (m AIModel) asOpenAIOptions(logger *logrus.Entry) ([]option.RequestOption, error) {
	apiKey, err := m.APIKey(logger)
	if err != nil {
		return nil, err
	}
	return []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(m.BaseAPIUrl),
	}, nil
}

type AIClientInterface interface {
	APIKey() string
}

// should work with all openai api compatible clients
type OpenAIClient struct {
	apiKey string
	client *openai.Client
	model  AIModel
}

func (ai *OpenAIClient) APIKey() string {
	return ai.apiKey
}

// NewAIClient return new AI Client
// always returns OpenAI compatable client for now
func NewAIClient(aiModel AIModel, logger *logrus.Entry) (AIClientInterface, error) {
	options, err := aiModel.asOpenAIOptions(logger)
	if err != nil {
		return nil, err
	}
	openaiClient := openai.NewClient(options...)
	return &OpenAIClient{apiKey: "", client: &openaiClient, model: aiModel}, nil
}
