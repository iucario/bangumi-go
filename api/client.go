package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

const (
	API       string = "https://bgm.tv/oauth/access_token"
	GrantType string = "authorization_code"
	UserAgent string = "iucario/bangumi-go"
	AppSecret string = "f4f057619facdba407afb48c9dce9114"
	ClientId  string = "bgm250163bec16210c2d"
)

type Credential struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	UserId       int    `json:"user_id"`
}

// Save credential to file
func SaveCredential(credential Credential, configDir string) {
	err := os.MkdirAll(configDir, 0o755)
	AbortOnError(err)

	jsonBytes, err := json.Marshal(credential)
	AbortOnError(err)

	credentialPath := fmt.Sprintf("%s/credential.json", configDir)
	err = os.WriteFile(credentialPath, jsonBytes, 0o644)
	AbortOnError(err)
}

// Load credential JSON from file
func LoadCredential(configDir string) (Credential, error) {
	credentialPath := fmt.Sprintf("%s/credential.json", configDir)
	jsonBytes, err := os.ReadFile(credentialPath)
	if err != nil {
		return Credential{}, err
	}
	credential := Credential{}
	err = json.Unmarshal(jsonBytes, &credential)
	if err != nil {
		return Credential{}, err
	}
	return credential, nil
}

// Handle all errors and refresh token. Throw error if a login is required.
func GetCredential(configDir string) (Credential, error) {
	credential, err := LoadCredential(configDir)
	if err != nil {
		return Credential{}, err
	}
	statusFlag := GetStatus(credential.AccessToken)
	if statusFlag {
		return credential, nil
	}

	err = RefreshToken(configDir)
	if err != nil {
		return Credential{}, err
	}
	credential, _ = LoadCredential(configDir)
	return credential, nil
}

func RefreshToken(configDir string) error {
	credential, err := LoadCredential(configDir)
	if err != nil {
		fmt.Println("No token found")
		return err
	}

	data := []byte(fmt.Sprintf(`{
		"grant_type": "refresh_token",
		"client_id": "%s",
		"client_secret": "%s",
		"refresh_token": "%s",
		"redirect_uri": "http://localhost:9090/auth"
	}`, ClientId, AppSecret, credential.RefreshToken))

	req, err := http.NewRequest("POST", API, bytes.NewBuffer(data))
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}
	defer req.Body.Close()

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Failed to refresh token")
	}

	newCredential := Credential{}
	err = json.NewDecoder(res.Body).Decode(&newCredential)
	if err != nil {
		return err
	}

	SaveCredential(newCredential, configDir)
	return nil
}

func AbortOnError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		slog.Error(err.Error())
		os.Exit(1)
	}
}
