package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type RequestError struct {
	StatusCode int
	Body       string
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("Failed requesting API: status code: %d, response: %s", e.StatusCode, e.Body)
}

func NewRequestError(statusCode int, body []byte) *RequestError {
	return &RequestError{
		StatusCode: statusCode,
		Body:       string(body),
	}
}

type HTTPClient struct {
	http        *http.Client
	AccessToken string
}

func NewHTTPClient(accessToken string) *HTTPClient {
	return &HTTPClient{
		http:        &http.Client{},
		AccessToken: accessToken,
	}
}

func (c *HTTPClient) Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.request(req)
}

func (c *HTTPClient) Post(url string, data []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return c.request(req)
}

func (c *HTTPClient) request(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	if c.AccessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error(fmt.Sprintf("failed to close response body: %v", err))
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, NewRequestError(res.StatusCode, body)
	}

	return body, nil
}

func AuthenticatedGetRequest(url string, access_token string, data any) error {
	req, err := buildAuthRequest("GET", url, access_token, nil)
	if err != nil {
		return err
	}

	return sendRequest(req, data)
}

func GetRequest(url string, data any) error {
	req, err := buildRequest(url)
	if err != nil {
		return err
	}

	return sendRequest(req, data)
}

// Post request is always authenticated
func PostRequest(url string, access_token string, jsonBytes []byte, data any) error {
	req, err := buildAuthRequest("POST", url, access_token, jsonBytes)
	if err != nil {
		return err
	}

	return sendRequest(req, data)
}

func PatchRequest(url string, access_token string, jsonBytes []byte, data any) error {
	req, err := buildAuthRequest("PATCH", url, access_token, jsonBytes)
	if err != nil {
		return err
	}
	return sendRequest(req, data)
}

func PutRequest(url string, access_token string, jsonBytes []byte, data any) error {
	req, err := buildAuthRequest("PUT", url, access_token, jsonBytes)
	if err != nil {
		return err
	}
	return sendRequest(req, data)
}

func sendRequest(req *http.Request, data any) error {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return NewRequestError(res.StatusCode, bodyBytes)
	}

	if len(bodyBytes) == 0 {
		return nil
	} else {
		err = json.Unmarshal(bodyBytes, data)
		return err
	}
}

func buildRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	return req, err
}

func buildAuthRequest(method string, url string, access_token string, jsonBytes []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access_token))
	return req, err
}
