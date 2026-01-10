# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# Set environment variables
#ARG BUILD_ENV="development"
ARG SERVICE_PATH

# Set the working directory
WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .
# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -trimpath -o orchestrator ./${SERVICE_PATH}

# Stage 2: Final stage
FROM alpine:3.19

# Set the working directory
WORKDIR /app

# Copy the binary from the build stage
COPY --from=builder /app/orchestrator .

# Copy migrations
COPY db/migrations /app/db/migrations

# Copy private directory only if in development
#RUN if [ "$BUILD_ENV" = "development" ]; then \
#        cp -r private /app/private; \
#    fi

# Set the timezone and install CA certificates
#RUN apk add --no-cache git ca-certificates tzdata

# Set the timezone to UTC
#ENV TZ=UTC

# Set the entrypoint command
CMD ["/app/orchestrator"]