package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ConfigService provides configuration management operations
type ConfigService struct {
	instanceService *InstanceService
	metricsStore    *AnalyticsStore
}

// NewConfigService creates a new config service
func NewConfigService(instanceService *InstanceService, metricsStore *AnalyticsStore) *ConfigService {
	return &ConfigService{
		instanceService: instanceService,
		metricsStore:    metricsStore,
	}
}

// GetConfig retrieves the current configuration from an instance
func (s *ConfigService) GetConfig(instanceID string) (*Config, error) {
	inst, err := s.instanceService.Get(instanceID)
	if err != nil {
		return nil, err
	}

	client, err := NewClientFromInstance(inst, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client.GetConfig()
}

// ReloadConfig reloads configuration on an instance
func (s *ConfigService) ReloadConfig(instanceID string, configJSON []byte) error {
	inst, err := s.instanceService.Get(instanceID)
	if err != nil {
		return err
	}

	client, err := NewClientFromInstance(inst, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return client.ReloadConfig(configJSON)
}

// GetCaddyfile returns the configuration as a Caddyfile format
func (s *ConfigService) GetCaddyfile(instanceID string) (string, error) {
	config, err := s.GetConfig(instanceID)
	if err != nil {
		return "", err
	}

	// Convert JSON config to Caddyfile format
	// This is a simplified version - a full implementation would use Caddy's JSON-to-Caddyfile conversion
	return convertConfigToCaddyfile(config), nil
}

// UpdateCaddyfile parses and applies a Caddyfile
func (s *ConfigService) UpdateCaddyfile(instanceID string, caddyfile string) error {
	// Parse Caddyfile using Caddy's adapter
	// For now, we expect the Caddyfile to be converted to JSON externally
	// In a full implementation, we'd use caddy.Module.IDToPath etc.
	return fmt.Errorf("Caddyfile parsing not implemented - please provide JSON config")
}

// GetSites returns all sites configured on an instance
func (s *ConfigService) GetSites(instanceID string) ([]Site, error) {
	inst, err := s.instanceService.Get(instanceID)
	if err != nil {
		return nil, err
	}

	client, err := NewClientFromInstance(inst, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client.GetSites()
}

// CreateSite creates or updates a site
func (s *ConfigService) CreateSite(instanceID string, siteName string, config map[string]interface{}) error {
	inst, err := s.instanceService.Get(instanceID)
	if err != nil {
		return err
	}

	client, err := NewClientFromInstance(inst, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return client.CreateSite(siteName, config)
}

// DeleteSite removes a site
func (s *ConfigService) DeleteSite(instanceID string, siteName string) error {
	inst, err := s.instanceService.Get(instanceID)
	if err != nil {
		return err
	}

	client, err := NewClientFromInstance(inst, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return client.DeleteSite(siteName)
}

// GetLogs retrieves logs from an instance
func (s *ConfigService) GetLogs(instanceID string, tailLines int) ([]LogEntry, error) {
	inst, err := s.instanceService.Get(instanceID)
	if err != nil {
		return nil, err
	}

	client, err := NewClientFromInstance(inst, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client.GetLogs(tailLines)
}

// CollectMetrics collects and stores metrics from an instance
func (s *ConfigService) CollectMetrics(instanceID string) (*InstanceMetrics, error) {
	inst, err := s.instanceService.Get(instanceID)
	if err != nil {
		return nil, err
	}

	client, err := NewClientFromInstance(inst, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Get Prometheus metrics
	metricsText, err := client.GetMetrics()
	if err != nil {
		return nil, err
	}

	pm, err := ParsePrometheusMetrics(metricsText)
	if err != nil {
		return nil, err
	}

	metrics := &InstanceMetrics{
		InstanceID:   instanceID,
		Timestamp:    time.Now(),
		Uptime:       0, // Would need to calculate from server info
		NumRequests:  int64(pm.RequestsTotal),
		TotalTraffic: 0, // Would need to calculate from response size metrics
		StatusCodes:  make(map[int]int64),
	}

	// Convert status codes
	for codeStr, count := range pm.RequestsByCode {
		var code int
		fmt.Sscanf(codeStr, "%d", &code)
		metrics.StatusCodes[code] = int64(count)
	}

	// Store metrics if store is available
	if s.metricsStore != nil {
		s.metricsStore.SaveMetrics(instanceID, metrics)
	}

	return metrics, nil
}

// convertConfigToCaddyfile converts JSON config to Caddyfile format
func convertConfigToCaddyfile(config *Config) string {
	var buf bytes.Buffer

	// Add global options
	buf.WriteString("{\n")
	buf.WriteString("    admin off\n")
	if config.Admin != nil {
		buf.WriteString(fmt.Sprintf("    admin %s\n", config.Admin.Listen))
	}
	buf.WriteString("}\n\n")

	// This is a simplified conversion - a full implementation would properly
	// convert all JSON config options to their Caddyfile equivalents
	buf.WriteString("# Configuration converted from JSON\n")
	buf.WriteString("# See https://caddyserver.com/docs/ for full documentation\n")

	return buf.String()
}

// StopServer stops a Caddy server
func (s *ConfigService) StopServer(instanceID string) error {
	inst, err := s.instanceService.Get(instanceID)
	if err != nil {
		return err
	}

	client, err := NewClientFromInstance(inst, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return client.Stop()
}

// HealthCheck performs a health check on an instance
func (s *ConfigService) HealthCheck(instanceID string) (bool, string, error) {
	inst, err := s.instanceService.Get(instanceID)
	if err != nil {
		return false, "", err
	}

	client, err := NewClientFromInstance(inst, 5*time.Second)
	if err != nil {
		return false, fmt.Sprintf("Failed to create client: %v", err), nil
	}

	if err := client.Ping(); err != nil {
		return false, fmt.Sprintf("Ping failed: %v", err), nil
	}

	// Get server info for more details
	info, err := client.GetServerInfo()
	if err != nil {
		return true, "Reachable", nil
	}

	return true, fmt.Sprintf("Version: %s", info.Version), nil
}

// ExportConfig exports the configuration as JSON
func (s *ConfigService) ExportConfig(instanceID string) (string, error) {
	config, err := s.GetConfig(instanceID)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	return string(data), nil
}

// ImportConfig imports configuration from JSON
func (s *ConfigService) ImportConfig(instanceID string, configJSON io.Reader) error {
	data, err := io.ReadAll(configJSON)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	return s.ReloadConfig(instanceID, data)
}
