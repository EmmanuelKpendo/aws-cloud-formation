#!/bin/bash
set -e

BUCKET_NAME="cloud-formation-artifact"
ZIP_PATH="user-creation-lambda.zip"

echo "[*] Building Go Lambda..."
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
zip user-creation-lambda.zip bootstrap
rm bootstrap

echo "[*] Uploading Lambda ZIP to s3://$BUCKET_NAME/$ZIP_PATH..."
aws s3 cp user-creation-lambda.zip s3://$BUCKET_NAME/$ZIP_PATH
rm user-creation-lambda.zip

echo "[✔] Upload complete."
