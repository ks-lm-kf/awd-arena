FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /awd-arena ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /awd-migrator ./cmd/migrator

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata docker-cli

COPY --from=builder /awd-arena /usr/local/bin/awd-arena
COPY --from=builder /awd-migrator /usr/local/bin/awd-migrator
COPY configs/ /etc/awd-arena/
COPY migrations/ /migrations/

EXPOSE 8080 8443
ENTRYPOINT ["awd-arena"]
CMD ["server", "--config", "/etc/awd-arena/config.yaml"]
