FROM golang:1.23-alpine AS builder

WORKDIR /workspace

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY app ./

RUN CGO_ENABLED=0 GOOS=linux go build -o server -ldflags="-w -s" ./cmd/api

FROM gcr.io/distroless/static-debian11

WORKDIR /app

COPY --from=builder /workspace/server ./app
COPY app/internal/db/migrations ./db/migrations
COPY app/.env.example .
COPY api/specs ./api/specs

EXPOSE 8080

USER nonroot:nonroot

CMD ["./app"]
