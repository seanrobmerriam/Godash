package models

import "time"

// DashboardData represents the main dashboard data structure
type DashboardData struct {
	Title       string      `json:"title"`
	RefreshRate int         `json:"refresh_rate"` // seconds
	LastUpdate  time.Time   `json:"last_update"`
	Widgets     []Widget    `json:"widgets"`
	Stats       SystemStats `json:"stats"`
}

// Widget represents a dashboard widget
type Widget struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Title    string      `json:"title"`
	Position Position    `json:"position"`
	Size     Size        `json:"size"`
	Data     interface{} `json:"data"`
	Config   interface{} `json:"config,omitempty"`
}

// Position represents widget position on the dashboard
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Size represents widget dimensions
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// SystemStats represents system statistics
type SystemStats struct {
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	Uptime      int64     `json:"uptime"`
	LastUpdated time.Time `json:"last_updated"`
}

// Widget types
const (
	WidgetTypeChart     = "chart"
	WidgetTypeTable     = "table"
	WidgetTypeMetric    = "metric"
	WidgetTypeText      = "text"
	WidgetTypeActivity  = "activity"
	WidgetTypeProgress  = "progress"
)

// ChartData represents data for chart widgets
type ChartData struct {
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"datasets"`
}

// Dataset represents a dataset in a chart
type Dataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor string    `json:"backgroundColor,omitempty"`
	BorderColor     string    `json:"borderColor,omitempty"`
}

// TableData represents data for table widgets
type TableData struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

// MetricData represents data for metric widgets
type MetricData struct {
	Value       interface{} `json:"value"`
	Unit        string      `json:"unit,omitempty"`
	Description string      `json:"description,omitempty"`
	Trend       string      `json:"trend,omitempty"` // "up", "down", "stable"
	Change      float64     `json:"change,omitempty"`
}

// ActivityData represents data for activity widgets
type ActivityData struct {
	Items []ActivityItem `json:"items"`
}

// ActivityItem represents a single activity item
type ActivityItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type"` // "info", "warning", "error", "success"
	Timestamp   time.Time `json:"timestamp"`
	User        string    `json:"user,omitempty"`
}

// ProgressData represents data for progress widgets
type ProgressData struct {
	Value       float64 `json:"value"`       // Current value
	Max         float64 `json:"max"`         // Maximum value
	Label       string  `json:"label"`       // Progress label
	Description string  `json:"description"` // Additional description
}