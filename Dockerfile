
# Build stage
FROM golang:1.21-alpine AS builder

# Install required packages
RUN apk add --no-cache git make

# Set working directory
WORKDIR /src

# Clone CoreDNS repository
RUN git clone https://github.com/coredns/coredns.git .

# Clone the ratelimit plugin
RUN git clone https://github.com/milgradesec/ratelimit.git /src/ratelimit

# Enable the ratelimit plugin
RUN sed -i '/^hosts:hosts/a ratelimit:github.com/milgradesec/ratelimit' plugin.cfg

# Modify coredns.go to include ratelimit
RUN sed -i '/	_ "github.com\/coredns\/coredns\/plugin\/whoami"/a \	_ "github.com/milgradesec/ratelimit"' coredns.go

# Update go.mod to include ratelimit plugin
RUN go mod edit -require github.com/milgradesec/ratelimit@v1.1.0
RUN go mod edit -replace github.com/milgradesec/ratelimit=/src/ratelimit
RUN go mod tidy

# Build CoreDNS
RUN make

# Final stage
FROM alpine:latest

# Copy the built binary from the builder stage
COPY --from=builder /src/coredns /coredns

# Copy a default Corefile
COPY Corefile /Corefile

# Expose DNS ports
EXPOSE 53 53/udp

# Set the entrypoint
ENTRYPOINT ["/coredns"]
CMD ["-conf", "/Corefile"]
