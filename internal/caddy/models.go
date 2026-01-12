package caddy

import (
	"encoding/json"
	"os"
	"time"
)

// InstanceStatus represents the status of a Caddy instance
type InstanceStatus string

const (
	StatusOnline   InstanceStatus = "online"
	StatusOffline  InstanceStatus = "offline"
	StatusUnknown  InstanceStatus = "unknown"
	StatusUpdating InstanceStatus = "updating"
)

// CaddyInstance represents a managed Caddy server
type CaddyInstance struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	URL        string         `json:"url"`          // Admin API URL (e.g., http://localhost:2019)
	APIKeyFile string         `json:"api_key_file"` // Path to file containing API key
	Status     InstanceStatus `json:"status"`
	Tags       []string       `json:"tags,omitempty"` // Grouping tags
	LastPing   time.Time      `json:"last_ping,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// GetAPIKey loads the API key from the file specified in APIKeyFile
func (c *CaddyInstance) GetAPIKey() (string, error) {
	if c.APIKeyFile == "" {
		return "", nil
	}
	data, err := os.ReadFile(c.APIKeyFile)
	if err != nil {
		return "", err
	}
	return string(json.RawMessage(data)), nil
}

// IsOnline returns true if the instance is online
func (c *CaddyInstance) IsOnline() bool {
	return c.Status == StatusOnline
}

// InstanceMetrics represents metrics collected from a Caddy instance
type InstanceMetrics struct {
	InstanceID   string                 `json:"instance_id"`
	Timestamp    time.Time              `json:"timestamp"`
	Uptime       int64                  `json:"uptime"`       // Uptime in seconds
	NumRequests  int64                  `json:"num_requests"` // Total requests
	TotalTraffic int64                  `json:"total_bytes"`  // Total bytes served
	Sites        map[string]SiteMetrics `json:"sites,omitempty"`
	StatusCodes  map[int]int64          `json:"status_codes"` // HTTP status code counts
}

// SiteMetrics represents per-site metrics
type SiteMetrics struct {
	Name          string  `json:"name"`
	Requests      int64   `json:"requests"`
	BytesSent     int64   `json:"bytes_sent"`
	BytesReceived int64   `json:"bytes_received"`
	LatencyAvg    float64 `json:"latency_avg_ms"`
}

// PrometheusMetrics represents parsed Prometheus metrics from Caddy
type PrometheusMetrics struct {
	RequestsTotal    float64            `json:"requests_total"`
	ResponseSizes    map[string]float64 `json:"response_sizes"`
	RequestDurations map[string]float64 `json:"request_durations"`
	RequestsByCode   map[string]float64 `json:"requests_by_code"`
	RequestsByHost   map[string]float64 `json:"requests_by_host"`
}

// ServerInfo represents basic Caddy server information
type ServerInfo struct {
	Version string `json:"version"`
	Arch    string `json:"arch"`
	OS      string `json:"os"`
	NumCPU  int    `json:"num_cpu"`
}

// Site represents a configured site in Caddy
type Site struct {
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
	Listen []string    `json:"listen"`
}

// Config represents Caddy configuration
type Config struct {
	Admin   *AdminConfig   `json:"admin,omitempty"`
	Apps    map[string]any `json:"apps,omitempty"`
	Storage interface{}    `json:"storage,omitempty"`
}

// AdminConfig represents the admin API configuration
type AdminConfig struct {
	Listen string `json:"listen"`
}

// LogEntry represents a log entry from Caddy
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"msg"`
	Logger  string    `json:"logger,omitempty"`
}

// InstanceRequest represents a request to add/update a Caddy instance
type InstanceRequest struct {
	Name       string   `json:"name" binding:"required"`
	URL        string   `json:"url" binding:"required"`
	APIKeyFile string   `json:"api_key_file"`
	Tags       []string `json:"tags,omitempty"`
}

// InstanceResponse represents the API response for an instance
type InstanceResponse struct {
	Instance *CaddyInstance   `json:"instance"`
	Metrics  *InstanceMetrics `json:"metrics,omitempty"`
	Error    string           `json:"error,omitempty"`
}

// InstancesListResponse represents the API response for listing instances
type InstancesListResponse struct {
	Instances []*CaddyInstance `json:"instances"`
	Total     int              `json:"total"`
}

// AnalyticsResponse represents the API response for analytics data
type AnalyticsResponse struct {
	Metrics    *InstanceMetrics   `json:"metrics,omitempty"`
	History    []*InstanceMetrics `json:"history,omitempty"`
	TotalReqs  int64              `json:"total_requests"`
	TotalBytes int64              `json:"total_bytes"`
	Error      string             `json:"error,omitempty"`
}
