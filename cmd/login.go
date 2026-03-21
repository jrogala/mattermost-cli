package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/config"
	"github.com/spf13/cobra"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

var (
	loginFirefox       bool
	loginMattermostURL string
	loginTLSSkip       bool
)

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().BoolVar(&loginFirefox, "firefox", false, "extract session token from Firefox cookies")
	loginCmd.Flags().StringVar(&loginMattermostURL, "mattermost-url", "", "Mattermost instance URL")
	loginCmd.Flags().BoolVar(&loginTLSSkip, "tls-skip-verify", false, "skip TLS certificate verification")
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Mattermost. Use --firefox to extract session from browser.",
	RunE: func(_ *cobra.Command, _ []string) error {
		if loginFirefox {
			if loginMattermostURL == "" {
				return fmt.Errorf("--mattermost-url is required with --firefox")
			}
			return loginFromFirefox(loginMattermostURL)
		}
		return loginManual()
	},
}

func loginManual() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Mattermost URL (e.g. https://mattermost.example.com): ")
	urlInput, _ := reader.ReadString('\n')
	urlInput = strings.TrimSpace(urlInput)
	if urlInput == "" {
		return fmt.Errorf("URL is required")
	}

	fmt.Print("Session token: ")
	tokenInput, _ := reader.ReadString('\n')
	tokenInput = strings.TrimSpace(tokenInput)
	if tokenInput == "" {
		return fmt.Errorf("token is required")
	}

	return saveAndVerify(urlInput, tokenInput)
}

func loginFromFirefox(mmURL string) error {
	mmURL = strings.TrimRight(mmURL, "/")
	host := mmURL
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}

	profiles, err := firefoxCookiePaths()
	if err != nil {
		return err
	}

	for _, dbPath := range profiles {
		token, err := extractCookie(dbPath, host)
		if err != nil {
			continue
		}
		if token != "" {
			fmt.Printf("Found session token for %s\n", host)
			return saveAndVerify(mmURL, token)
		}
	}

	return fmt.Errorf("no MMAUTHTOKEN cookie found for %s. Are you logged in via Firefox?", host)
}

// firefoxCookiePaths returns cookie database paths for all Firefox profiles on the current OS.
func firefoxCookiePaths() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find home dir: %w", err)
	}

	var searchDirs []string
	switch runtime.GOOS {
	case "linux":
		searchDirs = []string{
			filepath.Join(homeDir, ".mozilla", "firefox"),
			filepath.Join(homeDir, "snap", "firefox", "common", ".mozilla", "firefox"),
		}
	case "darwin":
		searchDirs = []string{
			filepath.Join(homeDir, "Library", "Application Support", "Firefox", "Profiles"),
		}
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		searchDirs = []string{
			filepath.Join(appData, "Mozilla", "Firefox", "Profiles"),
		}
	default:
		return nil, fmt.Errorf("firefox cookie extraction not supported on %s", runtime.GOOS)
	}

	var profiles []string
	for _, dir := range searchDirs {
		matches, err := filepath.Glob(filepath.Join(dir, "*", "cookies.sqlite"))
		if err != nil {
			continue
		}
		profiles = append(profiles, matches...)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no Firefox cookie databases found (searched: %s)", strings.Join(searchDirs, ", "))
	}

	return profiles, nil
}

func extractCookie(dbPath, host string) (string, error) {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return "", err
	}
	tmpFile, err := os.CreateTemp("", "mm-cookies-*.sqlite")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return "", err
	}
	_ = tmpFile.Close()

	conn, err := sqlite.OpenConn(tmpPath, sqlite.OpenReadOnly)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	var token string
	err = sqlitex.Execute(conn,
		"SELECT value FROM moz_cookies WHERE name = 'MMAUTHTOKEN' AND host = ?",
		&sqlitex.ExecOptions{
			Args: []any{host},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				token = stmt.ColumnText(0)
				return nil
			},
		})
	if err != nil {
		return "", err
	}

	return token, nil
}

func saveAndVerify(mmURL, token string) error {
	cfg := &config.Config{
		URL:           mmURL,
		Token:         token,
		TLSSkipVerify: loginTLSSkip,
	}
	c := client.New(cfg)

	user, err := c.Me()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("Authenticated as %s (%s)\n", user.Username, user.Email)

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("cannot save config: %w", err)
	}

	dir, _ := config.ConfigDir()
	fmt.Printf("Config saved to %s/config.yaml\n", dir)
	return nil
}
