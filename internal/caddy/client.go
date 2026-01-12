package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client provides methods to interact with a Caddy server's admin API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	timeout    time.Duration
}

// NewClient creates a new Caddy client
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: timeout},
		timeout:    timeout,
	}
}

// NewClientFromInstance creates a new Caddy client from an instance
func NewClientFromInstance(instance *CaddyInstance, timeout time.Duration) (*Client, error) {
	apiKey, err := instance.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load API key: %w", err)
	}
	return NewClient(instance.URL, apiKey, timeout), nil
}

// Ping checks if the Caddy instance is reachable
func (c *Client) Ping() error {
	req, err := http.NewRequest("GET", c.baseURL+"/id", nil)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// GetServerInfo returns basic server information
func (c *Client) GetServerInfo() (*ServerInfo, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/id", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get server info: status %d", resp.StatusCode)
	}

	// Caddy returns the admin ID, not full info
	// For more info, we'd need to parse config
	info := &ServerInfo{
		Version: "unknown", // Would need to parse from /buildinfo or similar
	}

	return info, nil
}

// GetConfig returns the current Caddy configuration
func (c *Client) GetConfig() (*Config, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/config/", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get config: %s", string(body))
	}

	var config Config
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// ReloadConfig reloads the Caddy configuration
func (c *Client) ReloadConfig(configJSON []byte) error {
	req, err := http.NewRequest("POST", c.baseURL+"/load", bytes.NewReader(configJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Caddy returns 200 on success, or an error
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("config reload failed: %s", string(body))
	}

	return nil
}

// Stop stops the Caddy server
func (c *Client) Stop() error {
	req, err := http.NewRequest("POST", c.baseURL+"/stop", nil)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stop failed: %s", string(body))
	}

	return nil
}

// GetMetrics retrieves Prometheus metrics from Caddy
func (c *Client) GetMetrics() (string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/metrics", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get metrics: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetLogs retrieves recent logs from Caddy
func (c *Client) GetLogs(tailLines int) ([]LogEntry, error) {
	// For now, return empty - proper implementation would use WebSocket
	// Caddy's /admin/log endpoint supports WebSocket for real-time logs
	return []LogEntry{}, nil
}

// GetSites returns the list of configured sites
func (c *Client) GetSites() ([]Site, error) {
	config, err := c.GetConfig()
	if err != nil {
		return nil, err
	}

	var sites []Site

	// Parse apps to find HTTP app and its sites
	if httpApp, ok := config.Apps["http"].(map[string]interface{}); ok {
		if servers, ok := httpApp["servers"].(map[string]interface{}); ok {
			for name, srv := range servers {
				srvMap := srv.(map[string]interface{})
				site := Site{
					Name:   name,
					Config: srv,
				}
				if listen, ok := srvMap["listen"].([]interface{}); ok {
					for _, l := range listen {
						site.Listen = append(site.Listen, fmt.Sprintf("%v", l))
					}
				}
				sites = append(sites, site)
			}
		}
	}

	return sites, nil
}

// CreateSite creates or updates a site configuration
func (c *Client) CreateSite(name string, config map[string]interface{}) error {
	url := fmt.Sprintf("%s/config/apps/http/servers/%s", c.baseURL, name)

	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(configJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create site: %s", string(body))
	}

	return nil
}

// DeleteSite removes a site configuration
func (c *Client) DeleteSite(name string) error {
	url := fmt.Sprintf("%s/config/apps/http/servers/%s", c.baseURL, name)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete site: %s", string(body))
	}

	return nil
}

// doRequest performs an HTTP request with proper headers and error handling
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	// Set default headers
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	// Add API key authorization if set
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Perform the request with timeout
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// ParsePrometheusMetrics parses Prometheus metrics format into structured data
func ParsePrometheusMetrics(metricsText string) (*PrometheusMetrics, error) {
	pm := &PrometheusMetrics{
		ResponseSizes:    make(map[string]float64),
		RequestDurations: make(map[string]float64),
		RequestsByCode:   make(map[string]float64),
		RequestsByHost:   make(map[string]float64),
	}

	lines := strings.Split(metricsText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse metric line: metric_name{labels} value
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		metricName := parts[0]
		valueStr := parts[1]

		var value float64
		if _, err := fmt.Sscanf(valueStr, "%f", &value); err != nil {
			continue
		}

		switch {
		case strings.HasSuffix(metricName, "_total") && strings.Contains(metricName, "requests"):
			pm.RequestsTotal = value
			// Try to extract labels
			if strings.Contains(metricName, "code=") {
				code := extractLabel(metricName, "code")
				pm.RequestsByCode[code] = value
			}
			if strings.Contains(metricName, "host=") {
				host := extractLabel(metricName, "host")
				pm.RequestsByHost[host] = value
			}
		case strings.Contains(metricName, "response_size"):
			pm.ResponseSizes["total"] = value
		case strings.Contains(metricName, "request_duration"):
			pm.RequestDurations["total"] = value
		}
	}

	return pm, nil
}

// extractLabel extracts a label value from a metric name
func extractLabel(metricName, label string) string {
	pattern := label + "=\""
	start := strings.Index(metricName, pattern)
	if start == -1 {
		return ""
	}
	start += len(pattern)
	end := strings.Index(metricName[start:], "\"")
	if end == -1 {
		return ""
	}
	return metricName[start : start+end]
}
