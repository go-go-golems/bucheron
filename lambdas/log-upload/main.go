package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
)

func main() {
	lambda.Start(handleUpload)
}

func handleUpload(sqsEvent events.SQSEvent) error {
	// create an AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return fmt.Errorf("Error creating session: %v", err)
	}

	// create an S3 pkg
	s3Svc := s3.New(sess)

	// create an SES pkg
	sesSvc := ses.New(sess)

	for _, message := range sqsEvent.Records {
		// get the S3 bucket and key from the message body
		bucket := message.Body
		key := message.Body

		// get the log file from S3
		result, err := s3Svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return fmt.Errorf("Error getting object from S3: %v", err)
		}
		defer result.Body.Close()

		// get the comment from the metadata
		comment := *result.Metadata["Comment"]

		// send an email using SES
		_, err = sesSvc.SendEmail(&ses.SendEmailInput{
			Destination: &ses.Destination{
				ToAddresses: []*string{
					aws.String("you@example.com"),
				},
			},
			Message: &ses.Message{
				Subject: &ses.Content{
					Data: aws.String("Log file uploaded"),
				},
				Body: &ses.Body{
					Text: &ses.Content{
						Data: aws.String(fmt.Sprintf("A log file was uploaded to S3 with the following comment: %s", comment)),
					},
				},
			},
			Source: aws.String("you@example.com"),
		})
		if err != nil {
			return fmt.Errorf("Error sending email: %v", err)
		}
	}

	return nil
}
