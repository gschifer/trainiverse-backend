FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o bin/app ./cmd

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache tzdata

COPY --from=builder /app/bin/app .

COPY --from=builder /app/internal/db/migrations /app/internal/db/migrations

ENV TZ=America/Sao_Paulo

EXPOSE 8081

CMD ["./app"]
