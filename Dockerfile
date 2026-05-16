FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build the nucleus CLI binary. The previous Dockerfile built an
# examples/mvc_api server; that example tree was removed in the
# ADR-010 Phase 1 iteration on 2026-05-16. A reference application
# image returns with the v0.9.X reference applications (ADR-010
# Phase 4). For now this Dockerfile produces a working `nucleus` CLI
# image — operators downstream that previously consumed the
# example-server image should pin to a pre-2026-05-16 tag until v0.9.X.
RUN CGO_ENABLED=1 go build -o /app/nucleus ./cmd/nucleus || (apk add --no-cache gcc musl-dev && CGO_ENABLED=1 go build -o /app/nucleus ./cmd/nucleus)

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/nucleus /app/nucleus

ENTRYPOINT ["/app/nucleus"]
CMD ["--help"]
