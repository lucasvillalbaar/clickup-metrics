# Start from the official Go image
FROM golang:alpine AS build

# Set the working directory
WORKDIR /build

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go dependencies
RUN go mod download

# Copy the source code
COPY . .

# Move to /dist directory as the place for resulting binary folder
WORKDIR cmd/metrics

# Build the Go application
RUN go build -buildvcs=false -o main

# Start a new stage
FROM alpine:latest

# Install the necessary system packages
RUN apk --no-cache add ca-certificates

# Set the working directory in the container
WORKDIR /app

# Copy the built Go binary from the previous stage
COPY --from=build /build/cmd/metrics/main .

# Expose the port that the microservice listens on
EXPOSE 8080

# Run the Go application
CMD ["/app/main"]