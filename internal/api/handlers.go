package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mohammadreza-ashouri/blockhawk/internal/database"
	"github.com/mohammadreza-ashouri/blockhawk/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type API struct {
	store *database.PostgresStore
}

func NewAPI(store *database.PostgresStore) *API {
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

	// Check if user exists
	existing, _ := api.store.GetUserByEmail(req.Email)
	if existing != nil {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := api.store.CreateUser(user); err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
		"api_key": user.APIKey,
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

	user, err := api.store.GetUserByEmail(req.Email)
	if err != nil || user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"api_key": user.APIKey,
	})
}

// CreateProject creates a new project
func (api *API) CreateProject(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	user, err := api.store.GetUserByAPIKey(apiKey)
	if err != nil || user == nil {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

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

	project := &models.Project{
		UserID:            user.ID,
		Name:              req.Name,
		ContractAddresses: req.ContractAddresses,
		Network:           req.Network,
	}

	if req.WebhookURL != "" {
		project.WebhookURL.String = req.WebhookURL
		project.WebhookURL.Valid = true
	}

	if err := api.store.CreateProject(project); err != nil {
		http.Error(w, "Error creating project", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(project)
}

// GetProjects returns all projects for a user
func (api *API) GetProjects(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	user, err := api.store.GetUserByAPIKey(apiKey)
	if err != nil || user == nil {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	projects, err := api.store.GetUserProjects(user.ID)
	if err != nil {
		http.Error(w, "Error getting projects", http.StatusInternalServerError)
		return
	}

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

	// TODO: Implement GetProject method in store
	w.WriteHeader(http.StatusNotImplemented)
}

// GetAlerts returns alerts for a project
func (api *API) GetAlerts(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}

	alerts, err := api.store.GetProjectAlerts(projectID, 50)
	if err != nil {
		http.Error(w, "Error getting alerts", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(alerts)
}

// ServeDashboard serves the dashboard template
func (api *API) ServeDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/dashboard.html")
}
