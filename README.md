# LibBot
Go telegram bot to fetch and convert ebook

# Run

You need to set the env `BOT_TOKEN` to your bot token given by botfather

You will need to install `Calibre` (or at least have "ebook-convert")

```
go run .
```

# Build Docker image

```
GOOS=linux go build .
docker build -t geobeau/libbot:latest .
```

```
docker push geobeau/libbot:latest
```