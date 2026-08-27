# Build stage
FROM golang:1.26-alpine AS build

ENV GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/main.go

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=build /out/server ./server
COPY application.properties application-docker.properties ./

ENV KENAITO_ACTIVE=docker

EXPOSE 8080

ENTRYPOINT ["./server", "-config", "application.properties"]
