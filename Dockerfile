FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o telegram-downloader-bot main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create downloads directory
RUN mkdir -p downloads

# Copy the binary from builder stage
COPY --from=builder /app/telegram-downloader-bot .

# Make sure the binary is executable
RUN chmod +x telegram-downloader-bot

CMD ["./telegram-downloader-bot"] 
