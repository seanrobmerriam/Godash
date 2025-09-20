package services

import (
	"errors"
	"godash/internal/models"
	"sync"
)

// UserService handles user-related business logic
type UserService struct {
	users []models.User
	mutex sync.RWMutex
	idCounter int
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	service := &UserService{
		users: make([]models.User, 0),
		idCounter: 1,
	}
	
	// Create default admin user
	defaultAdmin := models.NewUser("admin", "admin@localhost", "password", models.RoleAdmin)
	defaultAdmin.ID = service.idCounter
	service.idCounter++
	service.users = append(service.users, *defaultAdmin)
	
	return service
}

// Authenticate validates user credentials
func (s *UserService) Authenticate(username, password string) (*models.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	for _, user := range s.users {
		if user.Username == username && user.Password == password && user.Active {
			// Return a copy to avoid modifying the original
			userCopy := user
			return &userCopy, nil
		}
	}
	
	return nil, errors.New("invalid credentials")
}

// GetByID returns a user by ID
func (s *UserService) GetByID(id int) (*models.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	for _, user := range s.users {
		if user.ID == id {
			userCopy := user
			return &userCopy, nil
		}
	}
	
	return nil, errors.New("user not found")
}

// GetByUsername returns a user by username
func (s *UserService) GetByUsername(username string) (*models.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	for _, user := range s.users {
		if user.Username == username {
			userCopy := user
			return &userCopy, nil
		}
	}
	
	return nil, errors.New("user not found")
}

// Create creates a new user
func (s *UserService) Create(user *models.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Check if username already exists
	for _, existingUser := range s.users {
		if existingUser.Username == user.Username {
			return errors.New("username already exists")
		}
		if existingUser.Email == user.Email {
			return errors.New("email already exists")
		}
	}
	
	user.ID = s.idCounter
	s.idCounter++
	s.users = append(s.users, *user)
	
	return nil
}

// Update updates an existing user
func (s *UserService) Update(user *models.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	for i, existingUser := range s.users {
		if existingUser.ID == user.ID {
			s.users[i] = *user
			return nil
		}
	}
	
	return errors.New("user not found")
}

// Delete deactivates a user (soft delete)
func (s *UserService) Delete(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	for i, user := range s.users {
		if user.ID == id {
			s.users[i].Active = false
			return nil
		}
	}
	
	return errors.New("user not found")
}

// List returns all active users
func (s *UserService) List() []models.User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var activeUsers []models.User
	for _, user := range s.users {
		if user.Active {
			activeUsers = append(activeUsers, user)
		}
	}
	
	return activeUsers
}