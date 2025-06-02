package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type CloudTrailEvent struct {
	Detail struct {
		RequestParameters struct {
			UserName string `json:"userName"`
		} `json:"requestParameters"`
	} `json:"detail"`
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

	// Parse the CloudTrail event
	var ctEvent CloudTrailEvent
	if err := json.Unmarshal(event.Detail, &ctEvent); err != nil {
		log.Printf("Failed to unmarshal CloudTrail event: %v", err)
		return "", err
	}

	userName := ctEvent.Detail.RequestParameters.UserName
	if userName == "" {
		err := fmt.Errorf("empty userName in event")
		log.Printf("Error: %v", err)
		return "", err
	}
	log.Printf("Processing user: %s", userName)

	// Initialize SSM and Secrets Manager clients
	ssmClient := ssm.NewFromConfig(cfg)
	secretsClient := secretsmanager.NewFromConfig(cfg)

	// Fetch user email from SSM
	emailParam := fmt.Sprintf("/cf-users/%s/email", userName)
	emailOutput, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(emailParam),
	})
	if err != nil {
		log.Printf("Failed to get email parameter for %s: %v", userName, err)
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

	// Final logging
	log.Printf("New user created: %s, Email: %s, Temporary password: %s", userName, email, password)
	return "User creation logged successfully", nil
}

func main() {
	lambda.Start(handler)
}
