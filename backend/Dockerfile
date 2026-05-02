# Copyright: proyecto conteo de votos
# Build estático del API en Go
FROM golang:1.23-alpine AS builder
WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /votos-api .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /votos-api /app/votos-api
COPY data /app/data
RUN mkdir -p /app/data/logs

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/votos-api"]
