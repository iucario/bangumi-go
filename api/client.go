package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/iucario/bangumi-go/util"
)

const (
	API       string = "https://api.bgm.tv/v0"
	API_AUTH  string = "https://bgm.tv/oauth/access_token"
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
func SaveCredential(credential Credential) {
	configDir := util.ConfigDir()
	err := os.MkdirAll(configDir, 0o755)
	AbortOnError(err)

	jsonBytes, err := json.Marshal(credential)
	AbortOnError(err)

	credentialPath := fmt.Sprintf("%s/credential.json", configDir)
	err = os.WriteFile(credentialPath, jsonBytes, 0o644)
	AbortOnError(err)
}

// Handle all errors and refresh token. Throw error if a login is required.
func GetCredential() (*Credential, error) {
	credential, err := loadCredential()
	if err != nil {
		return nil, err
	}
	authClient := NewAuthClient(credential.AccessToken)
	statusFlag := authClient.GetStatus()
	if statusFlag {
		return &credential, nil
	}

	newCredential, err := authClient.RefreshToken()
	if err != nil {
		return nil, err
	}
	return newCredential, nil
}

func AbortOnError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		slog.Error(err.Error())
		os.Exit(1)
	}
}

// Load credential JSON from file
func loadCredential() (Credential, error) {
	configDir := util.ConfigDir()
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
