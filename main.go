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
	//Initialize AWS SDK config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("failed to load AWS config: %v", err)
		return "Failed to load AWS config", err
	}

	//Parse cloudTrail Event
	var cloudTrailEvent CloudTrailEvent
	if err := json.Unmarshal(event.Detail, &cloudTrailEvent); err != nil {
		log.Printf("failed to unmarshal CloudTrailEvent: %v", err)
		return "Failed to unmarshal CloudTrailEvent", err
	}
	userName := cloudTrailEvent.Detail.RequestParameters.UserName

	//Initialize SSM Client
	ssmClient := ssm.NewFromConfig(cfg)
	emailParam := fmt.Sprintf("lab-2-user-email-%s", userName)

	//Get Email from Parameter Store
	emailOutput, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name: &emailParam,
	})
	var email string
	if err != nil {
		log.Printf("failed to get email parameter for %s: %v", userName, err)
	} else {
		email = *emailOutput.Parameter.Value
	}

	//Initialize secret manager client
	secretsClient := secretsmanager.NewFromConfig(cfg)
	secretsOutput, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("OneTimePasswordSecret"),
	})
	var password string
	if err != nil {
		log.Printf("failed to get secret: %v", err)
		password = "Failed to retrieve password"
	} else {
		var secretData SecretData
		if err := json.Unmarshal([]byte(*secretsOutput.SecretString), &secretData); err != nil {
			log.Printf("failed to unmarshal secret: %v", err)
			password = "Failed to parse password"
		} else {
			password = secretData.Password
		}
	}

	//log the details
	log.Printf("New user created: %s, Email: %s, Temporary password: %s", userName, email, password)
	return "Logged user creation", nil
}

func main() {
	lambda.Start(handler)
}
