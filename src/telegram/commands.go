package telegram

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	startText = `🎬 Welcome to Video Downloader Bot!

I can help you download videos from direct URLs.

Just send me a video URL and I'll download it for you!

Use /help for more information.`

	unknownCommandText = `Unknown command. Use /help to see available commands.`
)

func getHelpText() string {
	isLocalServer := os.Getenv("TELEGRAM_API_URL") != ""

	var apiInfo, sizeLimit string
	if isLocalServer {
		apiInfo = "🔧 API: TDLib Local Server (Enhanced)"
		sizeLimit = "📏 Upload limit: 2000MB"
	} else {
		apiInfo = "🔧 API: Official Telegram API"
		sizeLimit = "📏 Upload limit: 50MB"
	}

	return fmt.Sprintf(`🎬 Video Downloader Bot Commands:

/start - Start the bot
/help - Show this help message
/info <url> - Get video information without downloading
/status - Show current API status and limits

📝 Usage:
• Send a direct video URL to download
• Supported formats: MP4, WebM, AVI, MOV, WMV, FLV, MKV
• Files are saved to the downloads directory

%s
%s

⚠️ Note: Only direct video URLs are supported. YouTube and other streaming platforms may not work.`, apiInfo, sizeLimit)
}

func (t *TelegramBot) handleCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	parts := strings.Fields(message.Text)
	command := parts[0]

	switch command {
	case "/start":
		msg := tgbotapi.NewMessage(chatID, startText)
		if _, err := t.Bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}

	case "/help":
		msg := tgbotapi.NewMessage(chatID, getHelpText())
		if _, err := t.Bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}

	case "/status":
		statusText := t.getStatusText()
		msg := tgbotapi.NewMessage(chatID, statusText)
		if _, err := t.Bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}

	case "/info":
		if len(parts) < 2 {
			msg := tgbotapi.NewMessage(chatID, "Please provide a URL. Usage: /info <url>")
			if _, err := t.Bot.Send(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}
			return
		}

		url := parts[1]
		t.handleVideoInfo(chatID, url)

	default:
		msg := tgbotapi.NewMessage(chatID, unknownCommandText)
		if _, err := t.Bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
