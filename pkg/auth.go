package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type TokenRequest struct {
	AppId string `json:"app_id"`
}

type Credentials struct {
	AccessKeyID     string    `json:"access_key"`
	SecretAccessKey string    `json:"secret_key"`
	SessionToken    string    `json:"session_token"`
	Expiration      time.Time `json:"expiration"`
}

func GetUploadCredentials(ctx context.Context, api string) (*Credentials, error) {
	requestBody, err := json.Marshal(TokenRequest{
		AppId: "12345",
	})
	if err != nil {
		return nil, err
	}

	// POST to /token on api to get credentials
	req, err := http.NewRequestWithContext(ctx, "POST", api+"/token", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read the response and parse json
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	credentials := &Credentials{}
	err = json.Unmarshal(body, credentials)
	if err != nil {
		return nil, err
	}

	return credentials, nil
}
