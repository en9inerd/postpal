# ---------- Build ----------
FROM golang:1.25-alpine AS builder

# Install packages with --no-scripts to avoid trigger errors in QEMU emulation
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

# Install packages
# Use --no-scripts to skip triggers that fail in QEMU emulation for multi-arch builds
RUN apk update && \
    apk add --no-cache --no-scripts \
        ca-certificates \
        tzdata

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app /app/app

USER app

EXPOSE 8000

HEALTHCHECK CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

CMD ["/app/app"]
