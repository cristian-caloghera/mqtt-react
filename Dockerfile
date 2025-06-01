FROM golang:1.24-bookworm AS development

WORKDIR /build

COPY mqtt-react.go go.mod go.sum /build

# Install dependencies
RUN go mod download

# Turn off CGO to ensure static binaries
RUN CGO_ENABLED=0 go build

# =======================================================
# Create a small image based on alpine to have a minimal 
# toolset (shell, echo, etc). 
FROM alpine

# Move to working directory /app
WORKDIR /app

# Copy binary from builder stage
COPY --from=development /build/mqtt-react /app

# copy the sample config too
COPY mqtt-react.yaml /app

# Start the application
ENTRYPOINT ["/app/mqtt-react"]

CMD ["/app/mqtt-react.yaml"]
