# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /out/api ./services/api-gateway/cmd/api

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
