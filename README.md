# LibBot
Go telegram bot to fetch and convert ebook

# Run

You need to set the env `BOT_TOKEN` to your bot token given by botfather

You will need to install `Calibre` (or at least have "ebook-convert")

```
GO111MODULE=off go run .
```

# Build Docker image

## Build for linux

```
GOOS=linux GO111MODULE=off go build .
docker build -t geobeau/libbot:latest .
```
## Build for linux - arm

```
GOOS=linux GO111MODULE=off GOARCH=arm go build .
docker build -t geobeau/libbot:latest .
```

## Push
```
docker push geobeau/libbot:latest
```