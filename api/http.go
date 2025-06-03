package api

import (
	"bytes"
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

type Client interface {
	Get(url string) ([]byte, error)
	Post(url string, data []byte) ([]byte, error)
	Patch(url string, data []byte) ([]byte, error)
	Put(url string, data []byte) ([]byte, error)
}

// HTTPClient is a simple HTTP client with authentication support.
// Access token is optional.
type HTTPClient struct {
	http        *http.Client
	AccessToken string
}

// NewHTTPClient creates a new HTTP client with the specified access token.
// Leave access token empty will let the client send unauthenticated requests.
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

func (c *HTTPClient) Patch(url string, data []byte) ([]byte, error) {
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return c.request(req)
}

func (c *HTTPClient) Put(url string, data []byte) ([]byte, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
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
