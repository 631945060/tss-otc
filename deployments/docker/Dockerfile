FROM golang:1.27-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o wallet-api ./apps/tss-wallet-service/cmd
FROM alpine:3.21
WORKDIR /
COPY --from=build /app/wallet-api /wallet-api
COPY --from=build /app/web/admin-console /web/admin-console
EXPOSE 8080
ENTRYPOINT ["/wallet-api"]
