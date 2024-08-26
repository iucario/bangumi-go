package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show auth status",
	Run: func(cmd *cobra.Command, args []string) {
		credential, err := LoadCredential()
		if err != nil {
			fmt.Println("No credential found.")
			return
		}
		statusFlag := GetStatus(credential.AccessToken)
		if statusFlag {
			fmt.Println("Auth status: OK")
		} else {
			fmt.Println("Auth status: Failed")
		}
	},
}

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

func init() {
	authCmd.AddCommand(statusCmd)
}
