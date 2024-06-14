FROM golang:1.22

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all dependencies
RUN go mod download

# Build the Go app
RUN go build -o bin/kwikquiz cmd/kwikquiz/main.go

# Command to run the executable
CMD ["PROD=1 ./bin/kwikquiz"]
