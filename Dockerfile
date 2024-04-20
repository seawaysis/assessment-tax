FROM golang:1.21.9-alpine3.19 AS builder
WORKDIR /app
COPY . /app

#RUN go run main.go
RUN go build -o main main.go

# Build small images
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["/app/main"]