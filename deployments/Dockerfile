# syntax=docker/dockerfile:1.7

# ---------------- Stage 1: builder ----------------
FROM golang:1.26-alpine3.21 AS builder

WORKDIR /src

# Cache deps separately
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 make build

# ---------------- Stage 2: runtime (distroless) ----------------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /src/bin/rest /app/rest
COPY --from=builder /src/bin/grpc /app/grpc
COPY --from=builder /src/configs /app/configs

EXPOSE 9090 9091

USER nonroot:nonroot

# Default to REST; override with ["/app/grpc"] for gRPC
ENTRYPOINT ["/app/rest"]
