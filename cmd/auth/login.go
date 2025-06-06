package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/iucario/bangumi-go/api"
	"github.com/spf13/cobra"
)

var port = 9090

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to https://bgm.tv (bangumi.tv)",
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
	fmt.Println("Login to https://bgm.tv")
	apiAuth := "https://bgm.tv/oauth/authorize"
	responseType := "code"
	host := fmt.Sprintf("http://localhost:%d/auth", port)
	LOGIN_URL := fmt.Sprintf("%s?client_id=%s&response_type=%s&redirect_uri=%s", apiAuth, api.ClientId, responseType, host)

	serverDone := &sync.WaitGroup{}
	serverDone.Add(1)
	start(serverDone)

	openBrowser(LOGIN_URL)
	fmt.Println("If your browser is not opened automatically. Manually open this URL in browser and login:")
	fmt.Println(LOGIN_URL)

	serverDone.Wait()
}

func start(wg *sync.WaitGroup) {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			Client.GetAccessToken(code)
			fmt.Println("Login success.")
			w.Header().Set("Connection", "close")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintln(w, "Login success. You can close this page now.")
			// Give browser time to render response before shutdown
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := srv.Shutdown(ctx); err != nil {
					slog.Error("Server Shutdown", "error", err)
				}
			}()
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintln(w, "Hi")
		}
	})

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("ListenAndServe", "error", err)
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
