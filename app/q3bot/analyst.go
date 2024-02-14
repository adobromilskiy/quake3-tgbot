package q3bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/adobromilskiy/quake3-stats/app/api"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Analyze(ctx context.Context, b *bot.Bot) {
	var lastMatchID primitive.ObjectID
	ticker := time.NewTicker(config.Interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				match, err := getLastMatch()
				if err != nil {
					log.Printf("[Q3BOT] [ERROR] failed to get last match: %s", err)
					continue
				}

				if lastMatchID == match.ID {
					continue
				}

				lastMatchID = match.ID

				// prompt := "You are a game commentator. Please, analyze match about Quake 3 game and make short summary about each player in comedian style with gaslighting. PLEASE REPLY SHORTLY LESS THAN 100 WORDS IN RUSSIAN LANGUAGE. WINNER IS THE ONE WHO HAS THE MOST SCORES."
				prompt := "Imagine you're commenting on the outcome of a Quake 3 game. Winner is the one who has the most scores. YOUR COMMENT SHOULD BE WRITTEN IN A SARCASTIC TONE IN RUSSIAN LANGUAGE AND BE UNDER 500 CHARACTERS."

				response, err := analyzeMatchInfo(ctx, createMatchInfo(match), prompt)
				if err != nil {
					log.Printf("[Q3BOT] [ERROR] failed to analyze match info: %s", err)
					continue
				}

				log.Println("RESPONSE", len(response))

				if len(response) > 1024 {
					_, err = b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID:                config.ChatID,
						Text:                  response,
						ParseMode:             "HTML",
						DisableWebPagePreview: true,
					})

					if err != nil {
						log.Printf("[Q3BOT] [ERROR] failed to send message: %s", err)
					}

					continue
				}

				prompt = fmt.Sprintf("Create an image based on next description:\n\n %s", response)

				image, err := generateImage(ctx, prompt)
				if err != nil {
					log.Printf("[Q3BOT] [ERROR] failed to generate image: %s", err)
					continue
				}

				params := &bot.SendPhotoParams{
					ChatID:  config.ChatID,
					Photo:   &models.InputFileString{Data: image},
					Caption: response,
				}

				_, err = b.SendPhoto(ctx, params)

				if err != nil {
					log.Printf("[Q3BOT] [ERROR] failed to send photo message: %s", err)
				}
			}
		}
	}()

}

func getLastMatch() (api.Match, error) {
	data, err := makeRequest(fmt.Sprintf("%s/api/ffa/matches?perpage=1", config.ServerURL), "GET", nil)
	if err != nil {
		return api.Match{}, err
	}

	match := []api.Match{}

	if err := json.Unmarshal(data, &match); err != nil {
		return api.Match{}, err
	}

	if len(match) == 0 {
		return api.Match{}, fmt.Errorf("no matches found")
	}

	return match[0], nil
}

func createMatchInfo(match api.Match) string {
	sort.Slice(match.Players, func(i, j int) bool {
		return match.Players[i].Score > match.Players[j].Score
	})

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

	result := fmt.Sprintf("Map: %s, Duration: %s\n", match.Map, secondsToTime(match.Duration))

	for _, player := range match.Players {
		kdr := float64(player.Kills) / float64(player.Deaths)
		result += fmt.Sprintf("Player: %s, Scores: %d, kills: %d, deaths: %d, suicides: %d, kdr: (%.2f)\n", player.Name, player.Score, player.Kills, player.Deaths, player.Suicides, kdr)
		result += fmt.Sprintf("Damage given: %d, damage taken: %d\n", player.DamageGiven, player.DamageTaken)
		result += fmt.Sprintf("Health taken: %d, armor taken: %d\n", player.HealtTotal, player.ArmorTotal)

		for _, weapon := range player.Weapons {
			result += fmt.Sprintf("With %s made shots: %d, hits: %d, kills: %d\n", weapons[weapon.Name], weapon.Shots, weapon.Hits, weapon.Kills)
		}

		result += "\n\n"
	}

	return result
}
