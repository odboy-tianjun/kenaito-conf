# Build stage
FROM golang:1.25-alpine AS build

# 设置国内 Go 代理
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /src

# Copy client library
COPY kenaito-conf-go-client/ ./kenaito-conf-go-client/

# Copy demo project
COPY kenaito-conf-go-client-demo/ ./kenaito-conf-go-client-demo/

# Build
WORKDIR /src/kenaito-conf-go-client-demo
RUN go mod tidy && go build -o /app/demo .

# Runtime stage
FROM alpine:3.20

WORKDIR /app
COPY --from=build /app/demo .

ENTRYPOINT ["./demo"]
