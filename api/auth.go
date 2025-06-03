package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

// TODO: should have token field and HTTPClient should not have it
type AuthClient struct {
	*HTTPClient
}

func NewAuthClient(accessToken string) *AuthClient {
	return &AuthClient{
		HTTPClient: NewHTTPClient(accessToken),
	}
}

// NewAuthClientWithConfig creates a new AuthClient with the access token loaded from the credential file.
// Returns a new AuthClient with empty access token if loading credential fails.
func NewAuthClientWithConfig() *AuthClient {
	if credential, err := LoadCredential(); err == nil {
		return &AuthClient{
			HTTPClient: NewHTTPClient(credential.AccessToken),
		}
	}
	slog.Error("failed loading credential")
	return &AuthClient{
		HTTPClient: NewHTTPClient(""),
	}
}

type AuthStatus struct {
	AccessToken string `json:"access_token"`
	ClientId    string `json:"client_id"`
	Expires     int    `json:"expires"`
	Scope       string `json:"scope"`
	UserId      int    `json:"user_id"`
}

// Get token status from the API.
func (c *AuthClient) GetStatus() bool {
	api := "https://bgm.tv/oauth/token_status"
	b, err := c.Get(api)
	if err == nil {
		return true
	}
	slog.Error(fmt.Sprintf("get status: %v", err))

	status := AuthStatus{}
	if err := json.Unmarshal(b, &status); err == nil {
		slog.Info(fmt.Sprintf("auth status: %s", status.AccessToken))
	}

	return false
}

type AccessPayload struct {
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	RedirectUri  string `json:"redirect_uri"`
}

func (c *AuthClient) GetAccessToken(code string) {
	payload := AccessPayload{
		GrantType:    GrantType,
		ClientId:     ClientId,
		ClientSecret: AppSecret,
		Code:         code,
		RedirectUri:  "http://localhost:9090/auth",
	}
	data, err := json.Marshal(payload)
	AbortOnError(err)
	b, err := c.Post(API_AUTH, data)
	AbortOnError(err)

	credential := Credential{}
	err = json.Unmarshal(b, &credential)
	if err != nil {
		slog.Error(fmt.Sprintf("unmarshalling credential: %v", err))
		return
	}

	// Update the access token in the AuthClient
	c.AccessToken = credential.AccessToken

	SaveCredential(credential)
}

type RefreshPayload struct {
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	RedirectUri  string `json:"redirect_uri"`
}

// RefreshToken refreshes the access token using the refresh token stored in the credential file.
// It returns the new credential or an error if the refresh token is invalid or expired.
// Updates the token in AuthClient and saves the new credential to the file.
func (c *AuthClient) RefreshToken() (*Credential, error) {
	credential, err := LoadCredential()
	if err != nil {
		fmt.Println("No token found")
		return nil, err
	}

	payload := RefreshPayload{
		GrantType:    "refresh_token",
		ClientId:     ClientId,
		ClientSecret: AppSecret,
		RefreshToken: credential.RefreshToken,
		RedirectUri:  "http://localhost:9090/auth",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	b, err := c.Post(API_AUTH, data)
	if err != nil {
		return nil, err
	}

	newCredential := Credential{}
	err = json.Unmarshal(b, &newCredential)
	if err != nil {
		return nil, err
	}

	// Update the access token in the AuthClient
	c.AccessToken = newCredential.AccessToken

	SaveCredential(newCredential)
	return &newCredential, nil
}
