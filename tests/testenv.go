package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	networkName   = "mm-test-net"
	pgName        = "mm-test-postgres"
	mmName        = "mm-test-mattermost"
	idleTimeout   = 10 * time.Minute
	killerPidFile = "/tmp/mm-test-killer.pid"
)

// TestEnvironment holds containers and credentials for integration tests.
type TestEnvironment struct {
	ctx         context.Context
	pgContainer testcontainers.Container
	mmContainer testcontainers.Container
	reused      bool

	MattermostURL string // http://localhost:<port>
	AdminToken    string // personal access token
	AdminUserID   string
	AdminUsername  string
	TeamID        string

	Users      map[string]*TestUser // username -> user info
	httpClient *http.Client
}

// TestUser holds info about a test user.
type TestUser struct {
	ID       string
	Username string
	Email    string
	Password string
	Token    string // session token
}

// NewTestEnvironment starts or reuses postgres + mattermost containers and sets up test data.
func NewTestEnvironment(ctx context.Context) (*TestEnvironment, error) {
	cancelIdleKiller()

	env := &TestEnvironment{
		ctx:        ctx,
		Users:      make(map[string]*TestUser),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	if err := env.startContainers(); err != nil {
		return nil, fmt.Errorf("start containers: %w", err)
	}

	if err := env.setupMattermost(); err != nil {
		return nil, fmt.Errorf("setup mattermost: %w", err)
	}

	return env, nil
}

func (env *TestEnvironment) startContainers() error {
	// Ensure network exists
	ensureNetwork(env.ctx)

	var err error

	// PostgreSQL (reusable)
	env.pgContainer, err = testcontainers.GenericContainer(env.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:         pgName,
			Image:        "postgres:15",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       "mattermost",
				"POSTGRES_USER":     "mmuser",
				"POSTGRES_PASSWORD": "mmtest",
			},
			Networks:       []string{networkName},
			NetworkAliases: map[string][]string{networkName: {"postgres"}},
			WaitingFor:     wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
		},
		Started: true,
		Reuse:   true,
	})
	if err != nil {
		return fmt.Errorf("start postgres: %w", err)
	}

	// Mattermost (reusable)
	env.mmContainer, err = testcontainers.GenericContainer(env.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:         mmName,
			Image:        "mattermost/mattermost-enterprise-edition:9.11",
			ExposedPorts: []string{"8065/tcp"},
			Env: map[string]string{
				"MM_SQLSETTINGS_DRIVERNAME":                 "postgres",
				"MM_SQLSETTINGS_DATASOURCE":                "postgres://mmuser:mmtest@postgres:5432/mattermost?sslmode=disable",
				"MM_SERVICESETTINGS_SITEURL":               "http://localhost:8065",
				"MM_SERVICESETTINGS_ENABLETESTING":          "true",
				"MM_SERVICESETTINGS_ENABLEDEVELOPER":        "true",
				"MM_SERVICESETTINGS_ENABLEUSERACCESSTOKENS": "true",
				"MM_TEAMSETTINGS_ENABLEOPENSERVER":          "true",
				"MM_LOGSETTINGS_CONSOLELEVEL":               "ERROR",
				"MM_PASSWORDSETTINGS_MINIMUMLENGTH":         "5",
				"MM_SERVICESETTINGS_ENABLELOCALMODE":        "false",
			},
			Networks:       []string{networkName},
			NetworkAliases: map[string][]string{networkName: {"mattermost"}},
			WaitingFor:     wait.ForHTTP("/api/v4/system/ping").WithPort("8065/tcp").WithStartupTimeout(120 * time.Second),
		},
		Started: true,
		Reuse:   true,
	})
	if err != nil {
		return fmt.Errorf("start mattermost: %w", err)
	}

	host, err := env.mmContainer.Host(env.ctx)
	if err != nil {
		return err
	}
	port, err := env.mmContainer.MappedPort(env.ctx, "8065/tcp")
	if err != nil {
		return err
	}
	env.MattermostURL = fmt.Sprintf("http://%s:%s", host, port.Port())

	return nil
}

func (env *TestEnvironment) setupMattermost() error {
	// Try to login first — if it works, the instance was already set up (reused)
	if token, err := env.login("testadmin", "Admin1!"); err == nil {
		env.reused = true
		return env.reconnect(token)
	}

	// Fresh instance: create everything
	return env.freshSetup()
}

// reconnect re-establishes credentials on a reused instance.
func (env *TestEnvironment) reconnect(sessionToken string) error {
	// Get admin user info
	data, _, err := env.apiCall("GET", "/users/me", nil, sessionToken)
	if err != nil {
		return fmt.Errorf("get admin user: %w", err)
	}
	var adminUser struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(data, &adminUser); err != nil {
		return err
	}
	env.AdminUserID = adminUser.ID
	env.AdminUsername = adminUser.Username

	// Create a fresh personal access token
	pat, err := env.createPersonalAccessToken(sessionToken, adminUser.ID)
	if err != nil {
		return fmt.Errorf("create PAT: %w", err)
	}
	env.AdminToken = pat

	// Re-login test users
	for _, u := range []struct{ name, email, pass string }{
		{"alice", "alice@test.local", "Alice1!"},
		{"bob", "bob@test.local", "Bob12!"},
		{"testbot", "testbot@test.local", "Bot123!"},
	} {
		token, err := env.login(u.name, u.pass)
		if err != nil {
			return fmt.Errorf("login %s: %w", u.name, err)
		}
		// Get user ID
		userData, _, err := env.apiCall("GET", "/users/me", nil, token)
		if err != nil {
			return fmt.Errorf("get user %s: %w", u.name, err)
		}
		var userInfo struct{ ID string `json:"id"` }
		if err := json.Unmarshal(userData, &userInfo); err != nil {
			return err
		}
		env.Users[u.name] = &TestUser{
			ID:       userInfo.ID,
			Username: u.name,
			Email:    u.email,
			Password: u.pass,
			Token:    token,
		}
	}

	// Get team ID
	data, _, err = env.apiCall("GET", fmt.Sprintf("/teams/name/%s", "testteam"), nil, env.AdminToken)
	if err != nil {
		return fmt.Errorf("get team: %w", err)
	}
	var team struct{ ID string `json:"id"` }
	if err := json.Unmarshal(data, &team); err != nil {
		return err
	}
	env.TeamID = team.ID

	return nil
}

func (env *TestEnvironment) freshSetup() error {
	admin, err := env.createUser("testadmin", "admin@test.local", "Admin1!")
	if err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	env.AdminUserID = admin.ID
	env.AdminUsername = admin.Username

	sessionToken, err := env.login("testadmin", "Admin1!")
	if err != nil {
		return fmt.Errorf("login admin: %w", err)
	}

	pat, err := env.createPersonalAccessToken(sessionToken, admin.ID)
	if err != nil {
		return fmt.Errorf("create PAT: %w", err)
	}
	env.AdminToken = pat

	for _, u := range []struct{ name, email, pass string }{
		{"alice", "alice@test.local", "Alice1!"},
		{"bob", "bob@test.local", "Bob12!"},
		{"testbot", "testbot@test.local", "Bot123!"},
	} {
		user, err := env.createUser(u.name, u.email, u.pass)
		if err != nil {
			return fmt.Errorf("create user %s: %w", u.name, err)
		}
		token, err := env.login(u.name, u.pass)
		if err != nil {
			return fmt.Errorf("login %s: %w", u.name, err)
		}
		user.Token = token
		env.Users[u.name] = user
	}

	teamID, err := env.createTeam(env.AdminToken)
	if err != nil {
		return fmt.Errorf("create team: %w", err)
	}
	env.TeamID = teamID

	for _, u := range env.Users {
		if err := env.addToTeam(env.AdminToken, teamID, u.ID); err != nil {
			return fmt.Errorf("add %s to team: %w", u.Username, err)
		}
	}

	return nil
}

// ScheduleIdleShutdown spawns a background process that stops the containers
// after idleTimeout of inactivity. Canceled on next test run.
func (env *TestEnvironment) ScheduleIdleShutdown() {
	secs := int(idleTimeout.Seconds())
	script := fmt.Sprintf("sleep %d && docker rm -f %s %s 2>/dev/null; docker network rm %s 2>/dev/null",
		secs, mmName, pgName, networkName)

	cmd := exec.Command("bash", "-c", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // detach from parent
	if err := cmd.Start(); err != nil {
		return
	}

	// Write PID so next run can cancel it
	_ = os.WriteFile(killerPidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0600)

	// Don't wait — let it run in background
	go func() { _ = cmd.Wait() }()
}

// cancelIdleKiller kills any previously scheduled idle shutdown.
func cancelIdleKiller() {
	data, err := os.ReadFile(killerPidFile)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return
	}
	// Kill the process group (the sleep + docker stop chain)
	_ = syscall.Kill(-pid, syscall.SIGKILL)
	_ = os.Remove(killerPidFile)
}

// ensureNetwork creates the Docker network if it doesn't exist.
func ensureNetwork(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "docker", "network", "create", networkName)
	_ = cmd.Run() // ignore "already exists" error
}

// --- Admin API helpers ---

func (env *TestEnvironment) apiCall(method, path string, body any, authToken string) ([]byte, http.Header, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(env.ctx, method, env.MattermostURL+"/api/v4"+path, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := env.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("API %s %s: %d %s", method, path, resp.StatusCode, string(respBody))
	}

	return respBody, resp.Header, nil
}

func (env *TestEnvironment) createUser(username, email, password string) (*TestUser, error) {
	data, _, err := env.apiCall("POST", "/users", map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}, "")
	if err != nil {
		return nil, err
	}
	var resp struct{ ID string `json:"id"` }
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &TestUser{
		ID:       resp.ID,
		Username: username,
		Email:    email,
		Password: password,
	}, nil
}

func (env *TestEnvironment) login(username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"login_id": username,
		"password": password,
	})

	req, err := http.NewRequestWithContext(env.ctx, "POST", env.MattermostURL+"/api/v4/users/login", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := env.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed: %d %s", resp.StatusCode, string(respBody))
	}

	return resp.Header.Get("Token"), nil
}

func (env *TestEnvironment) createPersonalAccessToken(sessionToken, userID string) (string, error) {
	data, _, err := env.apiCall("POST", fmt.Sprintf("/users/%s/tokens", userID), map[string]string{
		"description": "integration-test",
	}, sessionToken)
	if err != nil {
		return "", err
	}
	var resp struct{ Token string `json:"token"` }
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.Token, nil
}

func (env *TestEnvironment) createTeam(token string) (string, error) {
	data, _, err := env.apiCall("POST", "/teams", map[string]string{
		"name":         "testteam",
		"display_name": "Test Team",
		"type":         "O",
	}, token)
	if err != nil {
		return "", err
	}
	var resp struct{ ID string `json:"id"` }
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (env *TestEnvironment) addToTeam(token, teamID, userID string) error {
	_, _, err := env.apiCall("POST", fmt.Sprintf("/teams/%s/members", teamID), map[string]string{
		"team_id": teamID,
		"user_id": userID,
	}, token)
	return err
}

// GetChannelByName looks up a channel by its slug name. Returns ID or empty string.
func (env *TestEnvironment) GetChannelByName(name string) (string, error) {
	data, _, err := env.apiCall("GET",
		fmt.Sprintf("/teams/%s/channels/name/%s", env.TeamID, name),
		nil, env.AdminToken)
	if err != nil {
		return "", err
	}
	var resp struct{ ID string `json:"id"` }
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

// CreateChannel creates a channel in the test team and adds all test users.
func (env *TestEnvironment) CreateChannel(name, displayName, chanType string) (string, error) {
	data, _, err := env.apiCall("POST", "/channels", map[string]any{
		"team_id":      env.TeamID,
		"name":         name,
		"display_name": displayName,
		"type":         chanType,
	}, env.AdminToken)
	if err != nil {
		return "", err
	}
	var resp struct{ ID string `json:"id"` }
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}

	// Add all test users so they can post
	for _, u := range env.Users {
		_ = env.AddUserToChannel(resp.ID, u.ID)
	}

	return resp.ID, nil
}

// AddUserToChannel adds a user to a channel.
func (env *TestEnvironment) AddUserToChannel(channelID, userID string) error {
	_, _, err := env.apiCall("POST", fmt.Sprintf("/channels/%s/members", channelID), map[string]string{
		"user_id": userID,
	}, env.AdminToken)
	return err
}

// PostMessage posts a message to a channel as a specific user.
func (env *TestEnvironment) PostMessage(token, channelID, message string) (string, error) {
	data, _, err := env.apiCall("POST", "/posts", map[string]string{
		"channel_id": channelID,
		"message":    message,
	}, token)
	if err != nil {
		return "", err
	}
	var resp struct{ ID string `json:"id"` }
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

// MuteChannel mutes a channel for the admin user.
func (env *TestEnvironment) MuteChannel(channelID string) error {
	_, _, err := env.apiCall("PUT",
		fmt.Sprintf("/channels/%s/members/%s/notify_props", channelID, env.AdminUserID),
		map[string]string{"mark_unread": "mention"},
		env.AdminToken)
	return err
}

// SetPostTimestamp modifies a post's create_at via the postgres container.
func (env *TestEnvironment) SetPostTimestamp(postID string, t time.Time) error {
	ms := t.UnixMilli()
	cmd := []string{
		"psql", "-U", "mmuser", "-d", "mattermost", "-c",
		fmt.Sprintf("UPDATE posts SET createat = %d, updateat = %d WHERE id = '%s'", ms, ms, postID),
	}
	_, _, err := env.pgContainer.Exec(env.ctx, cmd)
	return err
}

// ViewChannel marks a channel as viewed for the admin user (resets unread count).
func (env *TestEnvironment) ViewChannel(channelID string) error {
	_, _, err := env.apiCall("POST", "/channels/members/me/view", map[string]any{
		"channel_id": channelID,
	}, env.AdminToken)
	return err
}

// DeletePost deletes a post by ID.
func (env *TestEnvironment) DeletePost(token, postID string) error {
	_, _, err := env.apiCall("DELETE", fmt.Sprintf("/posts/%s", postID), nil, token)
	return err
}

// pidFilePath returns the path to the idle killer PID file.
func pidFilePath() string {
	return filepath.Join(os.TempDir(), "mm-test-killer.pid")
}
