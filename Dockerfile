FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies.
# Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o bin/kwikquiz cmd/main.go

CMD ["./bin/kwikquiz"]
