#!/bin/bash
set -e

# Check required environment variables
if [ -z "${MINIO_BUCKET}" ] || [ -z "${MINIO_ROOT_USER}" ] || [ -z "${MINIO_ROOT_PASSWORD}" ]; then
  echo "ERROR: Required environment variables MINIO_BUCKET, MINIO_ROOT_USER or MINIO_ROOT_PASSWORD are not set"
  exit 1
fi

# Wait for MinIO to be ready (max 30 seconds)
echo "Waiting for MinIO to be ready..."
for i in {1..30}; do
  if mc alias set myminio https://minio:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD} --insecure >/dev/null 2>&1; then
    echo "MinIO connection established"
    break
  fi
  echo "MinIO not ready yet, waiting... (attempt $i/30)"
  sleep 1
done

# Verify MinIO connection
if ! mc ls myminio >/dev/null 2>&1; then
  echo "ERROR: Failed to connect to MinIO after 30 seconds"
  exit 1
fi

echo "Creating bucket ${MINIO_BUCKET}..."
if ! mc mb --ignore-existing myminio/${MINIO_BUCKET}; then
  echo "ERROR: Failed to create bucket ${MINIO_BUCKET}"
  exit 1
fi

echo "Setting bucket policy..."
if ! mc policy set public myminio/${MINIO_BUCKET}; then
  echo "ERROR: Failed to set bucket policy"
  exit 1
fi

echo "MinIO bucket ${MINIO_BUCKET} successfully created and configured"
