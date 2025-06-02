package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type SimpleEvent struct {
	UserName string `json:"userName"`
	Email    string `json:"email"`
}

type SecretData struct {
	Password string `json:"password"`
}

func handler(ctx context.Context, event events.CloudWatchEvent) (string, error) {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return "", err
	}

	log.Printf("Detail (string): %s", string(event.Detail))

	var evt SimpleEvent
	if err := json.Unmarshal(event.Detail, &evt); err != nil {
		log.Printf("Failed to unmarshal event detail: %v", err)
		return "", err
	}

	if evt.UserName == "" {
		log.Printf("Parsed event: %+v", evt)
		return "", fmt.Errorf("empty userName in event")
	}
	log.Printf("Processing user: %s", evt.UserName)

	// Initialize AWS clients
	ssmClient := ssm.NewFromConfig(cfg)
	secretsClient := secretsmanager.NewFromConfig(cfg)

	// Fetch email from SSM (even though it's included in the event, we replicate the Python logic)
	emailParam := fmt.Sprintf("/cf-users/%s/email", evt.UserName)
	emailOutput, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(emailParam),
	})
	if err != nil {
		log.Printf("Failed to get email parameter for %s: %v", evt.UserName, err)
		return "", err
	}
	email := aws.ToString(emailOutput.Parameter.Value)

	// Fetch one-time password from Secrets Manager
	secretOutput, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("OneTimePassword"),
	})
	if err != nil {
		log.Printf("Failed to get secret: %v", err)
		return "", err
	}

	var secret SecretData
	if err := json.Unmarshal([]byte(aws.ToString(secretOutput.SecretString)), &secret); err != nil {
		log.Printf("Failed to unmarshal secret string: %v", err)
		return "", err
	}
	password := secret.Password

	// Final log output
	log.Printf("New user created: %s, Email: %s, Temporary password: %s", evt.UserName, email, password)
	return fmt.Sprintf("User %s processed successfully.", evt.UserName), nil
}

func main() {
	lambda.Start(handler)
}
