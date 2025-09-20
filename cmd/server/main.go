package main

import (
	"fmt"
	"godash/internal/config"
	"godash/internal/handlers"
	"godash/internal/middleware"
	"godash/internal/services"
	"log"
	"net/http"

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

	// Initialize handlers
	h, err := handlers.New(userService, dashboardService, authMiddleware)
	if err != nil {
		log.Fatalf("Failed to initialize handlers: %v", err)
	}

	// Setup routes
	r := mux.NewRouter()

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

	// Admin API routes
	adminAPI := api.PathPrefix("/admin").Subrouter()
	adminAPI.Use(authMiddleware.RequireAdmin)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Dashboard available at: http://%s/dashboard", addr)
	log.Printf("Default credentials: admin / password")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}