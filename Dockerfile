# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Dependencias de compilación
RUN apk add --no-cache git

WORKDIR /app

# Copiar todo el código fuente
COPY . .

# Descargar dependencias y compilar binario estático
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/api ./cmd/main.go

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.20

# Certificados para TLS y timezone
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copiar solo el binario compilado
COPY --from=builder /app/api .

EXPOSE 8080

ENTRYPOINT ["/app/api"]
