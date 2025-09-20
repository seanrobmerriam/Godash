package middleware

import (
	"context"
	"godash/internal/models"
	"godash/internal/services"
	"net/http"

	"github.com/gorilla/sessions"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	store       *sessions.CookieStore
	userService *services.UserService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(secretKey string, userService *services.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		store:       sessions.NewCookieStore([]byte(secretKey)),
		userService: userService,
	}
}

// RequireAuth is middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := m.store.Get(r, "session")
		
		userID, ok := session.Values["user_id"].(int)
		if !ok || userID == 0 {
			// Not authenticated, redirect to login
			if isAPIRequest(r) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		
		// Get user from service
		user, err := m.userService.GetByID(userID)
		if err != nil || !user.Active {
			// User not found or inactive, clear session and redirect
			session.Values["user_id"] = nil
			session.Save(r, w)
			
			if isAPIRequest(r) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		
		// Add user to context
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin is middleware that requires admin privileges
func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(*models.User)
		
		if !user.IsAdmin() {
			if isAPIRequest(r) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		
		next.ServeHTTP(w, r)
	}))
}

// Login authenticates a user and creates a session
func (m *AuthMiddleware) Login(w http.ResponseWriter, r *http.Request, username, password string) error {
	user, err := m.userService.Authenticate(username, password)
	if err != nil {
		return err
	}
	
	session, _ := m.store.Get(r, "session")
	session.Values["user_id"] = user.ID
	session.Save(r, w)
	
	return nil
}

// Logout destroys the user session
func (m *AuthMiddleware) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := m.store.Get(r, "session")
	session.Values["user_id"] = nil
	session.Save(r, w)
}

// GetCurrentUser returns the current authenticated user from context
func GetCurrentUser(r *http.Request) *models.User {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		return nil
	}
	return user
}

// isAPIRequest checks if the request is an API request
func isAPIRequest(r *http.Request) bool {
	return r.Header.Get("Content-Type") == "application/json" ||
		   r.Header.Get("Accept") == "application/json" ||
		   r.URL.Path[:4] == "/api"
}