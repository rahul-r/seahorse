# The build stage
FROM golang:1.22 as builder
WORKDIR /app
COPY . .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o seahorse /app/main.go /app/server.go

# The run stage
FROM scratch
WORKDIR /
COPY --from=builder /app/seahorse .
EXPOSE 9843
CMD ["/seahorse"]
