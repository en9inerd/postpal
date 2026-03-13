# ---------- Build ----------
# Pin builder to the build machine's native platform so Go cross-compiles
# natively instead of running under QEMU emulation.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
      -gcflags="all=-l -B" \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /app \
      ./cmd/postpal

# ---------- Runtime ----------
FROM alpine:3.23

RUN apk update && \
    apk add --no-cache \
        ca-certificates \
        tzdata

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Create directories for session and repo storage
RUN mkdir -p /app/session /app/repo && chown -R app:app /app

COPY --from=builder /app /app/postpal

USER app

EXPOSE 8080

HEALTHCHECK CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

CMD ["/app/postpal"]
