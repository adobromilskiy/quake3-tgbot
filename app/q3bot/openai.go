package q3bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

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
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "getPlayerInfo",
				Description: "Get info or how to play player X",
				Parameters:  json.RawMessage(`{"type": "object","properties":{"player":{"type":"string","description":"Player name in quake3"}}}`),
			},
		},
	}

	funcPool = map[string]func(openAIFuncArg) (string, error){
		"getLastMatchesInfo": getLastMatchesInfo,
		"getPlayerInfo":      getPlayerInfo,
	}
)

type openAIFuncArg struct {
	Number int    `json:"number"`
	Player string `json:"player"`
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

		sort.Slice(match.Players, func(i, j int) bool {
			return match.Players[i].Score > match.Players[j].Score
		})

		for _, player := range match.Players {
			kdr := float64(player.Kills) / float64(player.Deaths)
			result += fmt.Sprintf("<b>%s</b> [%d]: %d/%d/%d - <i>(%.2f)</i>\n", player.Name, player.Score, player.Kills, player.Deaths, player.Suicides, kdr)
		}

		result += "\n\n\n"
	}

	result += fmt.Sprintf("<a href=\"%s\">more info</a>", config.ServerURL)

	return result, nil
}

func getPlayerInfo(args openAIFuncArg) (string, error) {
	url := fmt.Sprintf("%s/api/ffa/players", config.ServerURL)
	resp, err := makeRequest(url, http.MethodGet, nil)
	if err != nil {
		return "", err
	}

	players := []api.Player{}
	if err := json.Unmarshal(resp, &players); err != nil {
		return "", err
	}

	var result string

	weapons := map[string]string{
		"G":   "Gauntlet",
		"MG":  "Machine Gun",
		"SG":  "Shotgun",
		"GL":  "Grenade Launcher",
		"RL":  "Rocket Launcher",
		"LG":  "Lightning Gun",
		"RG":  "Railgun",
		"PG":  "Plasma Gun",
		"BFG": "BFG",
	}

	for _, player := range players {
		if strings.EqualFold(player.Name, args.Player) {
			result += fmt.Sprintf("<b>%s's</b> stats after %d games:\n\n", player.Name, player.Games)
			rate := float64(player.Kills) / float64(player.Deaths)
			result += fmt.Sprintf("Kill-Death Ratio: %.2f\n", rate)
			rate = float64(player.Score) / float64(player.Games)
			result += fmt.Sprintf("Score per Game: %.2f\n", rate)
			rate = float64(player.Kills) / float64(player.Games)
			result += fmt.Sprintf("Kills per Game: %.2f\n", rate)
			rate = float64(player.Deaths) / float64(player.Games)
			result += fmt.Sprintf("Deaths per Game: %.2f\n", rate)
			rate = float64(player.Suicides) / float64(player.Deaths)
			result += fmt.Sprintf("Suicides Rate: %.2f\n\n\n", rate)

			for _, weapon := range player.Weapons {
				if weapon.Name == "G" {
					result += fmt.Sprintf("<b>%s</b>:\nHit Efficiency: %.2f\n\n", weapons[weapon.Name], float64(weapon.Kills)/float64(weapon.Hits)*100)
					continue
				}

				result += fmt.Sprintf("<b>%s</b>:\nAccuracy: %.2f / Hit Efficiency: %.2f\n", weapons[weapon.Name], float64(weapon.Hits)/float64(weapon.Shots)*100, float64(weapon.Kills)/float64(weapon.Hits)*100)
				result += fmt.Sprintf("Kill-Shot Ratio: %.2f\n\n", float64(weapon.Shots)/float64(weapon.Kills))
			}
		}
	}

	if result == "" {
		result = fmt.Sprintf("Player %s not found\n\n", args.Player)
	}

	result += fmt.Sprintf("<a href=\"%s\">more info</a>", config.ServerURL)

	return result, nil
}

func secondsToTime(sec uint) string {
	minutes := sec / 60
	seconds := sec % 60

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func analyzeMatchInfo(ctx context.Context, jsonData, prompt string) (result string, err error) {
	client := openai.NewClient(config.OpenAIToken)

	messages := make([]openai.ChatCompletionMessage, 0)

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: prompt,
	})

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: jsonData,
	})

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	if err != nil {
		return "", nil
	}

	return resp.Choices[0].Message.Content, nil
}
