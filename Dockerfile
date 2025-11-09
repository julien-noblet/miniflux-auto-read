FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o miniflux-auto-read

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/miniflux-auto-read .
EXPOSE 8080
CMD ["./miniflux-auto-read"]