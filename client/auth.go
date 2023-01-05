package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// TokenRequest represents the request payload for the token vending machine.
type TokenRequest struct {
	// UserID is the ID of the user for whom to generate temporary credentials.
	UserID string `json:"user_id"`
}

// TokenResponse represents the response payload from the token vending machine.
type TokenResponse struct {
	// AccessKey is the access key for the temporary AWS credentials.
	AccessKey string `json:"access_key"`
	// SecretKey is the secret key for the temporary AWS credentials.
	SecretKey string `json:"secret_key"`
	// SessionToken is the session token for the temporary AWS credentials.
	SessionToken string `json:"session_token"`
	// Expiration is the expiration time of the temporary AWS credentials.
	Expiration time.Time `json:"expiration"`
}

func GetToken() {
	// Set up the HTTP client.
	client := &http.Client{}

	// Set up the request to the token vending machine.
	requestBody, err := json.Marshal(TokenRequest{
		UserID: "12345",
	})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", "https://example.com/token", bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request to the token vending machine.
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the response from the token vending machine.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Parse the response from the token vending machine.
	var tokenResponse TokenResponse
	err = json.Unmarshal(respBody, &tokenResponse)
	if err != nil {
		panic(err)
	}

	uploadLogs(err, tokenResponse)
}

func uploadLogs(err error, tokenResponse TokenResponse) {
	// Set up a new AWS session using the temporary AWS credentials.
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(tokenResponse.AccessKey, tokenResponse.SecretKey, tokenResponse.SessionToken),
		Region:      aws.String("us-east-1"),
	})
	if err != nil {
		panic(err)
	}

	// Create a new S3 client.
	s3Client := s3.New(sess)

	// Read the file that you want to upload.
	fileBytes, err := ioutil.ReadFile("example.txt")
	if err != nil {
		panic(err)
	}

	// Upload the file to S3.
	_, err = s3Client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String("example-bucket"),
		Key:    aws.String("example.txt"),
		Body:   bytes.NewReader(fileBytes),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully uploaded file to S3")
}
