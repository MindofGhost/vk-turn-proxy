FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o vk-turn-proxy ./server

FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY docker-entrypoint.sh .
COPY docker-healthcheck.sh .
COPY --from=builder /build/vk-turn-proxy .
RUN sed -i 's/\r$//' docker-entrypoint.sh docker-healthcheck.sh && \
    chmod +x docker-entrypoint.sh docker-healthcheck.sh

EXPOSE 56000/tcp
EXPOSE 56000/udp

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
    CMD ["./docker-healthcheck.sh"]

ENTRYPOINT ["./docker-entrypoint.sh"]
