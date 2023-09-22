FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . ./
RUN go build -o ./auth-app  cmd/auth/main.go

FROM alpine

COPY --from=builder /app/auth-app /
COPY  config.env /
COPY /config /config

EXPOSE 8080

CMD ["/auth-app"]