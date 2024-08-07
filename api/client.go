package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/iucario/bangumi-go/cmd/auth"
)

func AuthenticatedGetRequest(url string, access_token string, data interface{}) error {
	req, err := buildAuthRequest("GET", url, access_token, nil)
	if err != nil {
		return err
	}

	return sendRequest(req, data)
}

func GetRequest(url string, data interface{}) error {
	req, err := buildRequest(url)
	if err != nil {
		return err
	}

	return sendRequest(req, data)
}

// Post request is always authenticated
func PostRequest(url string, access_token string, jsonBytes []byte, data interface{}) error {
	req, err := buildAuthRequest("POST", url, access_token, jsonBytes)
	if err != nil {
		return err
	}

	return sendRequest(req, data)
}

func PatchRequest(url string, access_token string, jsonBytes []byte, data interface{}) error {
	req, err := buildAuthRequest("PATCH", url, access_token, jsonBytes)
	if err != nil {
		return err
	}
	return sendRequest(req, data)
}

func PutRequest(url string, access_token string, jsonBytes []byte, data interface{}) error {
	req, err := buildAuthRequest("PUT", url, access_token, jsonBytes)
	if err != nil {
		return err
	}
	return sendRequest(req, data)
}

func sendRequest(req *http.Request, data interface{}) error {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	bodyString := string(bodyBytes)
	if res.StatusCode >= 300 {
		slog.Error(fmt.Sprintf("status code: %d, response: %s", res.StatusCode, bodyString))
		return fmt.Errorf(fmt.Sprintf("[error] status code: %d, response: %s", res.StatusCode, bodyString))
	}

	err = json.Unmarshal(bodyBytes, data)
	if err != nil {
		return err
	}
	return nil
}

func buildRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	return req, err
}

func buildAuthRequest(method string, url string, access_token string, jsonBytes []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access_token))
	return req, err
}
