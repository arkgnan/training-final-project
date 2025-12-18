# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install git for go modules that might need it
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init

# Build the application
# add -tags=embed to embed env.
# example : go build -tags=embed -ldflags="-X main.isReleaseBuild=yes" -o mygram-api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'main.isReleaseBuild=yes'" -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create app directory
WORKDIR /app

ENV TZ=Asia/Jakarta

RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy swagger docs
COPY --from=builder /app/docs ./docs

# Copy email templates
COPY --from=builder /app/templates ./templates

# Copy static assets (untuk email templates dan static files lainnya)
COPY --from=builder /app/assets ./assets

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/swagger/index.html || exit 1

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]
