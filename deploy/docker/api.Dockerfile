FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum* ./
RUN go mod download
COPY apps/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/sentinelops-api ./cmd/api

FROM alpine:3.20
RUN addgroup -S sentinelops && adduser -S sentinelops -G sentinelops
WORKDIR /app
COPY --from=builder /out/sentinelops-api /app/sentinelops-api
USER sentinelops
EXPOSE 8080
ENTRYPOINT ["/app/sentinelops-api"]
