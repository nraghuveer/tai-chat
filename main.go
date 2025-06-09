package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nraghuveer/tai-chat/tai-chat/libs/openai"
	"github.com/sirupsen/logrus"
)

const OLLAMA_BASE_URL = "http://localhost:11434/v1"

type ConversationEvent string

const (
	Exit     ConversationEvent = "exit"
	Continue ConversationEvent = "continue"
	Finished ConversationEvent = "continue"
)

func main() {
	logger := logrus.WithField("env", "test")
	client, err := openai.NewAIClient(openai.AIModel{
		APIKeyEnvVariable: "OLLAMA_KEY",
		Provider:          openai.Ollama,
		Model:             "gemma3:1b",
		BaseAPIUrl:        OLLAMA_BASE_URL,
	}, logger)
	if err != nil {
		panic(err)
	}
	openaiClient, ok := client.(*openai.OpenAIClient)
	if !ok {
		panic(errors.New("not a openai client"))
	}
	// // get input from user for developer prompt and system prompt
	// defaultDeveloperPrompt := "You are a helpful assistant expert in computer science and programming."
	// defaultSystemPrompt := "You are a senior software engineer experienced in programming and low level computer science concepts."
	// var userDeveloperPrompt, userSystemPrompt string
	// fmt.Scan(fmt.Sprintf("Enter developer prompt (default: %s): ", defaultDeveloperPrompt), &userDeveloperPrompt)
	// if userDeveloperPrompt == "" {
	// 	userDeveloperPrompt = defaultDeveloperPrompt
	// }
	// if userSystemPrompt == "" {
	// 	userSystemPrompt = defaultSystemPrompt
	// }
	// logger.Debugf("Using developer prompt: %s", userDeveloperPrompt)
	// fmt.Printf("Using system prompt: %s\n", userSystemPrompt)

	c, err := openai.NewOpenAIConversation(context.TODO(), openaiClient, "", "")
	if err != nil {
		panic(err)
	}

	for {
		var userMessage string
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("[user] >> ")
		userMessage, err = reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		var buffer strings.Builder

		if userMessage == "exit" || userMessage == "quit" {
			fmt.Println("Used tokens:", c.TokensUsed())
			fmt.Println("Exiting conversation.")
			return
		}

		if userMessage == "" {
			panic("No user message")
		}
		getResponse(c, userMessage, buffer, logger)
		fmt.Print("\n")
	}
}

func getResponse(c openai.ConversationInterface, userMessage string, buffer strings.Builder, logger *logrus.Entry) ConversationEvent {
	if err := c.AddMessage(userMessage); err != nil {
		logger.Errorf("Error adding message: %v", err)
		return Continue
	}

	for streamRes := range c.StreamResponse(context.TODO(), logger) {
		if streamRes.Err != nil {
			logger.Errorf("Error getting response: %v", streamRes.Err)
			continue
		}
		if streamRes.Finished {
			return Finished
		}
		if streamRes.Content != nil {
			if buffer.Len() == 0 {
				fmt.Print("[ai] >> ")
			}
			buffer.WriteString(*streamRes.Content)
			fmt.Print(*streamRes.Content)
		}
	}

	if err := c.AddSystemResponse(buffer.String()); err != nil {
		logger.Errorf("Error adding system  response: %v", err)
		return Continue
	}
	return Continue
}
