# ---------- Build ----------
FROM golang:1.25-alpine AS builder

RUN apk update && apk add --no-cache --no-scripts git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-amd64} \
    go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /app \
      ./cmd/app

# ---------- Runtime ----------
FROM alpine:latest

RUN apk update && \
    apk add --no-cache --no-scripts \
        ca-certificates \
        tzdata

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Create directories for session and repo storage
RUN mkdir -p /app/session /app/repo && chown -R app:app /app

COPY --from=builder /app /app/postpal

USER app

CMD ["/app/postpal"]
