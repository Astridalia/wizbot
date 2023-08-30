FROM golang:1.20.4

# Install any additional dependencies if needed

WORKDIR /go/src/app

# Copy your Go application files to the container
COPY . .

# Add the missing go.sum entry
RUN go mod download github.com/onsi/ginkgo

# Build the Go application
RUN go build

# Run the Go application with environment variables
CMD ["./bot"]

# EXPOSE 6379
EXPOSE 27017
