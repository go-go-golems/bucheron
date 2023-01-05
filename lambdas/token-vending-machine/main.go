package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func handler(ctx context.Context) (string, error) {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	svc := sts.New(sess)

	input := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(3600), // 1 hour
	}

	result, err := svc.GetSessionToken(input)
	if err != nil {
		log.Fatal(err)
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
