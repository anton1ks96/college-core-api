FROM golang:1.25-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

RUN adduser -D -g '' appuser

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o college-core-api ./cmd/app

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /build/configs /configs

COPY --from=builder /build/college-core-api /college-core-api

USER appuser

EXPOSE 8081

ENTRYPOINT ["/college-core-api"]