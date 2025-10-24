package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"service-short-link/internal/infrastructure/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type Scenario struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name    string            `yaml:"name"`
	Request StepRequest       `yaml:"request"`
	Extract map[string]string `yaml:"extract"`
	Expect  StepExpect        `yaml:"expect"`
}

type StepRequest struct {
	Method string                 `yaml:"method"`
	URL    string                 `yaml:"url"`
	Body   map[string]interface{} `yaml:"body"`
}

type StepExpect struct {
	Status   int               `yaml:"status"`
	Header   map[string]string `yaml:"header"`
	Contains []string          `yaml:"contains"`
}

func substituteVars(s string, vars map[string]string) string {
	now := time.Now().Format("150405") // HHMMSS format (6 characters)
	s = strings.ReplaceAll(s, "${NOW}", now)

	// Generate random short code if needed
	if strings.Contains(s, "${RANDOM}") {
		random := fmt.Sprintf("%d", time.Now().UnixNano()%100000)
		s = strings.ReplaceAll(s, "${RANDOM}", random)
	}

	for k, v := range vars {
		s = strings.ReplaceAll(s, "${"+k+"}", v)
	}
	return s
}

// extractSimpleJSONPath extracts values from JSON response using simple JSON path
func extractSimpleJSONPath(data []byte, path string) (string, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return "", err
	}

	if path == "$.data.short_url" {
		if d, ok := m["data"]; ok {
			if dm, ok := d.(map[string]interface{}); ok {
				if url, ok := dm["short_url"].(string); ok {
					// Extract just the short code from the URL
					parts := strings.Split(url, "/")
					if len(parts) > 0 {
						return parts[len(parts)-1], nil
					}
					return url, nil
				}
			}
		}
		return "", fmt.Errorf("cannot extract short_url from response data")
	}

	if path == "$.data.short_code" {
		if d, ok := m["data"]; ok {
			if dm, ok := d.(map[string]interface{}); ok {
				if shortCode, ok := dm["short_code"].(string); ok {
					return shortCode, nil
				}
			}
		}
		return "", fmt.Errorf("cannot extract short_code from response data")
	}

	return "", fmt.Errorf("unsupported extract path: %s", path)
}

func getTestDB(t *testing.T) *sql.DB {
	// Use config service to get test database config
	configService := config.NewConfigService()
	dbConfig := configService.GetTestDatabaseConfig()

	// If running from host (not Docker), use localhost with mapped port
	if dbConfig.Host == "mariadb_test" {
		// Check if we can resolve mariadb_test (Docker network) with timeout
		done := make(chan bool, 1)
		var err error

		go func() {
			_, err = net.LookupHost("mariadb_test")
			done <- true
		}()

		select {
		case <-done:
			if err != nil {
				// Running from host, use localhost with mapped port
				dbConfig.Host = "localhost"
				dbConfig.Port = 3316 // Mapped port from docker-compose
				t.Logf("Running from host, using localhost:3316 instead of mariadb_test:3306")
			}
		case <-time.After(1 * time.Second):
			// Timeout - assume running from host
			dbConfig.Host = "localhost"
			dbConfig.Port = 3316 // Mapped port from docker-compose
			t.Logf("DNS lookup timeout, assuming host mode, using localhost:3316")
		}
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=false&allowNativePasswords=true&multiStatements=false&interpolateParams=true&readTimeout=3s&writeTimeout=3s&timeout=3s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)

	t.Logf("Using test database: %s@%s:%d/%s",
		dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("cannot connect to test db: %v", err)
	}
	return db
}

func cleanupDB(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	// Use TRUNCATE for faster cleanup (faster than DELETE)
	queries := []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"TRUNCATE TABLE analytics",
		"TRUNCATE TABLE links",
		"SET FOREIGN_KEY_CHECKS = 1",
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			t.Logf("Warning: failed to execute %s: %v", query, err)
		}
	}
}

func seedLinks(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	var fixturesPath string
	if strings.HasSuffix(wd, "e2e") {
		fixturesPath = wd + "/fixtures/links_seed.json"
	} else {
		fixturesPath = wd + "/internal/tests/e2e/fixtures/links_seed.json"
	}

	b, err := os.ReadFile(fixturesPath)
	if err != nil {
		return
	}

	var links []map[string]interface{}
	_ = json.Unmarshal(b, &links)
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	for _, l := range links {
		b, _ := json.Marshal(l)
		req, _ := http.NewRequest("POST", baseURL+"/api/v1/links", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-api-key")
		// Add JWT token for authentication
		req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwbGF0Zm9ybSI6InNlcnZpY2UiLCJyb2xlIjoic3lzdGVtIiwiZXhwIjoxNzYxMjc4MTAxLCJpYXQiOjE3NjExOTE3MDF9.bquK5fd0_kM-Kx0yT8E3EnOAw_ssSyk1Db9vkjjry58")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Logf("Warning: Cannot seed link %v - server not running: %v", l, err)
			continue
		}
		resp.Body.Close()
	}
}

func TestAllScenariosE2E(t *testing.T) {
	t.Parallel() // Enable parallel execution

	setEnvDatabase(t)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	var scenariosPath string
	if strings.HasSuffix(wd, "e2e") {
		scenariosPath = wd + "/scenarios"
	} else {
		scenariosPath = wd + "/internal/tests/e2e/scenarios"
	}

	files, err := os.ReadDir(scenariosPath)
	assert.NoError(t, err)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".yaml") {
			continue
		}
		t.Run(f.Name(), func(t *testing.T) {
			cleanupDB(t)
			seedLinks(t)
			runSingleYamlScenario(t, scenariosPath+"/"+f.Name())
			cleanupDB(t)
		})
	}
}

func setEnvDatabase(t *testing.T) {
	os.Setenv("E2E_TEST_MODE", "true")
	t.Logf("E2E_TEST_MODE=true - Server will use test database automatically")
}

func runSingleYamlScenario(t *testing.T, file string) {
	b, err := os.ReadFile(file)
	assert.NoError(t, err)
	var scenario Scenario
	err = yaml.Unmarshal(b, &scenario)
	assert.NoError(t, err)

	cleanupDB(t)

	vars := map[string]string{}
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	t.Run("scenario", func(t *testing.T) {
		for idx, step := range scenario.Steps {
			t.Run(fmt.Sprintf("%d_%s", idx+1, step.Name), func(t *testing.T) {
				url := baseURL + substituteVars(step.Request.URL, vars)
				t.Logf("Making request to URL: %s", url)
				method := step.Request.Method
				var reqBody []byte
				if len(step.Request.Body) > 0 {
					for k, v := range step.Request.Body {
						if s, ok := v.(string); ok && strings.Contains(s, "${") {
							step.Request.Body[k] = substituteVars(s, vars)
						}
					}
					reqBody, _ = json.Marshal(step.Request.Body)
				}
				req, _ := http.NewRequest(method, url, bytes.NewReader(reqBody))
				if method == "POST" || method == "PUT" {
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("X-API-Key", "test-api-key")
				}
				// Add JWT token for authentication
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwbGF0Zm9ybSI6InNlcnZpY2UiLCJyb2xlIjoic3lzdGVtIiwiZXhwIjoxNzYxMjc4MTAxLCJpYXQiOjE3NjExOTE3MDF9.bquK5fd0_kM-Kx0yT8E3EnOAw_ssSyk1Db9vkjjry58")

				// Create client that doesn't follow redirects with timeout
				client := &http.Client{
					Timeout: 10 * time.Second, // Reduced timeout for VS Code debugger
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
				}
				resp, err := client.Do(req)
				if err != nil {
					t.Skipf("Step %s: Server not running (HTTP error: %v). Start server with 'go run cmd/api/main.go'", step.Name, err)
					return
				}
				defer resp.Body.Close()
				assert.Equal(t, step.Expect.Status, resp.StatusCode, "Step %s: status", step.Name)
				respBytes, _ := io.ReadAll(resp.Body)
				for h, v := range step.Expect.Header {
					assert.Equal(t, v, resp.Header.Get(h), "Step %s: header %s", step.Name, h)
				}
				for _, s := range step.Expect.Contains {
					assert.Contains(t, string(respBytes), s, "Step %s: expect body contains %s", step.Name, s)
				}
				// Extract variables if needed (for future use)
				if step.Extract != nil {
					for k, path := range step.Extract {
						extracted, err := extractSimpleJSONPath(respBytes, path)
						if err != nil {
							t.Logf("Failed to extract %s from path %s: %v", k, path, err)
							continue
						}
						vars[k] = extracted
						t.Logf("Extracted %s = %s from path %s", k, extracted, path)
					}
				}
			})
		}
	})
}
