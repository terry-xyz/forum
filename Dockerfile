FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o forum .


FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/forum .
COPY --from=builder /app/database/schema.sql database/schema.sql

EXPOSE 8080

CMD ["./forum"]