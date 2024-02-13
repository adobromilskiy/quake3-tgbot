package q3bot

import (
	"os"
	"time"

	"github.com/jessevdk/go-flags"
)

var config struct {
	TelegramToken string        `long:"telegram" description:"Telegram bot token" required:"true"`
	OpenAIToken   string        `long:"openai" description:"OpenAI API key" required:"true"`
	ServerURL     string        `long:"server" description:"Quake 3 server URL" required:"true"`
	ChatID        int64         `long:"chat" description:"Telegram chat ID" required:"true"`
	Interval      time.Duration `long:"interval" description:"Interval in seconds to check for new matches" default:"60s"`
	Verbose       bool          `short:"v" long:"verbose" description:"Show verbose debug information"`
}

func init() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}

	if _, err := flags.Parse(&config); err != nil {
		if os.Args[1] == "--help" || os.Args[1] == "-h" {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
