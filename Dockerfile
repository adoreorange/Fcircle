FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN apk add --no-cache git ca-certificates
RUN GOPROXY=https://goproxy.cn,direct go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /fcircle ./cmd/fetch

FROM alpine:3.17

WORKDIR /app

# 创建必要的目录并设置权限
RUN mkdir -p /app/output && \
    chmod 755 /app/output

COPY --from=builder /fcircle /app/
COPY ./config/ /app/config/
COPY ./start.sh /app/start.sh

RUN chmod +x /app/start.sh && \
    chmod +x /app/fcircle

CMD ["/app/start.sh"]
