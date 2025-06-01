#!/bin/bash
set -e

BUCKET_NAME="cloud-formation-artifact"
ZIP_PATH="lambda-artifacts/user-logger.zip"

echo "[*] Building Go Lambda..."
GOOS=linux GOARCH=amd64 go build -o main main.go
zip user-logger.zip main
rm main

echo "[*] Uploading Lambda ZIP to s3://$BUCKET_NAME/$ZIP_PATH..."
aws s3 cp user-logger.zip s3://$BUCKET_NAME/$ZIP_PATH
rm user-logger.zip

echo "[✔] Upload complete."
