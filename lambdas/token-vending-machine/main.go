package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func handler(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	// log credentials from environment
	accessKeyId := os.Getenv("TOKEN_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("TOKEN_SECRET_ACCESS_KEY")
	log.Printf("AWS_ACCESS_KEY_ID: %s\n", accessKeyId)
	log.Printf("AWS_SECRET: %s\n", secretAccessKey)

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyId, secretAccessKey, ""),
	})
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to create session"))
	}

	svc := sts.New(sess)

	// Call the STS GetCallerIdentity action to retrieve the caller identity
	callerIdentity, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to get caller identity"))
	}

	// Log the caller identity
	log.Printf("Caller identity: %+v", callerIdentity)

	input := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(300), // 5 minutes
	}

	result, err := svc.GetSessionToken(input)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to get session token"))
	}

	formattedExpiration := result.Credentials.Expiration.Format(time.RFC3339)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body: fmt.Sprintf(
			`{"access_key": "%s", "secret_key": "%s", "session_token": "%s", "expiration": "%s"}`,
			*result.Credentials.AccessKeyId,
			*result.Credentials.SecretAccessKey,
			*result.Credentials.SessionToken,
			formattedExpiration,
		),
	}, nil
}

func main() {
	lambda.Start(handler)
}
