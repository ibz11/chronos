# --- Build stage ---
FROM golang:1.22-alpine AS build

# Install git (needed if you have modules)
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum if you have them, otherwise skip
# COPY go.mod go.sum ./
# RUN go mod download

# Copy the source code
COPY main.go .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o chronos main.go

# --- Run stage ---
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy the compiled binary
COPY --from=build /app/chronos .

# Expose port (make sure it matches your .env PORT)
EXPOSE 8080

# Run as non-root for security
USER nonroot:nonroot

# Run the binary
CMD ["./chronos"]
