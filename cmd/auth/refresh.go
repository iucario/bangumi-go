package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh token",
	Run: func(cmd *cobra.Command, args []string) {
		err := RefreshToken()
		if err != nil {
			slog.Error(fmt.Sprintln("Failed to refresh token", err))
		} else {
			slog.Info("Refresh token success")
		}
	},
}

func RefreshToken() error {
	credential, err := LoadCredential()
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

	SaveCredential(newCredential)
	return nil
}

func init() {
	authCmd.AddCommand(refreshCmd)
}
