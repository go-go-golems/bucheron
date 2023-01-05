package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
)

func main() {
	// create an AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}

	// create an S3 client
	svc := s3.New(sess)

	// open the log file
	file, err := os.Open("/path/to/logfile.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// upload the log file to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("my-bucket"),
		Key:    aws.String("logfile.txt"),
		Body:   file,
		// add the comment as metadata
		Metadata: map[string]*string{
			"Comment": aws.String("This is a log file"),
		},
	})
	if err != nil {
		fmt.Println("Error uploading file:", err)
		return
	}

	fmt.Println("File uploaded successfully")
}
