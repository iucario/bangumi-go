package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/iucario/bangumi-go/cmd/auth"
)

func AuthenticatedGetRequest(url string, access_token string, data interface{}) error {
	req, err := buildRequest(url)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access_token))

	return sendRequest(req, data)
}

func GetRequest(url string, data interface{}) error {
	req, err := buildRequest(url)
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
	if res.StatusCode != 200 {
		log.Fatalf("status code: %d, response: %s", res.StatusCode, bodyString)
		return fmt.Errorf("[error] status code: %d, response: %s", res.StatusCode, bodyString)
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
