package q3bot

import (
	"log"

	"github.com/jessevdk/go-flags"
)

var config struct {
	TelegramToken string `long:"telegram" description:"Telegram bot token" required:"true"`
	OpenAIToken   string `long:"openai" description:"OpenAI API key" required:"true"`
	Verbose       bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
}

func init() {
	if _, err := flags.Parse(&config); err != nil {
		log.Fatalf("[Q3BOT] [ERROR] init q3bot config failed: %s", err)
	}
}
