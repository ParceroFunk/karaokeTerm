# 1. Build the binary
FROM golang:1.26-alpine AS build
WORKDIR /usr/src/app

# Copy Dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the app and compile it
COPY . .
RUN go build -v -o /usr/local/bin/karaoketerm ./main.go

CMD ["karaoketerm"]
