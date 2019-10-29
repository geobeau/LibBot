# LibBot
Go telegram bot to fetch and convert ebook

# Run

You need to set the env `BOT_TOKEN` to your bot token given by botfather

```
go run cmd/libbot.go
```

# Build Docker image

```
GOOS=linux go build cmd/libbot.go
docker build -t geobeau/libbot:latest .
```

```
docker push geobeau/libbot:latest
```