package cmd

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/Aravind-pk/shush/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	loginServer  string
	loginTimeout time.Duration
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Shush backend",
	Long: `Opens your browser to log in to the Shush server. The CLI starts a
temporary local web server to receive the auth callback, then writes
your token to ~/.shush/credentials.json.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogin(cmd.Context())
	},
}

func init() {
	loginCmd.Flags().StringVar(&loginServer, "server", "http://localhost:8080",
		"Shush server URL")
	loginCmd.Flags().DurationVar(&loginTimeout, "timeout", 5*time.Minute,
		"how long to wait for the browser callback")
	rootCmd.AddCommand(loginCmd)
}

// callbackResult is what the HTTP handler sends back to the main goroutine.
// Exactly one of token / err is meaningful.
type callbackResult struct {
	token        string
	refreshToken string
	expiresAt    time.Time
	err          error
}

func runLogin(ctx context.Context) error {
	// 1. Bind an ephemeral port on loopback only.
	//    "127.0.0.1:0" — the :0 tells the kernel "pick any free port",
	//    127.0.0.1 ensures nothing on the LAN can reach us.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("bind callback listener: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// 2. Generate a CSRF state token. The backend must echo it back
	//    unchanged; if a request arrives with the wrong state, we reject it.
	state, err := randomState()
	if err != nil {
		return fmt.Errorf("generate state: %w", err)
	}

	// 3. Set up the callback channel + HTTP handler.
	//    Buffer size 1 so the handler never blocks even if main has
	//    moved on (e.g. context timeout fired first).
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if got := q.Get("state"); got != state {
			// Wrong state — possible CSRF. Don't accept the token, don't
			// confirm to the caller that we received their request.
			http.Error(w, "invalid state", http.StatusBadRequest)
			// Don't send to resultCh — keep waiting for a legitimate callback.
			return
		}

		token := q.Get("token")
		if token == "" {
			msg := q.Get("error")
			if msg == "" {
				msg = "missing token"
			}
			http.Error(w, msg, http.StatusBadRequest)
			resultCh <- callbackResult{err: fmt.Errorf("login failed: %s", msg)}
			return
		}

		// Optional refresh token + expiry; the backend may or may not send these.
		refresh := q.Get("refresh_token")
		var exp time.Time
		if expStr := q.Get("expires_at"); expStr != "" {
			if t, err := time.Parse(time.RFC3339, expStr); err == nil {
				exp = t
			}
		}
		if exp.IsZero() {
			// Reasonable default until the backend wires this up.
			exp = time.Now().Add(15 * time.Minute)
		}

		// Friendly success page in the browser.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(successHTML))

		resultCh <- callbackResult{
			token:        token,
			refreshToken: refresh,
			expiresAt:    exp,
		}
	})

	srv := &http.Server{Handler: mux}

	// 4. Serve in the background. srv.Serve returns http.ErrServerClosed
	//    on a clean Shutdown — that's not a real error.
	serveErr := make(chan error, 1)
	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()

	// 5. Build the URL we'll send the user's browser to and open it.
	authURL := fmt.Sprintf("%s/cli/login?redirect_uri=%s&state=%s",
		loginServer, callbackURL, state)

	fmt.Println("Opening your browser to:")
	fmt.Println("  ", authURL)
	fmt.Println()
	fmt.Println("If your browser doesn't open, paste that URL into it manually.")
	fmt.Printf("(Listening for callback on %s — timeout %s)\n", callbackURL, loginTimeout)

	if err := openBrowser(authURL); err != nil {
		// Not fatal — user can paste the URL themselves.
		fmt.Printf("(Couldn't auto-open browser: %v)\n", err)
	}

	// 6. Wait for the callback, the user's ctrl-C, or a timeout.
	timeoutCtx, cancel := context.WithTimeout(ctx, loginTimeout)
	defer cancel()

	var result callbackResult
	select {
	case result = <-resultCh:
		// got something
	case err := <-serveErr:
		return fmt.Errorf("callback server failed: %w", err)
	case <-timeoutCtx.Done():
		_ = srv.Close()
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("login timed out after %s", loginTimeout)
		}
		return timeoutCtx.Err()
	}

	// 7. Graceful shutdown — give the success page time to flush to the browser.
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	_ = srv.Shutdown(shutdownCtx)

	if result.err != nil {
		return result.err
	}

	// 8. Persist credentials.
	creds := &config.Credentials{
		ServerURL:    loginServer,
		AccessToken:  result.token,
		RefreshToken: result.refreshToken,
		ExpiresAt:    result.expiresAt,
	}
	if err := config.Save(creds); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	path, _ := config.Path()
	fmt.Println()
	fmt.Println("Logged in. Credentials saved to", path)
	return nil
}

// randomState returns a 32-byte cryptographically random URL-safe string,
// used as the CSRF state parameter on the auth round-trip.
func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// openBrowser tries to launch the user's default browser at url.
// Failure here is non-fatal — the URL is also printed to stdout.
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default: // linux, bsd, etc.
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}

//go:embed success.html
var successHTML string
