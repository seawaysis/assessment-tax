# FROM golang:1.21.9-alpine3.19 AS builder
# WORKDIR /app
# COPY . /app

# #RUN go run main.go
# RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

# # Build small images
# FROM alpine:3.19
# WORKDIR /app
# COPY --from=builder /app/main /app/main

# EXPOSE 8080

# CMD ["/app/main"]

# syntax=docker/dockerfile:1

FROM golang:1.21.9-alpine3.19 AS builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY . ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /caltax .

# # Build small images
# FROM alpine3.19
# WORKDIR /app
# COPY --from=builder /caltax .

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose
EXPOSE 8080

# Run
CMD ["/caltax"]

#docker build -t caltax .
#docker run -p 8080:8080 caltax
#https://dev.to/sadeedpv/creating-a-dockerfile-for-your-go-backend-20n5

#docker run --name postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=ktaxes -d -p 5432:5432 postgres
#docker exec -it caltax_db psql -U postgres
#/l //show all database
