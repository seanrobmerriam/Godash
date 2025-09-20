package services

import (
	"godash/internal/models"
	"math/rand"
	"runtime"
	"time"
)

// DashboardService handles dashboard-related business logic
type DashboardService struct {
}

// NewDashboardService creates a new dashboard service
func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

// GetDashboardData returns the current dashboard data
func (s *DashboardService) GetDashboardData() *models.DashboardData {
	now := time.Now()
	
	return &models.DashboardData{
		Title:       "Admin Dashboard",
		RefreshRate: 30, // 30 seconds
		LastUpdate:  now,
		Stats:       s.getSystemStats(),
		Widgets:     s.getDefaultWidgets(),
	}
}

// getSystemStats returns current system statistics
func (s *DashboardService) getSystemStats() models.SystemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return models.SystemStats{
		CPUUsage:    s.getCPUUsage(),
		MemoryUsage: float64(m.Sys) / 1024 / 1024, // Convert to MB
		DiskUsage:   s.getDiskUsage(),
		Uptime:      time.Now().Unix(), // Simplified uptime
		LastUpdated: time.Now(),
	}
}

// getCPUUsage returns simulated CPU usage (replace with actual implementation)
func (s *DashboardService) getCPUUsage() float64 {
	// Simulate CPU usage between 10-80%
	return 10 + rand.Float64()*70
}

// getDiskUsage returns simulated disk usage (replace with actual implementation)
func (s *DashboardService) getDiskUsage() float64 {
	// Simulate disk usage between 20-90%
	return 20 + rand.Float64()*70
}

// getDefaultWidgets returns the default set of dashboard widgets
func (s *DashboardService) getDefaultWidgets() []models.Widget {
	return []models.Widget{
		{
			ID:    "cpu-chart",
			Type:  models.WidgetTypeChart,
			Title: "CPU Usage",
			Position: models.Position{X: 0, Y: 0},
			Size:     models.Size{Width: 6, Height: 4},
			Data:     s.getCPUChartData(),
		},
		{
			ID:    "memory-metric",
			Type:  models.WidgetTypeMetric,
			Title: "Memory Usage",
			Position: models.Position{X: 6, Y: 0},
			Size:     models.Size{Width: 3, Height: 2},
			Data:     s.getMemoryMetricData(),
		},
		{
			ID:    "disk-progress",
			Type:  models.WidgetTypeProgress,
			Title: "Disk Space",
			Position: models.Position{X: 9, Y: 0},
			Size:     models.Size{Width: 3, Height: 2},
			Data:     s.getDiskProgressData(),
		},
		{
			ID:    "recent-activity",
			Type:  models.WidgetTypeActivity,
			Title: "Recent Activity",
			Position: models.Position{X: 6, Y: 2},
			Size:     models.Size{Width: 6, Height: 4},
			Data:     s.getActivityData(),
		},
		{
			ID:    "users-table",
			Type:  models.WidgetTypeTable,
			Title: "Recent Users",
			Position: models.Position{X: 0, Y: 4},
			Size:     models.Size{Width: 6, Height: 4},
			Data:     s.getUsersTableData(),
		},
		{
			ID:    "system-info",
			Type:  models.WidgetTypeText,
			Title: "System Information",
			Position: models.Position{X: 6, Y: 6},
			Size:     models.Size{Width: 6, Height: 2},
			Data:     s.getSystemInfoData(),
		},
	}
}

// getCPUChartData returns sample CPU chart data
func (s *DashboardService) getCPUChartData() models.ChartData {
	labels := make([]string, 12)
	data := make([]float64, 12)
	
	now := time.Now()
	for i := 0; i < 12; i++ {
		labels[11-i] = now.Add(-time.Duration(i)*5*time.Minute).Format("15:04")
		data[11-i] = 20 + rand.Float64()*60
	}
	
	return models.ChartData{
		Labels: labels,
		Datasets: []models.Dataset{
			{
				Label:           "CPU Usage %",
				Data:            data,
				BorderColor:     "#3b82f6",
				BackgroundColor: "rgba(59, 130, 246, 0.1)",
			},
		},
	}
}

// getMemoryMetricData returns memory usage metric data
func (s *DashboardService) getMemoryMetricData() models.MetricData {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	usageMB := float64(m.Sys) / 1024 / 1024
	
	return models.MetricData{
		Value:       int(usageMB),
		Unit:        "MB",
		Description: "System Memory",
		Trend:       "stable",
		Change:      2.1,
	}
}

// getDiskProgressData returns disk usage progress data
func (s *DashboardService) getDiskProgressData() models.ProgressData {
	usage := 20 + rand.Float64()*60
	
	return models.ProgressData{
		Value:       usage,
		Max:         100,
		Label:       "Disk Usage",
		Description: "Available storage space",
	}
}

// getActivityData returns sample activity data
func (s *DashboardService) getActivityData() models.ActivityData {
	now := time.Now()
	
	items := []models.ActivityItem{
		{
			ID:          "1",
			Title:       "User login",
			Description: "admin logged into the system",
			Type:        "success",
			Timestamp:   now.Add(-10 * time.Minute),
			User:        "admin",
		},
		{
			ID:          "2",
			Title:       "System backup completed",
			Description: "Daily backup finished successfully",
			Type:        "info",
			Timestamp:   now.Add(-30 * time.Minute),
		},
		{
			ID:          "3",
			Title:       "High memory usage detected",
			Description: "Memory usage exceeded 85%",
			Type:        "warning",
			Timestamp:   now.Add(-1 * time.Hour),
		},
		{
			ID:          "4",
			Title:       "Service restarted",
			Description: "API service was restarted",
			Type:        "info",
			Timestamp:   now.Add(-2 * time.Hour),
		},
	}
	
	return models.ActivityData{
		Items: items,
	}
}

// getUsersTableData returns sample users table data
func (s *DashboardService) getUsersTableData() models.TableData {
	return models.TableData{
		Headers: []string{"ID", "Username", "Email", "Role", "Status"},
		Rows: [][]interface{}{
			{1, "admin", "admin@localhost", "admin", "active"},
			{2, "johndoe", "john@example.com", "user", "active"},
			{3, "janedoe", "jane@example.com", "user", "inactive"},
		},
	}
}

// getSystemInfoData returns system information
func (s *DashboardService) getSystemInfoData() interface{} {
	return map[string]interface{}{
		"go_version":    runtime.Version(),
		"architecture":  runtime.GOARCH,
		"os":           runtime.GOOS,
		"goroutines":   runtime.NumGoroutine(),
		"uptime":       "2 days, 14 hours",
		"last_restart": time.Now().Add(-2*24*time.Hour - 14*time.Hour).Format("2006-01-02 15:04:05"),
	}
}