package caddy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditAction represents an audit action type
type AuditAction string

const (
	ActionCreateInstance AuditAction = "create_instance"
	ActionUpdateInstance AuditAction = "update_instance"
	ActionDeleteInstance AuditAction = "delete_instance"
	ActionTestConnection AuditAction = "test_connection"
	ActionRefreshStatus  AuditAction = "refresh_status"
	ActionReloadConfig   AuditAction = "reload_config"
	ActionStopServer     AuditAction = "stop_server"
	ActionStartServer    AuditAction = "start_server"
	ActionRestartServer  AuditAction = "restart_server"
	ActionCreateSite     AuditAction = "create_site"
	ActionDeleteSite     AuditAction = "delete_site"
	ActionViewConfig     AuditAction = "view_config"
	ActionViewLogs       AuditAction = "view_logs"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID           string      `json:"id"`
	Timestamp    time.Time   `json:"timestamp"`
	UserID       int         `json:"user_id"`
	Username     string      `json:"username,omitempty"`
	InstanceID   string      `json:"instance_id,omitempty"`
	InstanceName string      `json:"instance_name,omitempty"`
	Action       AuditAction `json:"action"`
	Details      string      `json:"details,omitempty"`
	IPAddress    string      `json:"ip_address"`
	Success      bool        `json:"success"`
	ErrorMsg     string      `json:"error_msg,omitempty"`
}

// AuditStore provides file-based audit logging
type AuditStore struct {
	logFile    string
	mu         sync.Mutex
	maxEntries int // Maximum number of entries to keep
}

// NewAuditStore creates a new audit store
func NewAuditStore(logDir string, maxEntries int) (*AuditStore, error) {
	// Ensure directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	return &AuditStore{
		logFile:    filepath.Join(logDir, "audit.log"),
		maxEntries: maxEntries,
	}, nil
}

// Log records an audit entry
func (s *AuditStore) Log(entry *AuditEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry.ID = generateAuditID()
	entry.Timestamp = time.Now()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Append to log file
	f, err := os.OpenFile(s.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(string(data) + "\n"); err != nil {
		return err
	}

	// Rotate if needed
	return s.rotateIfNeeded()
}

// GetEntries returns audit entries, optionally filtered
func (s *AuditStore) GetEntries(filters map[string]string, limit int) ([]*AuditEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.readEntries()
	if err != nil {
		return nil, err
	}

	// Apply filters
	var filtered []*AuditEntry
	for _, entry := range entries {
		if s.matchesFilters(entry, filters) {
			filtered = append(filtered, entry)
		}
		if len(filtered) >= limit {
			break
		}
	}

	return filtered, nil
}

// GetEntriesForInstance returns all entries for a specific instance
func (s *AuditStore) GetEntriesForInstance(instanceID string, limit int) ([]*AuditEntry, error) {
	return s.GetEntries(map[string]string{"instance_id": instanceID}, limit)
}

// GetEntriesForUser returns all entries for a specific user
func (s *AuditStore) GetEntriesForUser(userID int, limit int) ([]*AuditEntry, error) {
	return s.GetEntries(map[string]string{"user_id": string(rune(userID))}, limit)
}

// GetRecentEntries returns the most recent entries
func (s *AuditStore) GetRecentEntries(limit int) ([]*AuditEntry, error) {
	return s.GetEntries(nil, limit)
}

// Clear clears the audit log
func (s *AuditStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return os.WriteFile(s.logFile, []byte{}, 0644)
}

func (s *AuditStore) readEntries() ([]*AuditEntry, error) {
	data, err := os.ReadFile(s.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*AuditEntry{}, nil
		}
		return nil, err
	}

	var entries []*AuditEntry
	lines := string(data)
	for _, line := range splitLines(lines) {
		if line == "" {
			continue
		}
		var entry AuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		entries = append(entries, &entry)
	}

	// Reverse to get newest first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, nil
}

func (s *AuditStore) matchesFilters(entry *AuditEntry, filters map[string]string) bool {
	for key, value := range filters {
		switch key {
		case "instance_id":
			if entry.InstanceID != value {
				return false
			}
		case "user_id":
			// Convert user_id to string for comparison
			if entry.UserID != 0 && string(rune(entry.UserID)) != value {
				return false
			}
		case "action":
			if string(entry.Action) != value {
				return false
			}
		case "success":
			// Parse boolean string
			return false
		}
	}
	return true
}

func (s *AuditStore) rotateIfNeeded() error {
	entries, err := s.readEntries()
	if err != nil {
		return err
	}

	if len(entries) <= s.maxEntries {
		return nil
	}

	// Keep only the most recent entries
	kept := entries[len(entries)-s.maxEntries:]

	// Rewrite file
	f, err := os.Create(s.logFile)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, entry := range kept {
		data, _ := json.Marshal(entry)
		f.WriteString(string(data) + "\n")
	}

	return nil
}

func generateAuditID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
