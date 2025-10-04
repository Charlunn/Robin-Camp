FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app -ldflags="-w -s" ./...

FROM gcr.io/distroless/static-debian11

WORKDIR /app

COPY --from=builder /app/app .
COPY db/migrations ./db/migrations
COPY .env.example .

EXPOSE 8080

USER nonroot:nonroot

CMD ["./app"]
