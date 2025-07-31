package telegram

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/00mohamad00/telegram-downloader-bot/src/downloader"
	"github.com/00mohamad00/telegram-downloader-bot/src/pkg/videoinfo"
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
	// Check if custom API URL is set (for local TDLib server)
	apiURL := os.Getenv("TELEGRAM_API_URL")

	var bot *tgbotapi.BotAPI
	var err error

	if apiURL != "" {
		// Use custom API URL (TDLib local server)
		bot, err = tgbotapi.NewBotAPIWithAPIEndpoint(botToken, apiURL+"/bot%s/%s")
		log.Printf("Using custom Telegram API endpoint: %s", apiURL)
	} else {
		// Use official Telegram API
		bot, err = tgbotapi.NewBotAPI(botToken)
		log.Printf("Using official Telegram API endpoint")
	}

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

	// Upload the video to Telegram
	t.uploadVideoToTelegram(chatID, filePath, videoInfo)
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

	// Add metadata if available
	if videoInfo.Duration > 0 {
		infoText += fmt.Sprintf("\nDuration: %d seconds", videoInfo.Duration)
	}
	if videoInfo.Width > 0 && videoInfo.Height > 0 {
		infoText += fmt.Sprintf("\nDimensions: %dx%d", videoInfo.Width, videoInfo.Height)
	}

	msg := tgbotapi.NewMessage(chatID, infoText)
	if _, err := t.Bot.Send(msg); err != nil {
		log.Printf("Error sending info message: %v", err)
	}
}

func (t *TelegramBot) uploadVideoToTelegram(chatID int64, filePath string, videoInfo *videoinfo.VideoInfo) {
	// Check if we're using local TDLib server for enhanced limits
	isLocalServer := os.Getenv("TELEGRAM_API_URL") != ""

	var maxFileSize int64
	var fileSizeLimit string

	if isLocalServer {
		maxFileSize = 2000 * 1024 * 1024 // 2000MB for TDLib local server
		fileSizeLimit = "2000MB"
	} else {
		maxFileSize = 50 * 1024 * 1024 // 50MB for official API
		fileSizeLimit = "50MB"
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		errorMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Error accessing downloaded file: %v", err))
		if _, err := t.Bot.Send(errorMsg); err != nil {
			log.Printf("Error sending file access error message: %v", err)
		}
		return
	}

	fileSize := fileInfo.Size()
	if fileSize > maxFileSize {
		log.Printf("File too large for Telegram upload: %d bytes (max: %d bytes)", fileSize, maxFileSize)

		errorMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ File too large for Telegram upload!\n\n📊 File size: %s\n📏 Current limit: %s\n\n📁 Video saved locally to: %s\n\n💡 Tip: Use TDLib local server for 2000MB limit!",
			videoInfo.FormatSize(), fileSizeLimit, filePath))
		if _, err := t.Bot.Send(errorMsg); err != nil {
			log.Printf("Error sending file size error message: %v", err)
		}
		return
	}

	uploadingMsg := tgbotapi.NewMessage(chatID, "📤 Uploading video to Telegram...")
	if _, err := t.Bot.Send(uploadingMsg); err != nil {
		log.Printf("Error sending uploading message: %v", err)
	}

	// Both TDLib and official API can use file paths
	// TDLib server handles local files more efficiently internally
	video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(filePath))

	// Set video metadata for better previews in Telegram
	if videoInfo.Duration > 0 {
		video.Duration = videoInfo.Duration
		log.Printf("Setting video duration: %d seconds", videoInfo.Duration)
	}
	if videoInfo.Width > 0 && videoInfo.Height > 0 {
		log.Printf("Video dimensions: %dx%d (metadata extracted)", videoInfo.Width, videoInfo.Height)
	}

	if isLocalServer {
		log.Printf("Using TDLib local server for file: %s", filePath)
	} else {
		log.Printf("Using official API for file: %s", filePath)
	}

	caption := fmt.Sprintf("✅ Video uploaded successfully!\n\n📁 Filename: %s\n💾 Size: %s\n🔧 API: %s",
		videoInfo.Filename, videoInfo.FormatSize(),
		map[bool]string{true: "TDLib Local", false: "Official"}[isLocalServer])

	// Add metadata to caption if available
	if videoInfo.Duration > 0 {
		caption += fmt.Sprintf("\n⏱️ Duration: %d seconds", videoInfo.Duration)
	}
	if videoInfo.Width > 0 && videoInfo.Height > 0 {
		caption += fmt.Sprintf("\n📐 Resolution: %dx%d", videoInfo.Width, videoInfo.Height)
	}
	video.Caption = caption

	_, err = t.Bot.Send(video)
	if err != nil {
		log.Printf("Error uploading video: %v", err)

		errorMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Error uploading video: %v\n\n📁 Video saved locally to: %s", err, filePath))
		if _, err := t.Bot.Send(errorMsg); err != nil {
			log.Printf("Error sending upload error message: %v", err)
		}
		return
	}

	log.Printf("Video uploaded successfully to Telegram: %s", filePath)

	// Only delete file if upload was successful
	if err := os.Remove(filePath); err != nil {
		log.Printf("Warning: Could not delete local file %s: %v", filePath, err)
	} else {
		log.Printf("Local file deleted: %s", filePath)
	}
}

func (t *TelegramBot) getStatusText() string {
	isLocalServer := os.Getenv("TELEGRAM_API_URL") != ""
	apiURL := os.Getenv("TELEGRAM_API_URL")

	var statusText string
	if isLocalServer {
		statusText = fmt.Sprintf(`🤖 Bot Status Information:

🔧 API Type: TDLib Local Server
🌐 Endpoint: %s
📏 Upload Limit: 2000MB
📥 Download Limit: Unlimited
⚡ Local File Support: ✅ Enabled
🔒 Enhanced Features: ✅ Available

✨ Benefits:
• Much larger file uploads (2000MB vs 50MB)
• Faster uploads using local file paths
• No download size restrictions
• Better performance and control`, apiURL)
	} else {
		statusText = `🤖 Bot Status Information:

🔧 API Type: Official Telegram API
🌐 Endpoint: api.telegram.org
📏 Upload Limit: 50MB
📥 Download Limit: Standard
⚡ Local File Support: ❌ Not Available
🔒 Enhanced Features: ❌ Not Available

💡 Tip: Set up TDLib local server for enhanced features!`
	}

	return statusText
}
