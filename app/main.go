package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/adobromilskiy/quake3-tgbot/app/q3bot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b := q3bot.New()

	b.Start(ctx)
}
