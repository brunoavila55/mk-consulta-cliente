## build stage
FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/consulta-cliente .

## runtime stage
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/consulta-cliente /app/consulta-cliente

ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/consulta-cliente"]
