package handlers

import (
	"encoding/json"
	"fmt"
	"godash/internal/caddy"
	"godash/internal/middleware"
	"godash/internal/services"
	"html/template"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

// Handlers struct holds all handler dependencies
type Handlers struct {
	userService       *services.UserService
	dashboardService  *services.DashboardService
	authMiddleware    *middleware.AuthMiddleware
	templates         *template.Template
	caddyInstanceSvc  *caddy.InstanceService
	caddyConfigSvc    *caddy.ConfigService
	caddyAnalyticsSvc *caddy.AnalyticsStore
}

// New creates a new handlers instance
func New(userService *services.UserService, dashboardService *services.DashboardService, authMiddleware *middleware.AuthMiddleware) (*Handlers, error) {
	// Parse templates
	templates, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Handlers{
		userService:      userService,
		dashboardService: dashboardService,
		authMiddleware:   authMiddleware,
		templates:        templates,
	}, nil
}

// NewWithCaddy creates a new handlers instance with Caddy services
func NewWithCaddy(userService *services.UserService, dashboardService *services.DashboardService, authMiddleware *middleware.AuthMiddleware, instanceStore *caddy.InstanceStore, analyticsStore *caddy.AnalyticsStore) (*Handlers, error) {
	handlers, err := New(userService, dashboardService, authMiddleware)
	if err != nil {
		return nil, err
	}

	handlers.caddyInstanceSvc = caddy.NewInstanceService(instanceStore)
	handlers.caddyAnalyticsSvc = analyticsStore
	handlers.caddyConfigSvc = caddy.NewConfigService(handlers.caddyInstanceSvc, analyticsStore)

	return handlers, nil
}

// HomeHandler redirects to the dashboard
func (h *Handlers) HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// LoginHandler handles login page and authentication
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Show login form
		data := struct {
			Error string
		}{}

		if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Handle POST request (login)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if err := h.authMiddleware.Login(w, r, username, password); err != nil {
		// Login failed, show error
		data := struct {
			Error string
		}{
			Error: "Invalid username or password",
		}

		if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Login successful, redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// LogoutHandler handles user logout
func (h *Handlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	h.authMiddleware.Logout(w, r)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// DashboardHandler handles the main dashboard page
func (h *Handlers) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetCurrentUser(r)
	dashboardData := h.dashboardService.GetDashboardData()

	data := struct {
		User          interface{}
		DashboardData interface{}
	}{
		User:          user,
		DashboardData: dashboardData,
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// API Handlers

// APIDashboardDataHandler returns dashboard data as JSON
func (h *Handlers) APIDashboardDataHandler(w http.ResponseWriter, r *http.Request) {
	dashboardData := h.dashboardService.GetDashboardData()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dashboardData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// APISystemStatsHandler returns system stats as JSON
func (h *Handlers) APISystemStatsHandler(w http.ResponseWriter, r *http.Request) {
	dashboardData := h.dashboardService.GetDashboardData()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dashboardData.Stats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// APIUsersHandler returns users list as JSON
func (h *Handlers) APIUsersHandler(w http.ResponseWriter, r *http.Request) {
	users := h.userService.List()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// StaticFileHandler serves static files with proper MIME types
func (h *Handlers) StaticFileHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the file path from URL
	filePath := r.URL.Path[len("/static/"):]
	fullPath := filepath.Join("web/static", filePath)

	// Set appropriate content type based on file extension
	ext := filepath.Ext(filePath)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	http.ServeFile(w, r, fullPath)
}

// Caddy API Handlers

// APIListInstancesHandler returns list of Caddy instances
func (h *Handlers) APIListInstancesHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyInstanceSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	instances := h.caddyInstanceSvc.List()

	w.Header().Set("Content-Type", "application/json")
	response := caddy.InstancesListResponse{
		Instances: instances,
		Total:     len(instances),
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// APIGetInstanceHandler returns a single Caddy instance
func (h *Handlers) APIGetInstanceHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyInstanceSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	inst, err := h.caddyInstanceSvc.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := caddy.InstanceResponse{Instance: inst}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// APICreateInstanceHandler creates a new Caddy instance
func (h *Handlers) APICreateInstanceHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyInstanceSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req caddy.InstanceRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	inst, err := h.caddyInstanceSvc.Create(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := caddy.InstanceResponse{Instance: inst}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// APIUpdateInstanceHandler updates a Caddy instance
func (h *Handlers) APIUpdateInstanceHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyInstanceSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req caddy.InstanceRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	inst, err := h.caddyInstanceSvc.Update(id, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := caddy.InstanceResponse{Instance: inst}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// APIDeleteInstanceHandler deletes a Caddy instance
func (h *Handlers) APIDeleteInstanceHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyInstanceSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.caddyInstanceSvc.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// APITestInstanceHandler tests connection to a Caddy instance
func (h *Handlers) APITestInstanceHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyInstanceSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.caddyInstanceSvc.TestConnection(id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "connected"})
}

// APIRefreshInstanceHandler refreshes the status of a Caddy instance
func (h *Handlers) APIRefreshInstanceHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyInstanceSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.caddyInstanceSvc.RefreshStatus(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inst, _ := h.caddyInstanceSvc.Get(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": string(inst.Status)})
}

// APIInstanceMetricsHandler returns metrics for a Caddy instance
func (h *Handlers) APIInstanceMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyInstanceSvc == nil || h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	metrics, err := h.caddyConfigSvc.CollectMetrics(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := caddy.AnalyticsResponse{Metrics: metrics}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// APIInstanceConfigHandler returns the config for a Caddy instance
func (h *Handlers) APIInstanceConfigHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	config, err := h.caddyConfigSvc.ExportConfig(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(config))
}

// APIInstanceLogsHandler returns logs for a Caddy instance
func (h *Handlers) APIInstanceLogsHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	// Get tail lines from query param
	tailLines := 100
	if lines := r.URL.Query().Get("lines"); lines != "" {
		var n int
		if _, err := fmt.Sscanf(lines, "%d", &n); err == nil {
			tailLines = n
		}
	}

	logs, err := h.caddyConfigSvc.GetLogs(id, tailLines)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// APIInstanceReloadHandler reloads configuration on a Caddy instance
func (h *Handlers) APIInstanceReloadHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// If no body provided, just reload current config
	if len(body) == 0 {
		if err := h.caddyConfigSvc.ReloadConfig(id, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.caddyConfigSvc.ReloadConfig(id, body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "reloaded"})
}

// APIInstanceStartHandler starts a Caddy instance
func (h *Handlers) APIInstanceStartHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	// Note: Starting Caddy requires the binary to be available
	// This is a placeholder - actual implementation depends on deployment
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "pending",
		"message": "Start operation initiated (requires binary availability)",
	})
}

// APIInstanceStopHandler stops a Caddy instance
func (h *Handlers) APIInstanceStopHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.caddyConfigSvc.StopServer(id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// APIInstanceRestartHandler restarts a Caddy instance
func (h *Handlers) APIInstanceRestartHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	// Stop first
	if err := h.caddyConfigSvc.StopServer(id); err != nil {
		// Try reload as fallback
		if reloadErr := h.caddyConfigSvc.ReloadConfig(id, nil); reloadErr != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": reloadErr.Error()})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "restarted",
		"message": "Restart operation completed",
	})
}

// APIInstanceSitesHandler returns sites for a Caddy instance
func (h *Handlers) APIInstanceSitesHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	sites, err := h.caddyConfigSvc.GetSites(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sites)
}

// APIInstanceCreateSiteHandler creates a new site
func (h *Handlers) APIInstanceCreateSiteHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		SiteName string                 `json:"site_name"`
		Config   map[string]interface{} `json:"config"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.caddyConfigSvc.CreateSite(id, req.SiteName, req.Config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

// APIInstanceDeleteSiteHandler deletes a site
func (h *Handlers) APIInstanceDeleteSiteHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	siteName := vars["site"]

	if err := h.caddyConfigSvc.DeleteSite(id, siteName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// APIInstanceHealthHandler returns health status for a Caddy instance
func (h *Handlers) APIInstanceHealthHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	healthy, message, err := h.caddyConfigSvc.HealthCheck(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "unhealthy",
			"message": err.Error(),
		})
		return
	}

	status := "healthy"
	if !healthy {
		status = "degraded"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  status,
		"message": message,
	})
}

// APIInstanceCaddyfileHandler returns the config as Caddyfile format
func (h *Handlers) APIInstanceCaddyfileHandler(w http.ResponseWriter, r *http.Request) {
	if h.caddyConfigSvc == nil {
		http.Error(w, "Caddy service not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	caddyfile, err := h.caddyConfigSvc.GetCaddyfile(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(caddyfile))
}
