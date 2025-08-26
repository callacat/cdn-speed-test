# --- Build Stage ---
# Use the same Go version as specified in go.mod for consistency
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy module files first to leverage Docker's build cache.
# This layer only gets rebuilt if go.mod or go.sum changes.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your source code into the container.
COPY . .

# --- DEBUGGING STEP ---
# List all files and directories recursively to verify the copy operation.
# Check the build output for this step to see if the 'cmd' directory exists.
RUN ls -R

# Build the Go application.
# CGO_ENABLED=0 creates a static binary without C dependencies.
# -ldflags="-w -s" strips debug information, reducing the binary size.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /cdn-speed-test ./cmd/cdn-speed-test

# --- Final Stage ---
# Use a minimal base image for the final container to keep it small and secure.
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy only the compiled binary from the builder stage.
COPY --from=builder /cdn-speed-test .

# Set the command to run when the container starts.
ENTRYPOINT ["./cdn-speed-test"]
