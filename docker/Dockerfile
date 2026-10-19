FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOPROXY=https://proxy.golang.org,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build \
    -trimpath \
    -ldflags="-s -w" \
    -o application .

RUN apk add --no-cache ca-certificates tzdata


FROM scratch AS final

WORKDIR /app

COPY --from=builder /app/application .
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/application"]