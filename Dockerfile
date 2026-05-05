# ---- Stage 1: Build ----
FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/issue2md ./cmd/issue2md

# ---- Stage 2: Final ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 1000 appuser

COPY --from=builder /bin/issue2md /usr/local/bin/issue2md

USER appuser

ENTRYPOINT ["issue2md"]
