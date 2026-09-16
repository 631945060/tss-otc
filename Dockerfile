FROM golang:1.27-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o wallet-api ./cmd/server
FROM alpine:3.21
COPY --from=build /app/wallet-api /wallet-api
EXPOSE 8080
ENTRYPOINT ["/wallet-api"]
