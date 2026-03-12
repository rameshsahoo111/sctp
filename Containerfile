# ==========================================
# Stage 1: Build Environment
# ==========================================
FROM golang:1.22 AS builder

WORKDIR /app

# Initialize the Go module
RUN go mod init sctp-tools

# Copy source code
COPY sctp-server.go .
COPY sctp-client.go .

# Compile statically and strip debug symbols (-w -s) to shrink binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/sctp-server sctp-server.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/sctp-client sctp-client.go

# ==========================================
# Stage 2: Tiny Fedora Runtime
# ==========================================
# Use the official minimal Fedora image
FROM registry.fedoraproject.org/fedora-minimal:latest

WORKDIR /app

# Use microdnf to install strictly what is needed, including procps-ng for 'ps'
RUN microdnf install -y \
    tcpdump \
    iproute \
    iputils \
    nmap-ncat \
    procps-ng \
    && microdnf clean all

# Copy only the tiny compiled binaries from the builder stage
COPY --from=builder /app/sctp-server .
COPY --from=builder /app/sctp-client .

# Drop the user into a shell by default
CMD ["/bin/bash"]
