package q3bot

import (
	"context"
	"log"

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

	return b
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answer, err := openAIResponse(ctx, update.Message.Text)
	if err != nil {
		log.Printf("[Q3BOT] [ERROR] failed to get answer from OpenAI: %s", err)
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      answer,
		ParseMode: "HTML",
	})

	if err != nil {
		log.Printf("[Q3BOT] [ERROR] failed to send message: %s", err)
	}
}
