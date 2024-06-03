package auth

import (
	"encoding/json"
	"fmt"
	"log"
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
		GetStatus(credential.AccessToken)
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
	Check(err)

	client := &http.Client{}
	res, err := client.Do(req)
	Check(err)

	if res.StatusCode == 200 {
		fmt.Println("Auth status: OK")
		status := AuthStatus{}
		err = json.NewDecoder(res.Body).Decode(&status)
		Check(err)
		log.Println("Auth status OK", status)
		return true
	}
	fmt.Println("Auth status: Failed")
	return false
}

func init() {
	authCmd.AddCommand(statusCmd)
}
