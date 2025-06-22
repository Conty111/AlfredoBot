#!/bin/bash
set -e

# Create certificates directory if it doesn't exist
mkdir -p certs

# Generate private key
openssl genrsa -out certs/private.key 2048

# Generate self-signed certificate
openssl req -new -x509 -days 365 -key certs/private.key -out certs/public.crt -subj "/C=US/ST=State/L=City/O=Organization/CN=minio"

# Set permissions
chmod 600 certs/private.key
chmod 644 certs/public.crt

echo "Self-signed certificates for MinIO generated successfully!"
echo "Place these files in the ./certs directory for MinIO to use."