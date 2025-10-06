FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o relativisticd ./cmd/relativisticd

FROM gcr.io/distroless/static-debian11

LABEL org.opencontainers.image.title="Relativistic SDK" \
      org.opencontainers.image.description="Relativistic Blockchain Consensus SDK" \
      org.opencontainers.image.vendor="Your Company" \
      org.opencontainers.image.licenses="Apache-2.0"

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /app/relativisticd /app/relativisticd
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/relativisticd", "healthcheck"]

ENTRYPOINT ["/app/relativisticd"]