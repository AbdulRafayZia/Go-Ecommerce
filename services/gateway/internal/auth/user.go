package auth

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUsername   = errors.New("invalid username")
)

// User represents a user in the system
type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	Roles        []string
	Active       bool
}

// UserStore manages user data
type UserStore interface {
	GetByUsername(username string) (*User, error)
	GetByID(id string) (*User, error)
	Create(username, email, password string, roles []string) (*User, error)
	ValidateCredentials(username, password string) (*User, error)
}

// InMemoryUserStore is a simple in-memory user store for demonstration
// In production, this would be replaced with a database-backed store
type InMemoryUserStore struct {
	users map[string]*User // username -> user
	mu    sync.RWMutex
}

// NewInMemoryUserStore creates a new in-memory user store with default users
func NewInMemoryUserStore() *InMemoryUserStore {
	store := &InMemoryUserStore{
		users: make(map[string]*User),
	}

	// Create default users for testing
	// Password for all default users: "password123"
	defaultPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Use fixed UUIDs for default users to ensure consistency across restarts
	store.users["admin"] = &User{
		ID:           "00000000-0000-0000-0000-000000000001",
		Username:     "admin",
		Email:        "admin@gocommerce.com",
		PasswordHash: string(defaultPassword),
		Roles:        []string{"admin", "user"},
		Active:       true,
	}

	store.users["user"] = &User{
		ID:           "00000000-0000-0000-0000-000000000002",
		Username:     "user",
		Email:        "user@gocommerce.com",
		PasswordHash: string(defaultPassword),
		Roles:        []string{"user"},
		Active:       true,
	}

	store.users["testuser"] = &User{
		ID:           "00000000-0000-0000-0000-000000000003",
		Username:     "testuser",
		Email:        "test@gocommerce.com",
		PasswordHash: string(defaultPassword),
		Roles:        []string{"user"},
		Active:       true,
	}

	return store
}

// GetByUsername retrieves a user by username
func (s *InMemoryUserStore) GetByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (s *InMemoryUserStore) GetByID(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}

	return nil, ErrUserNotFound
}

// Create creates a new user
func (s *InMemoryUserStore) Create(username, email, password string, roles []string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	if _, exists := s.users[username]; exists {
		return nil, ErrUserAlreadyExists
	}

	// Validate username
	if username == "" || len(username) < 3 {
		return nil, ErrInvalidUsername
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Default to user role if no roles specified
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	user := &User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		Roles:        roles,
		Active:       true,
	}

	s.users[username] = user
	return user, nil
}

// ValidateCredentials validates username and password and returns the user
func (s *InMemoryUserStore) ValidateCredentials(username, password string) (*User, error) {
	user, err := s.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	// Check if user is active
	if !user.Active {
		return nil, errors.New("user account is disabled")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}
