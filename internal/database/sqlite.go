package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mohammadreza-ashouri/blockhawk/internal/models"
)

// SQLiteStore implements a database store using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Ensure the directory exists
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %v", err)
		}
	}

	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Set pragmas for better performance
	_, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to set pragmas: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &SQLiteStore{db: db}, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Init initializes the database schema
func (s *SQLiteStore) Init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE,
			password_hash TEXT,
			api_key TEXT UNIQUE,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			name TEXT,
			network TEXT,
			webhook_url TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS contract_addresses (
			project_id TEXT,
			address TEXT,
			PRIMARY KEY(project_id, address),
			FOREIGN KEY(project_id) REFERENCES projects(id)
		)`,
		`CREATE TABLE IF NOT EXISTS security_alerts (
			id TEXT PRIMARY KEY,
			project_id TEXT,
			alert_type TEXT,
			severity TEXT,
			title TEXT,
			description TEXT,
			transaction_hash TEXT,
			is_resolved INTEGER,
			created_at TIMESTAMP,
			FOREIGN KEY(project_id) REFERENCES projects(id)
		)`
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %s: %v", query, err)
		}
	}

	return nil
}

// CreateUser creates a new user
func (s *SQLiteStore) CreateUser(user *models.User) error {
	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Generate API key if not provided
	if user.APIKey == "" {
		user.APIKey = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (id, email, password_hash, api_key, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.APIKey,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (s *SQLiteStore) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, api_key, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	row := s.db.QueryRow(query, email)

	user := &models.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.APIKey,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %v", err)
	}

	return user, nil
}

// GetUserByAPIKey retrieves a user by API key
func (s *SQLiteStore) GetUserByAPIKey(apiKey string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, api_key, created_at, updated_at
		FROM users
		WHERE api_key = ?
	`

	row := s.db.QueryRow(query, apiKey)

	user := &models.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.APIKey,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by API key: %v", err)
	}

	return user, nil
}

// CreateProject creates a new project
func (s *SQLiteStore) CreateProject(project *models.Project) error {
	// Generate UUID if not provided
	if project.ID == "" {
		project.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	// Begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert project
	webhookURL := sql.NullString{}
	if project.WebhookURL.Valid {
		webhookURL = project.WebhookURL
	}

	query := `
		INSERT INTO projects (id, user_id, name, network, webhook_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = tx.Exec(
		query,
		project.ID,
		project.UserID,
		project.Name,
		project.Network,
		webhookURL,
		project.CreatedAt,
		project.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create project: %v", err)
	}

	// Insert contract addresses
	for _, address := range project.ContractAddresses {
		_, err = tx.Exec(
			"INSERT INTO contract_addresses (project_id, address) VALUES (?, ?)",
			project.ID,
			address,
		)
		if err != nil {
			return fmt.Errorf("failed to insert contract address: %v", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetUserProjects retrieves all projects for a user
func (s *SQLiteStore) GetUserProjects(userID string) ([]*models.Project, error) {
	query := `
		SELECT id, user_id, name, network, webhook_url, created_at, updated_at
		FROM projects
		WHERE user_id = ?
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user projects: %v", err)
	}
	defer rows.Close()

	var projects []*models.Project

	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID,
			&project.UserID,
			&project.Name,
			&project.Network,
			&project.WebhookURL,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project row: %v", err)
		}

		// Fetch contract addresses
		addressRows, err := s.db.Query(
			"SELECT address FROM contract_addresses WHERE project_id = ?",
			project.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query contract addresses: %v", err)
		}
		defer addressRows.Close()

		var addresses []string
		for addressRows.Next() {
			var address string
			if err := addressRows.Scan(&address); err != nil {
				return nil, fmt.Errorf("failed to scan address row: %v", err)
			}
			addresses = append(addresses, address)
		}

		project.ContractAddresses = addresses
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating project rows: %v", err)
	}

	return projects, nil
}

// GetProject retrieves a project by ID
func (s *SQLiteStore) GetProject(projectID string) (*models.Project, error) {
	query := `
		SELECT id, user_id, name, network, webhook_url, created_at, updated_at
		FROM projects
		WHERE id = ?
	`

	row := s.db.QueryRow(query, projectID)

	project := &models.Project{}
	err := row.Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.Network,
		&project.WebhookURL,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}

	// Fetch contract addresses
	addressRows, err := s.db.Query(
		"SELECT address FROM contract_addresses WHERE project_id = ?",
		project.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query contract addresses: %v", err)
	}
	defer addressRows.Close()

	var addresses []string
	for addressRows.Next() {
		var address string
		if err := addressRows.Scan(&address); err != nil {
			return nil, fmt.Errorf("failed to scan address row: %v", err)
		}
		addresses = append(addresses, address)
	}

	project.ContractAddresses = addresses
	return project, nil
}

// CreateAlert creates a new security alert
func (s *SQLiteStore) CreateAlert(alert *models.SecurityAlert) error {
	// Generate UUID if not provided
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}

	query := `
		INSERT INTO security_alerts 
		(id, project_id, alert_type, severity, title, description, transaction_hash, is_resolved, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		alert.ID,
		alert.ProjectID,
		alert.Type,
		alert.Severity,
		alert.Title,
		alert.Description,
		alert.TransactionHash,
		alert.IsResolved,
		alert.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create alert: %v", err)
	}

	return nil
}

// GetProjectAlerts retrieves alerts for a project
func (s *SQLiteStore) GetProjectAlerts(projectID string, limit int) ([]*models.SecurityAlert, error) {
	query := `
		SELECT id, project_id, alert_type, severity, title, description, transaction_hash, is_resolved, created_at
		FROM security_alerts
		WHERE project_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query project alerts: %v", err)
	}
	defer rows.Close()

	var alerts []*models.SecurityAlert

	for rows.Next() {
		alert := &models.SecurityAlert{}
		err := rows.Scan(
			&alert.ID,
			&alert.ProjectID,
			&alert.Type,
			&alert.Severity,
			&alert.Title,
			&alert.Description,
			&alert.TransactionHash,
			&alert.IsResolved,
			&alert.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert row: %v", err)
		}

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert rows: %v", err)
	}

	return alerts, nil
}