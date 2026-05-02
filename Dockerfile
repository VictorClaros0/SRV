# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY . .

# CGO_ENABLED=0 produces a fully static binary that runs on any Linux distro
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/api ./cmd/main.go

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
# Debian slim is used (not Alpine) because tesseract-ocr and poppler-utils
# are more reliably packaged there and the static Go binary runs on any Linux.
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        tzdata \
        tesseract-ocr \
        tesseract-ocr-spa \
        tesseract-ocr-eng \
        poppler-utils \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/internal/data/*.csv ./internal/data/

EXPOSE 8080

ENTRYPOINT ["/app/api"]
