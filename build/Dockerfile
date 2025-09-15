FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest && swag init -g cmd/main.go -o ./docs

RUN CGO_ENABLED=0 GOOS=linux go build -o employer ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/employer .


COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./employer"]
