
package database

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mohammadreza-ashouri/blockhawk/internal/models"
)

// now we use in-memory store. Replace with proper DB later
type MemoryStore struct {
	users    map[string]*models.User
	projects map[string]*models.Project
	mu       sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:    make(map[string]*models.User),
		projects: make(map[string]*models.Project),
	}
}

func (s *MemoryStore) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user.ID = uuid.New().String()
	user.APIKey = uuid.New().String()
	user.CreatedAt = time.Now()

	s.users[user.ID] = user
	return nil
}

func (s *MemoryStore) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) GetUserByAPIKey(apiKey string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.APIKey == apiKey {
			return user, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) CreateProject(project *models.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	project.ID = uuid.New().String()
	project.CreatedAt = time.Now()

	s.projects[project.ID] = project
	return nil
}

func (s *MemoryStore) GetUserProjects(userID string) ([]*models.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var projects []*models.Project
	for _, project := range s.projects {
		if project.UserID == userID {
			projects = append(projects, project)
		}
	}
	return projects, nil
}
