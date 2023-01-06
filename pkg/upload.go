package pkg

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
)

type UploadSettings struct {
	// The AWS region that the bucket is in.
	Region      string
	Bucket      string
	Credentials *Credentials
}

type UploadData struct {
	Files    []string
	Comment  string
	Metadata map[string]interface{}
}

func UploadLogs(settings *UploadSettings, data *UploadData) error {
	// Set up a new AWS session using the temporary AWS credentials.
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			settings.Credentials.AccessKeyID,
			settings.Credentials.SecretAccessKey,
			settings.Credentials.SessionToken),
		Region: aws.String(settings.Region),
	})
	if err != nil {
		panic(err)
	}

	// Create a new S3 pkg.
	s3Client := s3.New(sess)

	// Read the file that you want to upload.
	fileBytes, err := os.ReadFile("example.txt")
	if err != nil {
		panic(err)
	}

	s3Metadata := make(map[string]*string)
	for key, value := range data.Metadata {
		s3Metadata[key] = aws.String(fmt.Sprintf("%v", value))
	}
	s3Metadata["Comment"] = aws.String(data.Comment)

	// Upload the file to S3.
	_, err = s3Client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket:   aws.String(settings.Bucket),
		Key:      aws.String("example.txt"),
		Body:     bytes.NewReader(fileBytes),
		Metadata: s3Metadata,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully uploaded file to S3")
	return nil
}
