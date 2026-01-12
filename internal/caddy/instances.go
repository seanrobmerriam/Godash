package caddy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// InstanceStore provides file-based storage for Caddy instances
type InstanceStore struct {
	filePath  string
	mu        sync.RWMutex
	instances map[string]*CaddyInstance
}

// NewInstanceStore creates a new instance store
func NewInstanceStore(filePath string) (*InstanceStore, error) {
	store := &InstanceStore{
		filePath:  filePath,
		instances: make(map[string]*CaddyInstance),
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Load existing instances
	if err := store.load(); err != nil {
		// If file doesn't exist, that's ok - start with empty store
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load instances: %w", err)
		}
	}

	return store, nil
}

// load loads instances from the file
func (s *InstanceStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var instances []*CaddyInstance
	if err := json.Unmarshal(data, &instances); err != nil {
		return fmt.Errorf("failed to parse instances file: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances = make(map[string]*CaddyInstance)
	for _, inst := range instances {
		s.instances[inst.ID] = inst
	}

	return nil
}

// save saves instances to the file
func (s *InstanceStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instances := make([]*CaddyInstance, 0, len(s.instances))
	for _, inst := range s.instances {
		instances = append(instances, inst)
	}

	data, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal instances: %w", err)
	}

	// Write to temp file first, then rename for atomicity
	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmpPath, s.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// List returns all instances
func (s *InstanceStore) List() []*CaddyInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instances := make([]*CaddyInstance, 0, len(s.instances))
	for _, inst := range s.instances {
		instances = append(instances, inst)
	}
	return instances
}

// Get returns an instance by ID
func (s *InstanceStore) Get(id string) (*CaddyInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inst, ok := s.instances[id]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", id)
	}
	return inst, nil
}

// Create creates a new instance
func (s *InstanceStore) Create(req *InstanceRequest) (*CaddyInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate ID
	id := generateID()
	now := time.Now()

	inst := &CaddyInstance{
		ID:         id,
		Name:       req.Name,
		URL:        req.URL,
		APIKeyFile: req.APIKeyFile,
		Status:     StatusUnknown,
		Tags:       req.Tags,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	s.instances[id] = inst

	if err := s.save(); err != nil {
		delete(s.instances, id)
		return nil, err
	}

	return inst, nil
}

// Update updates an existing instance
func (s *InstanceStore) Update(id string, req *InstanceRequest) (*CaddyInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.instances[id]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", id)
	}

	inst.Name = req.Name
	inst.URL = req.URL
	inst.APIKeyFile = req.APIKeyFile
	inst.Tags = req.Tags
	inst.UpdatedAt = time.Now()

	if err := s.save(); err != nil {
		return nil, err
	}

	return inst, nil
}

// Delete deletes an instance
func (s *InstanceStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.instances[id]; !ok {
		return fmt.Errorf("instance not found: %s", id)
	}

	delete(s.instances, id)

	if err := s.save(); err != nil {
		return err
	}

	return nil
}

// UpdateStatus updates the status of an instance
func (s *InstanceStore) UpdateStatus(id string, status InstanceStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.instances[id]
	if !ok {
		return fmt.Errorf("instance not found: %s", id)
	}

	inst.Status = status
	inst.LastPing = time.Now()
	inst.UpdatedAt = time.Now()

	return s.save()
}

// GetByTag returns all instances with a specific tag
func (s *InstanceStore) GetByTag(tag string) []*CaddyInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*CaddyInstance
	for _, inst := range s.instances {
		for _, t := range inst.Tags {
			if t == tag {
				result = append(result, inst)
				break
			}
		}
	}
	return result
}

// generateID generates a unique ID for an instance
func generateID() string {
	return fmt.Sprintf("inst_%d", time.Now().UnixNano())
}

// InstanceService provides instance management operations
type InstanceService struct {
	store  *InstanceStore
	client *Client
}

// NewInstanceService creates a new instance service
func NewInstanceService(store *InstanceStore) *InstanceService {
	return &InstanceService{
		store: store,
	}
}

// List returns all instances
func (s *InstanceService) List() []*CaddyInstance {
	return s.store.List()
}

// Get returns an instance by ID
func (s *InstanceService) Get(id string) (*CaddyInstance, error) {
	return s.store.Get(id)
}

// Create creates a new instance
func (s *InstanceService) Create(req *InstanceRequest) (*CaddyInstance, error) {
	return s.store.Create(req)
}

// Update updates an instance
func (s *InstanceService) Update(id string, req *InstanceRequest) (*CaddyInstance, error) {
	return s.store.Update(id, req)
}

// Delete deletes an instance
func (s *InstanceService) Delete(id string) error {
	return s.store.Delete(id)
}

// TestConnection tests the connection to an instance
func (s *InstanceService) TestConnection(id string) error {
	inst, err := s.store.Get(id)
	if err != nil {
		return err
	}

	client, err := NewClientFromInstance(inst, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Ping(); err != nil {
		s.store.UpdateStatus(id, StatusOffline)
		return fmt.Errorf("connection failed: %w", err)
	}

	s.store.UpdateStatus(id, StatusOnline)
	return nil
}

// RefreshStatus refreshes the status of an instance
func (s *InstanceService) RefreshStatus(id string) error {
	inst, err := s.store.Get(id)
	if err != nil {
		return err
	}

	client, err := NewClientFromInstance(inst, 5*time.Second)
	if err != nil {
		s.store.UpdateStatus(id, StatusUnknown)
		return nil // Don't fail on refresh
	}

	if err := client.Ping(); err != nil {
		return s.store.UpdateStatus(id, StatusOffline)
	}

	return s.store.UpdateStatus(id, StatusOnline)
}

// RefreshAllStatuses refreshes the status of all instances
func (s *InstanceService) RefreshAllStatuses() {
	instances := s.store.List()
	for _, inst := range instances {
		go func(id string) {
			s.RefreshStatus(id)
		}(inst.ID)
	}
}

// GetMetrics returns metrics for an instance
func (s *InstanceService) GetMetrics(id string) (*InstanceMetrics, error) {
	inst, err := s.store.Get(id)
	if err != nil {
		return nil, err
	}

	client, err := NewClientFromInstance(inst, 10*time.Second)
	if err != nil {
		return nil, err
	}

	metricsText, err := client.GetMetrics()
	if err != nil {
		return nil, err
	}

	pm, err := ParsePrometheusMetrics(metricsText)
	if err != nil {
		return nil, err
	}

	metrics := &InstanceMetrics{
		InstanceID:  id,
		Timestamp:   time.Now(),
		NumRequests: int64(pm.RequestsTotal),
		StatusCodes: make(map[int]int64),
	}

	// Convert status codes
	for codeStr, count := range pm.RequestsByCode {
		var code int
		fmt.Sscanf(codeStr, "%d", &code)
		metrics.StatusCodes[code] = int64(count)
	}

	return metrics, nil
}
