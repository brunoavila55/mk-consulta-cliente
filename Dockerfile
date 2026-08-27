# syntax=docker/dockerfile:1

FROM golang:1.26.7-alpine3.24 AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck

FROM alpine:3.24.1 AS runtime

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -D -H -u 10001 -G app app

COPY --from=build --chown=app:app /out/api /usr/local/bin/api
COPY --from=build --chown=app:app /out/healthcheck /usr/local/bin/healthcheck

ENV HTTP_PORT=8080

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD ["/usr/local/bin/healthcheck"]

STOPSIGNAL SIGTERM
ENTRYPOINT ["/usr/local/bin/api"]
