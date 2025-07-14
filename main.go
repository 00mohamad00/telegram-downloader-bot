package main

import (
	"log"
	"os"
	"time"

	"github.com/00mohamad00/telegram-downloader-bot/downloader"
	"github.com/00mohamad00/telegram-downloader-bot/telegram"
)

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	videoDownloader := downloader.NewVideoDownloader("./downloads", 30*time.Minute)
	bot := telegram.NewTelegramBotOrPanic(botToken, true, videoDownloader)

	bot.Start()
}
