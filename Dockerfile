FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Binary'yi sensor-monitoring olarak derle
RUN go build -o sensor-monitoring

FROM alpine:latest

WORKDIR /app

# Gerekli kütüphaneler (örn. SSL sertifikaları)
RUN apk add --no-cache ca-certificates

# Build aşamasından binary'yi kopyala
COPY --from=builder /app/sensor-monitoring .

# Statik dosyaları kopyala
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./sensor-monitoring"]
