
FROM golang:1.21-alpine AS build

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/server/main.go


FROM alpine:latest
WORKDIR /app

COPY --from=build /app/app .
ENV PORT=8080
EXPOSE 8080

CMD ["./app"]
