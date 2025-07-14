package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Get bot token from environment variable
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	// Create bot instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// Set debug mode (optional)
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Configure updates
	u := tgbotapi.NewUpdate(1)
	u.Timeout = 60

	// Get updates channel
	updates := bot.GetUpdatesChan(u)

	// Handle incoming messages
	for update := range updates {
		// Check if we have a message
		if update.Message == nil {
			continue
		}

		// Log received message
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// Create reply message
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello!")

		// Send the reply
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
