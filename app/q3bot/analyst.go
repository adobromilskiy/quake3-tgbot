package q3bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"

	"github.com/adobromilskiy/quake3-stats/app/api"
	"github.com/go-telegram/bot"
)

var lastMatchID string

type match struct {
	ID string `json:"id"`
}

func Analyze(ctx context.Context, b *bot.Bot) {
	ticker := time.NewTicker(config.Interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				data, err := makeRequest(fmt.Sprintf("%s/api/ffa/matches?perpage=2", config.ServerURL), "GET", nil)
				if err != nil {
					log.Printf("[Q3BOT] [ERROR] failed to get data from server: %s", err)
					continue
				}

				matches := []match{}

				match := []api.Match{}

				if err := json.Unmarshal(data, &matches); err != nil {
					log.Printf("[Q3BOT] [ERROR] failed to unmarshal data: %s", err)
					continue
				}

				if len(matches) == 0 {
					log.Printf("[Q3BOT] [ERROR] no matches found")
					continue
				}

				if lastMatchID == matches[0].ID {
					continue
				}

				if err := json.Unmarshal(data, &match); err != nil {
					log.Printf("[Q3BOT] [ERROR] failed to unmarshal data: %s", err)
					continue
				}

				for m := range match {
					sort.Slice(match[m].Players, func(i, j int) bool {
						return match[m].Players[i].Score > match[m].Players[j].Score
					})
				}

				lastMatchID = matches[0].ID

				prompt := "You are a helpfull assistant. When asked for you name, you must respond with 'Quake3 Bot'. Imagine that you are a game commentator. Please compare matches about Quake 3 game and make summary for each player in RUSSIAN LANGUAGE with sarcastic form. COMMENT MUST BE NO MORE THEN 350 WORDS. WINNER IS THE ONE WHO HAS THE MOST SCORES."
				if rand.Intn(100) > 30 {
					prompt = "You are a helpfull assistant. When asked for you name, you must respond with 'Quake3 Bot'. Imagine that you are a game commentator. Please analyze the match info about Quake 3 game and make summary for each player in RUSSIAN LANGUAGE with sarcastic form. COMMENT MUST BE NO MORE THEN 350 WORDS. WINNER IS THE ONE WHO HAS THE MOST SCORES."

					match = match[0:]
				}

				data, err = json.Marshal(match)
				if err != nil {
					log.Printf("[Q3BOT] [ERROR] failed to marshal data: %s", err)
					continue
				}

				response, err := analyzeMatchInfo(ctx, string(data), prompt)
				if err != nil {
					log.Printf("[Q3BOT] [ERROR] failed to analyze match info: %s", err)
					continue
				}

				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:                config.ChatID,
					Text:                  response,
					ParseMode:             "HTML",
					DisableWebPagePreview: true,
				})

				if err != nil {
					log.Printf("[Q3BOT] [ERROR] failed to send message: %s", err)
				}
			}
		}
	}()

}
