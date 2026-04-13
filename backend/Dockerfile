# syntax=docker/dockerfile:1.6
FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags="-s -w" -o /out/api ./cmd/api \
 && go build -ldflags="-s -w" -o /out/importer ./cmd/importer

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /out/importer /app/importer
COPY --from=builder /src/internal/db/migrations /app/internal/db/migrations
EXPOSE 8080
ENTRYPOINT ["/app/api"]
