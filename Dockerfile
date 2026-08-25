FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o rate-limiter main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/rate-limiter .
EXPOSE 8080
CMD ["./rate-limiter"]
