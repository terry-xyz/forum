FROM golang:1.24@sha256:d2d2bc1c84f7e60d7d2438a3836ae7d0c847f4888464e7ec9ba3a1339a1ee804 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o forum .


FROM debian:bookworm-slim@sha256:0104b334637a5f19aa9c983a91b54c89887c0984081f2068983107a6f6c21eeb

WORKDIR /app

COPY --from=builder --chown=10001:10001 /app/forum .
COPY --from=builder --chown=10001:10001 /app/database/schema.sql database/schema.sql

RUN chmod 0755 /app/forum && chown -R 10001:10001 /app

USER 10001:10001

EXPOSE 8080

CMD ["./forum"]
