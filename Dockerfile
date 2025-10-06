FROM alpine:latest

# Install git and ca-certificates
RUN apk --no-cache add git ca-certificates

# Create a non-root user
RUN adduser -D -s /bin/sh ccswitch

# Copy the binary
COPY ccswitch /usr/local/bin/ccswitch

# Make it executable
RUN chmod +x /usr/local/bin/ccswitch

# Switch to non-root user
USER ccswitch

# Set the entrypoint
ENTRYPOINT ["ccswitch"]