FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod ./
# Since there are no external dependencies, we don't need go mod download
COPY . .
RUN go build -o prymis cmd/prymis/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/prymis .
COPY --from=builder /app/engine ./engine
# Since prymis cmd generates output.png, we might want to keep the structure
CMD ["./prymis"]
