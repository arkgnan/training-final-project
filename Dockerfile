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

ARG RAILWAY_GIT_COMMIT_SHA
ARG DEPLOY_MODE

# Build the application
# flags '-w -s' untuk memperkecil ukuran binary
# Logika: Jika APP_ENV=production, tambahkan flag -tags=embed
RUN BUILD_TAGS=""; \
    if [ "$DEPLOY_MODE" = "embed" ]; then BUILD_TAGS="-tags=embed"; fi; \
    CGO_ENABLED=0 GOOS=linux go build ${BUILD_TAGS} -ldflags="-s -w -X 'main.isReleaseBuild=yes' -X 'main.CommitHash=${RAILWAY_GIT_COMMIT_SHA}'" -a -installsuffix cgo -o out


# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create app directory
WORKDIR /app

ENV TZ=Asia/Jakarta

RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy the binary from builder stage
COPY --from=builder /app/out .

# Copy swagger docs
COPY --from=builder /app/docs ./docs

# Copy email templates
COPY --from=builder /app/templates ./templates

# Copy static assets (untuk email templates dan static files lainnya)
COPY --from=builder /app/assets ./assets

RUN echo ">>> Cek file hasil copy:" && ls -la /app

# Expose port
ARG APP_PORT
EXPOSE ${PORT}

# Command to run
CMD ["./out"]
