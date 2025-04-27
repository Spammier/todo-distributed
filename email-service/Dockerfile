# --- Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go module files
COPY go.mod ./
# COPY go.sum ./ # Uncomment if you have go.sum

# Set Go Proxy for faster downloads in certain regions
ENV GOPROXY=https://goproxy.cn,direct

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for static linking with Alpine
# -o /app/email-service specifies the output binary path
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/email-service .

# --- Final Stage ---
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/email-service .

# (Optional) If your service needs timezone data or certificates
# RUN apk --no-cache add tzdata ca-certificates

# Expose port (if the service listens on a port, though this one likely won't)
# EXPOSE 8081

# Command to run the executable
CMD ["./email-service"] 