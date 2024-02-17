package q3bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func New() *bot.Bot {
	options := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	if config.Verbose {
		options = append(options, bot.WithDebug())
	}

	b, err := bot.New(config.TelegramToken, options...)
	if err != nil {
		log.Fatalf("[Q3BOT] [ERROR] fail to create bot: %s", err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/chatid", bot.MatchTypeExact, chatIDHandler)

	return b
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.Text == "" {
		if config.Verbose {
			log.Println("[Q3BOT] [DEBUG] empty message")
		}
		return
	}
	if !strings.HasPrefix(update.Message.Text, "q3bot") {
		return
	}

	answer, err := openAIResponse(ctx, update.Message.Text)
	if err != nil {
		log.Printf("[Q3BOT] [ERROR] failed to get answer from OpenAI: %s", err)
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:                update.Message.Chat.ID,
		Text:                  answer,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	})

	if err != nil {
		log.Printf("[Q3BOT] [ERROR] failed to send message: %s", err)
	}
}

func chatIDHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:                update.Message.Chat.ID,
		Text:                  fmt.Sprintf("Chat ID: %d", update.Message.Chat.ID),
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	})

	if err != nil {
		log.Printf("[Q3BOT] [ERROR] failed to send message: %s", err)
	}
}
