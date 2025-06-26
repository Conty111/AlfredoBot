#!/bin/bash
set -e

# Wait for MinIO to be ready
until mc alias set minio https://minio:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD --insecure >/dev/null 2>&1; do
  echo "Waiting for MinIO to be ready..."
  sleep 5
done

# Create bucket if not exists
if ! mc ls minio/$MINIO_BUCKET --insecure >/dev/null 2>&1; then
  echo "Creating bucket $MINIO_BUCKET"
  mc mb minio/$MINIO_BUCKET --insecure
  mc anonymous set download minio/$MINIO_BUCKET --insecure
else
  echo "Bucket $MINIO_BUCKET already exists"
fi

echo "MinIO bucket setup completed"
