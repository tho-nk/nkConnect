# Use the official golang image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /go/src/app

# Copy the Go modules files into the container
COPY go.mod .
COPY go.sum .

# Download and install Go dependencies
RUN go mod download

# Copy the entire project into the container
COPY . .

# Build the Go application
RUN go build -o nkConnect ./cmd/nkConnect

# Expose the port the application runs on
EXPOSE 9096

# Command to run the executable
CMD ["./nkConnect"]
