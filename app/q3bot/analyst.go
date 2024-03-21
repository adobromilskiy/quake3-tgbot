package q3bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
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
					log.Printf("[ANALYST] [ERROR] failed to get last match: %s", err)
					continue
				}

				if lastMatchID == primitive.NilObjectID && !config.Verbose {
					lastMatchID = match.ID
					continue
				}

				if lastMatchID == match.ID {
					continue
				}

				lastMatchID = match.ID

				prompt := "Представь что ты комментируешь итоги игры в Quake 3. Победитель тот, у кого больше всех очков. Постарайся упомянуть название карты. Твой комментарий должен быть написан в саркастическом тоне на русском языке. ПОЖАЛУЙСТА, ТВОЕ СООБЩЕНИЕ ДОЛЖНО СОДЕРЖАТЬ МАКСИМУМ 120 СЛОВ!!!"

				response, err := analyzeMatchInfo(ctx, createMatchInfo(match), prompt)
				if err != nil {
					log.Printf("[ANALYST] [ERROR] failed to analyze match info: %s", err)
					continue
				}

				if config.Verbose {
					log.Printf("[ANALYST] [DEBUG] comment length %d", len(response))
				}

				if len(response) > 1024 {
					_, err = b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID:                config.ChatID,
						Text:                  response,
						ParseMode:             "HTML",
						DisableWebPagePreview: true,
					})

					if err != nil {
						log.Printf("[ANALYST] [ERROR] failed to send message: %s", err)
					}

					continue
				}

				prompt = fmt.Sprintf("Нарисуй изображение про киберспорт, основываясь на следующем тексте:\n\n %s", response)

				image, err := generateImage(ctx, prompt)
				if err != nil {
					log.Printf("[ANALYST] [ERROR] failed to generate image: %s", err)
					_, err = b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID:                config.ChatID,
						Text:                  response,
						ParseMode:             "HTML",
						DisableWebPagePreview: true,
					})

					if err != nil {
						log.Printf("[ANALYST] [ERROR] failed to send message: %s", err)
					}
					continue
				}

				params := &bot.SendPhotoParams{
					ChatID:  config.ChatID,
					Photo:   &models.InputFileString{Data: image},
					Caption: response,
				}

				_, err = b.SendPhoto(ctx, params)

				if err != nil {
					log.Printf("[ANALYST] [ERROR] failed to send photo message: %s", err)
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
		"G":   "Перчатка",
		"MG":  "Пулемет",
		"SG":  "Дробовик",
		"GL":  "Гранаты",
		"RL":  "Ракетница",
		"LG":  "Шкварка",
		"RG":  "Рельса",
		"PG":  "Плазма",
		"BFG": "BFG",
	}

	players := map[string]string{
		"ip":           "Дед",
		"javascripter": "Усы",
	}

	maps := map[string]string{
		"cpm1a":  "Потная",
		"cpm15":  "Сельская",
		"cpm28":  "Считалка",
		"cpm29":  "Рулетка",
		"q3dm17": "Космос",
	}

	mapname, ok := maps[strings.ToLower(match.Map)]
	if !ok {
		mapname = match.Map
	}

	result := fmt.Sprintf("Завершилась игра на карте '%s' с продолжительностю %s.\n", mapname, secondsToTime(match.Duration))

	for _, player := range match.Players {
		kdr := float64(player.Kills) / float64(player.Deaths)
		name, ok := players[strings.ToLower(player.Name)]
		if !ok {
			name = player.Name
		}

		result += fmt.Sprintf("%s набрал очков %d, сделав %d фрагов. Другие игроки набрали на нем %d фрагов. Себе причинил вред %d раз. При этом его КДР составил %.2f\n", name, player.Score, player.Kills, player.Deaths, player.Suicides, kdr)
		result += fmt.Sprintf("Нанес урона: %d, получил урона: %d\n", player.DamageGiven, player.DamageTaken)
		result += fmt.Sprintf("Подобрал аптечек: %d, подобрал брони: %d\n", player.HealtTotal, player.ArmorTotal)

		for _, weapon := range player.Weapons {
			if weapon.Name == "G" {
				result += fmt.Sprintf("Используя %s, попал %d раз и смог сделать %d фрагов.\n", weapons[weapon.Name], weapon.Hits, weapon.Kills)
				continue
			}
			result += fmt.Sprintf("Используя %s, сделал %d выстрелов. При этом попал %d раз и смог сделать %d фрагов.\n", weapons[weapon.Name], weapon.Shots, weapon.Hits, weapon.Kills)
		}

		result += "\n\n"
	}

	return result
}
