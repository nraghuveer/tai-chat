package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/nraghuveer/tai-chat/tai-chat/libs/openai"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.WithField("env", "test")
	client, err := openai.NewAIClient(openai.AIModel{}, logger)
	if err != nil {
		panic(err)
	}
	openaiClient, ok := client.(*openai.OpenAIClient)
	if !ok {
		panic(errors.New("not a openai client"))
	}
	c, err := openai.NewOpenAIConversation(context.TODO(), openaiClient, "You are expert in philosophy", "")
	if err != nil {
		panic(err)
	}
	c.AddMessage("recommend some good books")
	for msg := range c.StreamResponse(context.Background(), logger) {
		if msg.Content != nil {
			fmt.Println(*msg.Content)
		}
	}
	fmt.Printf("Done - credits used - %d", c.TokensUsed())
}
