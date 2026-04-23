FROM golang:1.24.1 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/tuubaa-bot ./main.go


FROM debian:stable-slim

WORKDIR /app

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/tuubaa-bot /app/tuubaa-bot
COPY --from=builder /src/assets /app/assets

ENV MONGO_DB=tuubaa

CMD ["/app/tuubaa-bot"]

