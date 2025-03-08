package main

import (
	"IT-dep-final_project/internal/bot"
	"log"
	"os"
)

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	b, err := bot.NewBot(botToken)
	if err != nil {
		log.Fatal(err)
	}

	b.Start()
}
