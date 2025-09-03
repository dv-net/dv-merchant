FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./.bin/github.com/dv-net/dv-merchant ./cmd/app/

FROM alpine:latest

COPY --from=build /app/.bin/github.com/dv-net/dv-merchant /app/

ADD config.yml /app/

EXPOSE 8080

WORKDIR /app/

CMD ["./github.com/dv-net/dv-merchant", "start"]
