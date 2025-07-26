FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o telegram-downloader-bot main.go

CMD ["./telegram-downloader-bot"] 
