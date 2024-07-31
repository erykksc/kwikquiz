FROM golang:1.22.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies.
# Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

# Build the Go app
RUN go build -o bin/kwikquiz kwikquiz.go

CMD ["./bin/kwikquiz"]

FROM golang:1.22.5

WORKDIR /app

COPY --from=builder /app/bin/kwikquiz /app/kwikquiz

EXPOSE 3000

CMD ["/app/kwikquiz", "-prod", "-port", "3000"]
