# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

# Allow Go to download newer toolchain if needed
ENV GOTOOLCHAIN=auto

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/server ./cmd/turso-server

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

COPY --from=builder /app/auth-server .

# PocketBase data directory
VOLUME /app/pb_data

EXPOSE 8090

CMD ["./auth-server"]
