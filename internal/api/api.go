package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mohammadreza-ashouri/blockhawk/internal/database"
	"github.com/mohammadreza-ashouri/blockhawk/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// API handles HTTP API requests
type API struct {
	// Change the type to interface{} to support different store types
	store interface{}
}

// NewAPI creates a new API handler with the given store
func NewAPI(store interface{}) *API {
	return &API{store: store}
}

// RegisterUser handles user registration
func (api *API) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check which store type is being used
	var user *models.User
	var err error

	// Check if user exists
	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		user, err = sqliteStore.GetUserByEmail(req.Email)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		user, err = postgresStore.GetUserByEmail(req.Email)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "Error checking user", http.StatusInternalServerError)
		return
	}

	if user != nil {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Create new user
	newUser := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	// Save user based on store type
	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		err = sqliteStore.CreateUser(newUser)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		err = postgresStore.CreateUser(newUser)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Return success and API key
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
		"api_key": newUser.APIKey,
	})
}

// LoginUser handles user login
func (api *API) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user based on store type
	var user *models.User
	var err error

	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		user, err = sqliteStore.GetUserByEmail(req.Email)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		user, err = postgresStore.GetUserByEmail(req.Email)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil || user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Return API key
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"api_key": user.APIKey,
	})
}

// CreateProject handles project creation
func (api *API) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string   `json:"name"`
		ContractAddresses []string `json:"contract_addresses"`
		Network           string   `json:"network"`
		WebhookURL        string   `json:"webhook_url,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract API key
	apiKey := r.Header.Get("X-API-Key")

	// Get user based on store type
	var user *models.User
	var err error

	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		user, err = sqliteStore.GetUserByAPIKey(apiKey)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		user, err = postgresStore.GetUserByAPIKey(apiKey)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil || user == nil {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	// Create project
	project := &models.Project{
		UserID:            user.ID,
		Name:              req.Name,
		ContractAddresses: req.ContractAddresses,
		Network:           req.Network,
	}

	// Set webhook URL if provided
	if req.WebhookURL != "" {
		project.WebhookURL.String = req.WebhookURL
		project.WebhookURL.Valid = true
	}

	// Save project based on store type
	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		err = sqliteStore.CreateProject(project)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		err = postgresStore.CreateProject(project)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "Error creating project", http.StatusInternalServerError)
		return
	}

	// Return created project
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// GetProjects returns all projects for a user
func (api *API) GetProjects(w http.ResponseWriter, r *http.Request) {
	// Extract API key
	apiKey := r.Header.Get("X-API-Key")

	// Get user based on store type
	var user *models.User
	var err error

	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		user, err = sqliteStore.GetUserByAPIKey(apiKey)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		user, err = postgresStore.GetUserByAPIKey(apiKey)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil || user == nil {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	// Get projects based on store type
	var projects []*models.Project

	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		projects, err = sqliteStore.GetUserProjects(user.ID)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		projects, err = postgresStore.GetUserProjects(user.ID)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "Error getting projects", http.StatusInternalServerError)
		return
	}

	// Return projects
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// GetProject returns a specific project
func (api *API) GetProject(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}
	projectID := parts[3]

	// Extract API key
	apiKey := r.Header.Get("X-API-Key")

	// Get user based on store type
	var user *models.User
	var err error

	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		user, err = sqliteStore.GetUserByAPIKey(apiKey)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		user, err = postgresStore.GetUserByAPIKey(apiKey)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil || user == nil {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	// Get project based on store type
	var project *models.Project

	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		project, err = sqliteStore.GetProject(projectID)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		project, err = postgresStore.GetProject(projectID)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "Error getting project", http.StatusInternalServerError)
		return
	}

	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Check if user owns this project
	if project.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Return project
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// GetAlerts returns alerts for a project
func (api *API) GetAlerts(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from query params
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}

	// Extract limit from query params, default to 50
	var limit int = 50
	// TODO: Parse limit from query params if needed

	// Extract API key
	apiKey := r.Header.Get("X-API-Key")

	// Get user based on store type
	var user *models.User
	var err error

	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		user, err = sqliteStore.GetUserByAPIKey(apiKey)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		user, err = postgresStore.GetUserByAPIKey(apiKey)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil || user == nil {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	// Get project based on store type
	var project *models.Project

	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		project, err = sqliteStore.GetProject(projectID)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		project, err = postgresStore.GetProject(projectID)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "Error getting project", http.StatusInternalServerError)
		return
	}

	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Check if user owns this project
	if project.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get alerts based on store type
	var alerts []*models.SecurityAlert

	if sqliteStore, ok := api.store.(*database.SQLiteStore); ok {
		alerts, err = sqliteStore.GetProjectAlerts(projectID, limit)
	} else if postgresStore, ok := api.store.(*database.PostgresStore); ok {
		alerts, err = postgresStore.GetProjectAlerts(projectID, limit)
	} else {
		http.Error(w, "Database store not properly configured", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "Error getting alerts", http.StatusInternalServerError)
		return
	}

	// Return alerts
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}
