#!/bin/bash
set -e

BUCKET_NAME="cloud-formation-artifact"
ZIP_PATH="user-creation-lambda.zip"

echo "[*] Checking if S3 bucket exists..."
if ! aws s3api head-bucket --bucket "$BUCKET_NAME" 2>/dev/null; then
  echo "[*] Bucket not found. Creating bucket: $BUCKET_NAME"
  aws s3api create-bucket --bucket "$BUCKET_NAME" --region $(aws configure get region) --create-bucket-configuration LocationConstraint=$(aws configure get region)
else
  echo "[*] Bucket $BUCKET_NAME already exists."
fi

echo "[*] Building Go Lambda..."
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
zip "$ZIP_PATH" bootstrap
rm bootstrap

echo "[*] Uploading Lambda ZIP to s3://$BUCKET_NAME/$ZIP_PATH..."
aws s3 cp "$ZIP_PATH" "s3://$BUCKET_NAME/$ZIP_PATH"
rm "$ZIP_PATH"

echo "[✔] Upload complete."
