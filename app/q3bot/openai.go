package q3bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/adobromilskiy/quake3-stats/app/api"
	"github.com/sashabaranov/go-openai"
)

var (
	tools = []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "getLastMatchesInfo",
				Description: "Get info about last N matches",
				Parameters:  json.RawMessage(`{"type": "object","properties":{"number":{"type":"integer","description":"Number of matches in quake3"}}}`),
			},
		},
	}

	funcPool = map[string]func(openAIFuncArg) (string, error){
		"getLastMatchesInfo": getLastMatchesInfo,
	}
)

type openAIFuncArg struct {
	Number int `json:"number"`
}

func openAIResponse(ctx context.Context, question string) (string, error) {
	client := openai.NewClient(config.OpenAIToken)

	messages := make([]openai.ChatCompletionMessage, 0)

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "Your a helpfull assistant. When asked for you name, you must repond with 'Quake3 Bot'. Your answer with no more than 50 words.",
	})

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:      openai.GPT3Dot5Turbo,
			Messages:   messages,
			Tools:      tools,
			ToolChoice: "auto",
		},
	)

	if err != nil {
		return "", nil
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	result := resp.Choices[0].Message.Content

	if resp.Choices[0].Message.ToolCalls != nil {
		var args openAIFuncArg
		err := json.Unmarshal([]byte(resp.Choices[0].Message.ToolCalls[0].Function.Arguments), &args)
		if err != nil {
			return "", fmt.Errorf("failed to decode arguments from openAI response: %w", err)
		}

		functionName := resp.Choices[0].Message.ToolCalls[0].Function.Name
		if function, ok := funcPool[functionName]; ok {
			result, err = function(args)
			if err != nil {
				return "", fmt.Errorf("failed to execute function %s: %w", functionName, err)
			}
		}
	}

	return result, nil
}

func getLastMatchesInfo(args openAIFuncArg) (string, error) {
	url := fmt.Sprintf("%s/api/ffa/matches?perpage=%d", config.ServerURL, args.Number)
	resp, err := makeRequest(url, http.MethodGet, nil)
	if err != nil {
		return "", err
	}

	matches := []api.Match{}

	if err := json.Unmarshal(resp, &matches); err != nil {
		return "", err
	}

	var result string

	for _, match := range matches {
		result += fmt.Sprintf("<b>%s</b> <i>(%s)</i>\n\n", match.Map, secondsToTime(match.Duration))

		for _, player := range match.Players {
			kdr := float64(player.Kills) / float64(player.Deaths)
			result += fmt.Sprintf("<b>%s</b>: %d/%d/%d - <i>(%.2f)</i>\n", player.Name, player.Kills, player.Deaths, player.Suicides, kdr)
		}

		result += "\n\n\n"
	}

	return result, nil
}

func secondsToTime(sec uint) string {
	minutes := sec / 60
	seconds := sec % 60

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
