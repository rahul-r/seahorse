# The build stage
FROM golang:1.22 as builder
WORKDIR /app
COPY . .

RUN curl -L -o /usr/bin/tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.4/tailwindcss-linux-x64 && \
    chmod +x /usr/bin/tailwindcss && \
    tailwindcss -i ./style/tailwind.css -o ./public/styles.css

RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o seahorse /app/main.go /app/server.go

# The run stage
FROM scratch
WORKDIR /
COPY --from=builder /app/seahorse .
EXPOSE 9843
CMD ["/seahorse"]
