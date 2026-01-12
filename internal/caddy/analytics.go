package caddy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AnalyticsStore provides file-based storage for analytics data
type AnalyticsStore struct {
	metricsDir string
	mu         sync.RWMutex
}

// NewAnalyticsStore creates a new analytics store
func NewAnalyticsStore(metricsDir string) (*AnalyticsStore, error) {
	// Ensure directory exists
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metrics directory: %w", err)
	}

	return &AnalyticsStore{
		metricsDir: metricsDir,
	}, nil
}

// instanceDir returns the directory for an instance's metrics
func (s *AnalyticsStore) instanceDir(instanceID string) string {
	return filepath.Join(s.metricsDir, instanceID)
}

// SaveMetrics saves metrics for an instance
func (s *AnalyticsStore) SaveMetrics(instanceID string, metrics *InstanceMetrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create instance directory if needed
	instDir := s.instanceDir(instanceID)
	if err := os.MkdirAll(instDir, 0755); err != nil {
		return fmt.Errorf("failed to create instance directory: %w", err)
	}

	// Use timestamp as filename
	filename := fmt.Sprintf("%s.json", metrics.Timestamp.Format("2006-01-02T15:04:05Z07:00"))
	filePath := filepath.Join(instDir, filename)

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// GetMetrics returns metrics for an instance within a time range
func (s *AnalyticsStore) GetMetrics(instanceID string, start, end time.Time) ([]*InstanceMetrics, error) {
	instDir := s.instanceDir(instanceID)

	entries, err := os.ReadDir(instDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*InstanceMetrics{}, nil
		}
		return nil, fmt.Errorf("failed to read metrics directory: %w", err)
	}

	var metrics []*InstanceMetrics
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Parse timestamp from filename
		timestamp, err := time.Parse("2006-01-02T15:04:05Z07:00", entry.Name()[:len("2006-01-02T15:04:05Z07:00")])
		if err != nil {
			continue // Skip files with invalid names
		}

		// Check if within range
		if timestamp.Before(start) || timestamp.After(end) {
			continue
		}

		// Read and parse metrics
		data, err := os.ReadFile(filepath.Join(instDir, entry.Name()))
		if err != nil {
			continue
		}

		var m InstanceMetrics
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}

		metrics = append(metrics, &m)
	}

	return metrics, nil
}

// GetLatestMetrics returns the most recent metrics for an instance
func (s *AnalyticsStore) GetLatestMetrics(instanceID string) (*InstanceMetrics, error) {
	instDir := s.instanceDir(instanceID)

	entries, err := os.ReadDir(instDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read metrics directory: %w", err)
	}

	if len(entries) == 0 {
		return nil, nil
	}

	// Find the latest file
	var latestFile string
	var latestTime time.Time
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		timestamp, err := time.Parse("2006-01-02T15:04:05Z07:00", entry.Name()[:len("2006-01-02T15:04:05Z07:00")])
		if err != nil {
			continue
		}

		if timestamp.After(latestTime) {
			latestTime = timestamp
			latestFile = entry.Name()
		}
	}

	if latestFile == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filepath.Join(instDir, latestFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read metrics file: %w", err)
	}

	var metrics InstanceMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %w", err)
	}

	return &metrics, nil
}

// CleanupOldMetrics removes metrics older than the specified duration
func (s *AnalyticsStore) CleanupOldMetrics(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)

	entries, err := os.ReadDir(s.metricsDir)
	if err != nil {
		return fmt.Errorf("failed to read metrics directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		instDir := filepath.Join(s.metricsDir, entry.Name())
		instEntries, err := os.ReadDir(instDir)
		if err != nil {
			continue
		}

		for _, instEntry := range instEntries {
			if instEntry.IsDir() {
				continue
			}

			timestamp, err := time.Parse("2006-01-02T15:04:05Z07:00", instEntry.Name()[:len("2006-01-02T15:04:05Z07:00")])
			if err != nil {
				continue
			}

			if timestamp.Before(cutoff) {
				os.Remove(filepath.Join(instDir, instEntry.Name()))
			}
		}
	}

	return nil
}

// GetAggregatedMetrics returns aggregated metrics across all instances
func (s *AnalyticsStore) GetAggregatedMetrics(instances []*CaddyInstance, start, end time.Time) (*AnalyticsResponse, error) {
	var totalReqs int64
	var totalBytes int64
	var history []*InstanceMetrics

	for _, inst := range instances {
		metrics, err := s.GetMetrics(inst.ID, start, end)
		if err != nil {
			continue
		}

		for _, m := range metrics {
			totalReqs += m.NumRequests
			totalBytes += m.TotalTraffic
		}

		// Get latest metrics for each instance
		if latest, err := s.GetLatestMetrics(inst.ID); err == nil && latest != nil {
			history = append(history, latest)
		}
	}

	return &AnalyticsResponse{
		History:    history,
		TotalReqs:  totalReqs,
		TotalBytes: totalBytes,
	}, nil
}
