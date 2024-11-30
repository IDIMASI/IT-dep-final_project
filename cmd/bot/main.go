package main

import (
	"IT-dep-final_project/internal/bot"
	"log"
	"os"
	"time"
)

func main() {
	time.Local = time.FixedZone("UTC+3", 3*60*60)
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
