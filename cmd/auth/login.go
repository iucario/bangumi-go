package auth

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os/exec"
	"runtime"
	"sync"

	"github.com/iucario/bangumi-go/api"
	"github.com/spf13/cobra"
)

var ctxShutdown, cancel = context.WithCancel(context.Background())

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to https://bangumi.tv",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := api.LoadCredential()
		if err != nil {
			// Token does not exist, login from browser
			BrowserLogin()
		} else if Client.GetStatus() {
			fmt.Println("Token is still valid")
			return
		}
		_, err = Client.RefreshToken()
		if err != nil {
			BrowserLogin()
		}
	},
}

func BrowserLogin() {
	fmt.Println("Login to https://bangumi.tv")
	apiAuth := "https://bgm.tv/oauth/authorize"
	responseType := "code"
	host := "http://localhost:9090/auth"
	LOGIN_URL := fmt.Sprintf("%s?client_id=%s&response_type=%s&redirect_uri=%s", apiAuth, api.ClientId, responseType, host)

	openBrowser(LOGIN_URL)
	fmt.Println("If your browser is not opened automatically. Manually open this URL in browser and login:")
	fmt.Println(LOGIN_URL)

	serverDone := &sync.WaitGroup{}
	serverDone.Add(1)
	Start(serverDone)
	serverDone.Wait()
}

func Start(wg *sync.WaitGroup) {
	srv := &http.Server{Addr: ":9090"}
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-ctxShutdown.Done():
			fmt.Println("Context Shuting down ...")
			return
		default:
		}
		code := r.URL.Query().Get("code")
		if code != "" {
			Client.GetAccessToken(code)
			// shutdown
			cancel()
			err := srv.Shutdown(context.Background())
			if err != nil {
				slog.Info("server.Shutdown:", "Error", err)
			}
		} else {
			fmt.Fprintln(w, "Hi") // Server HTML page to fetch token and return to server at /auth
		}
	})

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error(fmt.Sprintf("ListenAndServe(): %v", err))
		}
	}()
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	authCmd.AddCommand(loginCmd)
}
