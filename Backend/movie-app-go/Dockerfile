FROM golang:1.25.5-alpine3.23 AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY . .
ENV CGO_ENABLED=0
RUN go build -o movie-app-go main.go

FROM alpine:3.23
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/movie-app-go /app/movie-app-go
EXPOSE 5000
CMD ["/app/movie-app-go"]
