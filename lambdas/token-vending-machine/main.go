package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func handler(ctx context.Context) (string, error) {
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
		DurationSeconds: aws.Int64(3600), // 1 hour
	}

	result, err := svc.GetSessionToken(input)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to get session token"))
	}

	// Return the access key, secret key, and session token as a JSON string
	return fmt.Sprintf(
		`{"access_key": "%s", "secret_key": "%s", "session_token": "%s"}`,
		*result.Credentials.AccessKeyId,
		*result.Credentials.SecretAccessKey,
		*result.Credentials.SessionToken,
	), nil
}

func main() {
	lambda.Start(handler)
}
