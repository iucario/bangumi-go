package auth

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/user"

	"github.com/iucario/bangumi-go/cmd"
	"github.com/spf13/cobra"
)

const API string = "https://bgm.tv/oauth/access_token"
const GrantType string = "authorization_code"
const UserAgent string = "iucario/bangumi-go"
const AppSecret string = "f4f057619facdba407afb48c9dce9114"
const ClientId string = "bgm250163bec16210c2d"

var ConfigDir string

type Credential struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	UserId       int    `json:"user_id"`
}

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Auth commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`Available commands: 
bgm auth login
bgm auth logout
bgm auth status`)
	},
}

func init() {
	cmd.RootCmd.AddCommand(authCmd)

	usr, err := user.Current()
	if err != nil {
		slog.Error(err.Error())
	}
	ConfigDir = usr.HomeDir + "/.config/bangumi-go"
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// authCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// authCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Save credential to file
func SaveCredential(credential Credential) {
	err := os.MkdirAll(ConfigDir, 0755)
	AbortOnError(err)

	jsonBytes, err := json.Marshal(credential)
	AbortOnError(err)

	credentialPath := fmt.Sprintf("%s/credential.json", ConfigDir)
	err = os.WriteFile(credentialPath, jsonBytes, 0644)
	AbortOnError(err)
}

// Load credential from file
func LoadCredential() (Credential, error) {
	credentialPath := fmt.Sprintf("%s/credential.json", ConfigDir)
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

func AbortOnError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		slog.Error(err.Error())
		os.Exit(1)
	}
}

// Handle all errors and refresh token. Throw error if a login is required.
func GetCredential() (Credential, error) {
	credential, err := LoadCredential()
	if err != nil {
		slog.Error("credential file does not exist")
		return Credential{}, err
	}
	statusFlag := GetStatus(credential.AccessToken)
	if statusFlag {
		return credential, nil
	}

	err = RefreshToken()
	if err != nil {
		return Credential{}, err
	}
	credential, _ = LoadCredential()
	return credential, nil
}
