#!/bin/bash
# Generate self-signed SSL certificates for local development

echo "🔐 Generating SSL certificates for HTTPS..."

# Create certs directory if it doesn't exist
mkdir -p certs

# Generate private key and certificate
openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 365 -nodes \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=*.local" \
  -addext "subjectAltName=DNS:localhost,DNS:*.local,IP:192.168.1.1,IP:192.168.0.1,IP:10.0.0.1"

echo "✅ Certificates generated in ./certs/"
echo "   - cert.pem (certificate)"
echo "   - key.pem (private key)"
echo ""
echo "⚠️  Note: You'll need to accept the security warning in your browser"
echo "    as this is a self-signed certificate."

