# Stage 1: Build the Go application
FROM golang:1.19-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o main .

# Stage 2: Create the final image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built application from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/index.html .

# Expose the port the application will run on
EXPOSE 8080

# Command to run the application
CMD ["./main"]

