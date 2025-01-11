package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type AuthStatus struct {
	AccessToken string `json:"access_token"`
	ClientId    string `json:"client_id"`
	Expires     int    `json:"expires"`
	Scope       string `json:"scope"`
	UserId      int    `json:"user_id"`
}

func GetStatus(accessToken string) bool {
	api := "https://bgm.tv/oauth/token_status?access_token=" + accessToken
	req, err := http.NewRequest("POST", api, nil)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	AbortOnError(err)

	client := &http.Client{}
	res, err := client.Do(req)
	AbortOnError(err)

	if res.StatusCode == 200 {
		slog.Info("auth status: OK")
		status := AuthStatus{}
		err = json.NewDecoder(res.Body).Decode(&status)
		AbortOnError(err)
		slog.Info(fmt.Sprintln("auth status: OK", status))
		return true
	}
	slog.Warn("auth status: Failed")
	body, err := io.ReadAll(res.Body)
	AbortOnError(err)

	slog.Info(fmt.Sprintln("response Body:", string(body)))

	return false
}

func GetAccessToken(code, configDir string) {
	data := []byte(fmt.Sprintf(`{
		"grant_type": "%s",
		"client_id": "%s",
		"client_secret": "%s",
		"code": "%s",
		"redirect_uri": "http://localhost:9090/auth"
	}`, GrantType, ClientId, AppSecret, code))
	req, err := http.NewRequest("POST", API, bytes.NewBuffer(data))
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	AbortOnError(err)
	defer req.Body.Close()

	client := &http.Client{}
	res, err := client.Do(req)
	AbortOnError(err)

	credential := Credential{}
	err = json.NewDecoder(res.Body).Decode(&credential)
	AbortOnError(err)

	SaveCredential(credential, configDir)
	fmt.Println(res.StatusCode)
}
