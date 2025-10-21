# syntax=docker/dockerfile:1

# 1) ดึง Go 1.24 มาเป็น builder
FROM golang:1.24 AS builder
WORKDIR /src

# คัดลอกไฟล์ go.mod/go.sum จาก "รากโปรเจกต์" เพื่อดึง dependencies ก่อน (cache ดี)
COPY go.mod go.sum ./
RUN go mod download

# คัดลอกซอร์สทั้งหมด (proto/pkg/services)
COPY . .

# สร้างไบนารีของ auth-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /out/auth ./services/auth-service/cmd/auth

# 2) Runtime บางเบา
FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/auth /auth
EXPOSE 50051
ENTRYPOINT ["/auth"]
