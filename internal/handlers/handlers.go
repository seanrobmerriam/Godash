package handlers

import (
	"encoding/json"
	"godash/internal/middleware"
	"godash/internal/services"
	"html/template"
	"net/http"
	"path/filepath"
)

// Handlers struct holds all handler dependencies
type Handlers struct {
	userService      *services.UserService
	dashboardService *services.DashboardService
	authMiddleware   *middleware.AuthMiddleware
	templates        *template.Template
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