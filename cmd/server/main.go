package main

import (
	"fmt"
	"godash/internal/caddy"
	"godash/internal/config"
	"godash/internal/handlers"
	"godash/internal/middleware"
	"godash/internal/services"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize services
	userService := services.NewUserService()
	dashboardService := services.NewDashboardService()

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Session.SecretKey, userService)

	// Parse templates
	templates, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Initialize Caddy services
	var h *handlers.Handlers

	// Create data directory
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("Warning: Could not create data directory: %v", err)
	}

	// Initialize instance store
	instanceStore, err := caddy.NewInstanceStore(filepath.Join(dataDir, "instances.json"))
	if err != nil {
		log.Printf("Warning: Could not initialize instance store: %v", err)
		log.Println("Caddy features will be unavailable")
		h, _ = handlers.New(userService, dashboardService, authMiddleware)
	} else {
		// Initialize analytics store
		analyticsStore, err := caddy.NewAnalyticsStore(filepath.Join(dataDir, "analytics"))
		if err != nil {
			log.Printf("Warning: Could not initialize analytics store: %v", err)
			log.Println("Analytics features will be unavailable")
		}

		// Initialize handlers with Caddy services
		h, err = handlers.NewWithCaddy(userService, dashboardService, authMiddleware, instanceStore, analyticsStore)
		if err != nil {
			log.Fatalf("Failed to initialize handlers: %v", err)
		}
	}

	// Setup routes
	r := mux.NewRouter()

	// Template routes (protected)
	r.Handle("/caddy/instances", authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetCurrentUser(r)
		data := struct {
			User interface{}
		}{User: user}
		templates.ExecuteTemplate(w, "instances.html", data)
	}))).Methods("GET")

	r.Handle("/caddy/analytics", authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetCurrentUser(r)
		data := struct {
			User interface{}
		}{User: user}
		templates.ExecuteTemplate(w, "analytics.html", data)
	}))).Methods("GET")

	r.Handle("/caddy/instances/{id}/config", authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetCurrentUser(r)
		data := struct {
			User interface{}
		}{User: user}
		templates.ExecuteTemplate(w, "config-editor.html", data)
	}))).Methods("GET")

	// Public routes
	r.HandleFunc("/", h.HomeHandler)
	r.HandleFunc("/login", h.LoginHandler)
	r.HandleFunc("/logout", h.LogoutHandler)

	// Static files
	r.PathPrefix("/static/").HandlerFunc(h.StaticFileHandler)

	// Protected routes
	r.Handle("/dashboard", authMiddleware.RequireAuth(http.HandlerFunc(h.DashboardHandler)))

	// API routes (protected)
	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware.RequireAuth)

	api.HandleFunc("/dashboard", h.APIDashboardDataHandler).Methods("GET")
	api.HandleFunc("/stats", h.APISystemStatsHandler).Methods("GET")
	api.HandleFunc("/users", h.APIUsersHandler).Methods("GET")

	// Caddy API routes
	caddyAPI := api.PathPrefix("/caddy").Subrouter()

	// Instance management
	caddyAPI.HandleFunc("/instances", h.APIListInstancesHandler).Methods("GET")
	caddyAPI.HandleFunc("/instances", h.APICreateInstanceHandler).Methods("POST")
	caddyAPI.HandleFunc("/instances/{id}", h.APIGetInstanceHandler).Methods("GET")
	caddyAPI.HandleFunc("/instances/{id}", h.APIUpdateInstanceHandler).Methods("PUT")
	caddyAPI.HandleFunc("/instances/{id}", h.APIDeleteInstanceHandler).Methods("DELETE")
	caddyAPI.HandleFunc("/instances/{id}/test", h.APITestInstanceHandler).Methods("POST")
	caddyAPI.HandleFunc("/instances/{id}/refresh", h.APIRefreshInstanceHandler).Methods("POST")
	caddyAPI.HandleFunc("/instances/{id}/health", h.APIInstanceHealthHandler).Methods("GET")

	// Instance operations
	caddyAPI.HandleFunc("/instances/{id}/metrics", h.APIInstanceMetricsHandler).Methods("GET")
	caddyAPI.HandleFunc("/instances/{id}/config", h.APIInstanceConfigHandler).Methods("GET")
	caddyAPI.HandleFunc("/instances/{id}/config/json", h.APIInstanceConfigHandler).Methods("GET")
	caddyAPI.HandleFunc("/instances/{id}/config/caddyfile", h.APIInstanceCaddyfileHandler).Methods("GET")
	caddyAPI.HandleFunc("/instances/{id}/reload", h.APIInstanceReloadHandler).Methods("POST")
	caddyAPI.HandleFunc("/instances/{id}/start", h.APIInstanceStartHandler).Methods("POST")
	caddyAPI.HandleFunc("/instances/{id}/stop", h.APIInstanceStopHandler).Methods("POST")
	caddyAPI.HandleFunc("/instances/{id}/restart", h.APIInstanceRestartHandler).Methods("POST")
	caddyAPI.HandleFunc("/instances/{id}/logs", h.APIInstanceLogsHandler).Methods("GET")

	// Site management
	caddyAPI.HandleFunc("/instances/{id}/sites", h.APIInstanceSitesHandler).Methods("GET")
	caddyAPI.HandleFunc("/instances/{id}/sites", h.APIInstanceCreateSiteHandler).Methods("POST")
	caddyAPI.HandleFunc("/instances/{id}/sites/{site}", h.APIInstanceDeleteSiteHandler).Methods("DELETE")

	// Admin API routes (admin only)
	adminAPI := api.PathPrefix("/admin").Subrouter()
	adminAPI.Use(authMiddleware.RequireAdmin)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Dashboard available at: http://%s/dashboard", addr)
	log.Printf("Caddy Instances: http://%s/caddy/instances", addr)
	log.Printf("Caddy Analytics: http://%s/caddy/analytics", addr)
	log.Printf("Default credentials: admin / password")
	log.Printf("Caddy API available at: http://%s/api/caddy", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
