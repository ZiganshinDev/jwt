FROM golang:1.20

WORKDIR /auth

COPY . .

RUN go build -o auth.exe ./cmd/auth/main.go

CMD ["./auth.exe"]