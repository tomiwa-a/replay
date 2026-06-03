FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w \
    -X github.com/replay/replay/internal/version.Version=${VERSION} \
    -X github.com/replay/replay/internal/version.Commit=${COMMIT} \
    -X github.com/replay/replay/internal/version.Date=${DATE}" \
    -o /replay .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /replay /app/replay

RUN addgroup -S replay && adduser -S replay -G replay
USER replay

ENTRYPOINT ["/app/replay"]
