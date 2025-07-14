package telegram

import (
	"fmt"
	"log"
	"strings"

	"github.com/00mohamad00/telegram-downloader-bot/downloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	Bot             *tgbotapi.BotAPI
	VideoDownloader *downloader.VideoDownloader
}

func NewTelegramBotOrPanic(
	botToken string,
	debug bool,
	videoDownloader *downloader.VideoDownloader,
) *TelegramBot {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = debug

	return &TelegramBot{
		Bot:             bot,
		VideoDownloader: videoDownloader,
	}
}

func (t *TelegramBot) Start() {
	log.Printf("Starting Telegram bot %s", t.Bot.Self.UserName)

	u := tgbotapi.NewUpdate(1)
	u.Timeout = 60

	updates := t.Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		t.handleMessage(update.Message)
	}
}

func (t *TelegramBot) handleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	messageText := message.Text

	if strings.HasPrefix(messageText, "/") {
		t.handleCommand(message)
		return
	}

	if t.VideoDownloader.IsValidVideoURL(messageText) {
		t.handleVideoDownload(chatID, messageText)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Send me a video URL to download!\n\nUsage:\n- Just send a direct video URL\n- Use /info <url> to get video information\n- Use /help for more commands")
	if _, err := t.Bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (t *TelegramBot) handleVideoDownload(chatID int64, url string) {
	processingMsg := tgbotapi.NewMessage(chatID, "🔄 Processing your request...")
	if _, err := t.Bot.Send(processingMsg); err != nil {
		log.Printf("Error sending processing message: %v", err)
	}

	videoInfo, err := t.VideoDownloader.GetVideoInfo(url)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Error getting video info: %v", err))
		if _, err := t.Bot.Send(errorMsg); err != nil {
			log.Printf("Error sending error message: %v", err)
		}
		return
	}

	infoText := fmt.Sprintf("📹 Downloading video...\n\nFilename: %s\nSize: %s\nType: %s",
		videoInfo.Filename, videoInfo.FormatSize(), videoInfo.ContentType)

	infoMsg := tgbotapi.NewMessage(chatID, infoText)
	if _, err := t.Bot.Send(infoMsg); err != nil {
		log.Printf("Error sending info message: %v", err)
	}

	filePath, err := t.VideoDownloader.DownloadVideo(url)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Error downloading video: %v", err))
		if _, err := t.Bot.Send(errorMsg); err != nil {
			log.Printf("Error sending error message: %v", err)
		}
		return
	}

	successMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Video downloaded successfully!\n\nSaved to: %s", filePath))
	if _, err := t.Bot.Send(successMsg); err != nil {
		log.Printf("Error sending success message: %v", err)
	}

	log.Printf("Video downloaded successfully: %s", filePath)
}

func (t *TelegramBot) handleVideoInfo(chatID int64, url string) {
	videoInfo, err := t.VideoDownloader.GetVideoInfo(url)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Error getting video info: %v", err))
		if _, err := t.Bot.Send(errorMsg); err != nil {
			log.Printf("Error sending error message: %v", err)
		}
		return
	}

	infoText := fmt.Sprintf("📹 Video Information:\n\nURL: %s\nFilename: %s\nSize: %s\nContent Type: %s",
		videoInfo.URL, videoInfo.Filename, videoInfo.FormatSize(), videoInfo.ContentType)

	msg := tgbotapi.NewMessage(chatID, infoText)
	if _, err := t.Bot.Send(msg); err != nil {
		log.Printf("Error sending info message: %v", err)
	}
}
