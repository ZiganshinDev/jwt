FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . ./
RUN go build -o ./bin/app  cmd/auth/main.go

FROM alpine

COPY --from=builder /app/bin/app /
COPY --from=builder /app/config.env /
COPY --from=builder /app/config /config/

EXPOSE 8080

CMD ["/app"]