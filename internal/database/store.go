package database

import (
	"github.com/mohammadreza-ashouri/blockhawk/internal/models"
)

// Store defines the interface for database operations
type Store interface {
	// Initialize database schema
	Init() error

	// Close the database connection
	Close() error

	// User operations
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByAPIKey(apiKey string) (*models.User, error)

	// Project operations
	CreateProject(project *models.Project) error
	GetUserProjects(userID string) ([]*models.Project, error)
	GetProject(projectID string) (*models.Project, error)

	// Alert operations
	CreateAlert(alert *models.SecurityAlert) error
	GetProjectAlerts(projectID string, limit int) ([]*models.SecurityAlert, error)
}
