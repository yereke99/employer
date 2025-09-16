
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
        

COPY go.mod go.sum ./
RUN go mod download
        

COPY . .
        

        

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/employer ./cmd/main.go
        

FROM alpine:latest
        
RUN apk --no-cache add ca-certificates tzdata curl \
        && addgroup -S app && adduser -S app -G app
WORKDIR /app
        

COPY --from=builder /app/employer /app/employer

        
COPY --from=builder /app/static /app/static
ENV PORT=8080 \
            STATIC_DIR=/app/static \
            TZ=Asia/Almaty
        
EXPOSE 8080
        
USER app:app
        
HEALTHCHECK --interval=15s --timeout=3s --retries=10 \
CMD curl -fsS http://127.0.0.1:${PORT}/health || exit 1
        
CMD ["/app/employer"]
    