# TDLib Telegram Bot API Integration

This guide explains how to set up and use the TDLib Telegram Bot API server with your video downloader bot for enhanced features and capabilities.

## 🌟 Benefits of Using TDLib

When using the TDLib local server instead of the official Telegram API, you get:

- **📏 Larger file uploads**: Up to 2000MB (instead of 50MB)
- **📥 Unlimited downloads**: No size restrictions on downloads  
- **⚡ Faster uploads**: TDLib server handles local files more efficiently than official API
- **🔧 Better control**: Custom webhook configurations and enhanced features
- **🚀 Enhanced performance**: Direct local communication with reduced latency

## 🛠️ Prerequisites

1. **Telegram API Credentials**: You need to obtain `api_id` and `api_hash` from [https://my.telegram.org/apps](https://my.telegram.org/apps)
2. **Docker & Docker Compose**: For easy deployment
3. **Bot Token**: From [@BotFather](https://t.me/BotFather) on Telegram

## ⚙️ Setup Instructions

### Step 1: Get Telegram API Credentials

1. Go to [https://my.telegram.org/apps](https://my.telegram.org/apps)
2. Log in with your Telegram account
3. Create a new application
4. Note down your `api_id` and `api_hash`

### Step 2: Prepare Environment Variables

Copy the example environment file:
```bash
cp env.example .env
```

Edit `.env` and fill in your credentials:
```bash
# Your bot token from @BotFather
TELEGRAM_BOT_TOKEN=your_bot_token_here

# From https://my.telegram.org/apps
TELEGRAM_API_ID=your_api_id_here
TELEGRAM_API_HASH=your_api_hash_here

# This will be set automatically by docker-compose
TELEGRAM_API_URL=
```

### Step 3: Move Your Bot to Local Server

**Important**: Before starting the local server, you must log out your bot from the official API:

1. Make a request to log out your bot:
```bash
curl -X POST "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/logOut"
```

2. Or use this simple command:
```bash
curl -X POST "https://api.telegram.org/bot$(grep TELEGRAM_BOT_TOKEN .env | cut -d'=' -f2)/logOut"
```

### Step 4: Start the Services

Start both the TDLib server and your bot:
```bash
docker-compose up -d
```

This will:
- Start the TDLib Telegram Bot API server on port 8081
- Start your bot configured to use the local server
- Set up proper networking between services

### Step 5: Verify Setup

Check if everything is running:
```bash
docker-compose ps
```

Check logs:
```bash
# TDLib server logs
docker-compose logs telegram-bot-api

# Bot logs
docker-compose logs downloader-bot
```

## 📱 Usage

Once set up, your bot will automatically use the enhanced TDLib features:

1. **Send `/status`** to see current API configuration
2. **Send `/help`** to see updated limits and features
3. **Upload large videos** up to 2000MB
4. **Enjoy faster uploads** using local file paths

### Available Commands

- `/start` - Start the bot
- `/help` - Show help with current limits
- `/status` - Show detailed API status and capabilities
- `/info <url>` - Get video information without downloading

## 🔧 Configuration Options

### Using Official API (Fallback)

To switch back to official API, simply set:
```bash
TELEGRAM_API_URL=
```

Or stop the docker-compose and run with:
```bash
docker run --env-file .env your-bot-image
```

### Custom TDLib Options

You can modify the TDLib server configuration in `docker-compose.yml`:

```yaml
command: >
  telegram-bot-api
  --api-id=${TELEGRAM_API_ID}
  --api-hash=${TELEGRAM_API_HASH}
  --local
  --http-port=8081
  --dir=/var/lib/telegram-bot-api
  --max-webhook-connections=100000  # Custom option
  --verbosity=1                     # Add for debugging
```

## 🐛 Troubleshooting

### Bot Not Receiving Updates

1. Make sure you logged out from official API first
2. Check TDLib server logs: `docker-compose logs telegram-bot-api`
3. Verify environment variables are set correctly

### File Upload Issues

1. Check available disk space
2. Verify file permissions in downloads directory
3. Check bot logs for specific error messages

### Connection Issues

1. Ensure Docker containers can communicate
2. Check if port 8081 is available
3. Verify network configuration in docker-compose.yml

### Performance Issues

1. Increase TDLib server resources in docker-compose.yml:
```yaml
deploy:
  resources:
    limits:
      memory: 1G
    reservations:
      memory: 512M
```

## 📊 Monitoring

### Check API Status
```bash
curl http://localhost:8081/bot<YOUR_BOT_TOKEN>/getMe
```

### Monitor Resource Usage
```bash
docker stats
```

### View Real-time Logs
```bash
docker-compose logs -f
```

## 🔄 Updating

To update the TDLib server:
```bash
docker-compose pull telegram-bot-api
docker-compose up -d telegram-bot-api
```

To update your bot:
```bash
docker-compose build downloader-bot
docker-compose up -d downloader-bot
```

## 🛡️ Security Considerations

1. **Keep credentials secure**: Never commit `.env` files
2. **Use HTTPS in production**: Set up a reverse proxy (nginx/traefik)
3. **Firewall protection**: Limit access to port 8081
4. **Regular updates**: Keep TDLib server updated

## 📚 Additional Resources

- [TDLib Documentation](https://core.telegram.org/tdlib)
- [Telegram Bot API](https://core.telegram.org/bots/api)
- [TDLib GitHub Repository](https://github.com/tdlib/telegram-bot-api)

## ❓ FAQ

**Q: Can I use both official and local API simultaneously?**
A: No, a bot can only be active on one server at a time.

**Q: Will my bot lose messages during migration?**
A: If done properly (log out first), no messages should be lost.

**Q: Can I migrate back to official API?**
A: Yes, log out from local server and use official API again.

**Q: Does this work with webhooks?**
A: Yes, TDLib supports webhooks with additional configuration options. 
