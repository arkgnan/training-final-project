# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'main.isReleaseBuild=yes'" -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

ENV TZ=Asia/Jakarta
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/assets ./assets

EXPOSE 8080

CMD ["./main"]
