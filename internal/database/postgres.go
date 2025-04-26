// internal/database/postgres.go
package database

import (
    "database/sql"
    "fmt"
    "time"
    
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "github.com/mohammadreza-ashouri/blockhawk/internal/models"
)

type PostgresStore struct {
    db *sqlx.DB
}

func NewPostgresStore(connectionString string) (*PostgresStore, error) {
    db, err := sqlx.Connect("postgres", connectionString)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // Set connection pool settings
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return &PostgresStore{db: db}, nil
}

// User operations
func (s *PostgresStore) CreateUser(user *models.User) error {
    query := `
        INSERT INTO users (email, password_hash, api_key)
        VALUES ($1, $2, gen_random_uuid())
        RETURNING id, api_key, created_at
    `
    return s.db.QueryRowx(query, user.Email, user.Password).
        Scan(&user.ID, &user.APIKey, &user.CreatedAt)
}

func (s *PostgresStore) GetUserByEmail(email string) (*models.User, error) {
    user := &models.User{}
    query := `SELECT * FROM users WHERE email = $1`
    err := s.db.Get(user, query, email)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return user, err
}

func (s *PostgresStore) GetUserByAPIKey(apiKey string) (*models.User, error) {
    user := &models.User{}
    query := `SELECT * FROM users WHERE api_key = $1`
    err := s.db.Get(user, query, apiKey)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return user, err
}

// Project operations
func (s *PostgresStore) CreateProject(project *models.Project) error {
    tx, err := s.db.Beginx()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Insert project
    query := `
        INSERT INTO projects (user_id, name, network, webhook_url)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at
    `
    err = tx.QueryRowx(query, project.UserID, project.Name, project.Network, project.WebhookURL).
        Scan(&project.ID, &project.CreatedAt)
    if err != nil {
        return err
    }
    
    // Insert contract addresses
    if len(project.ContractAddresses) > 0 {
        for _, address := range project.ContractAddresses {
            _, err = tx.Exec(`
                INSERT INTO contract_addresses (project_id, address)
                VALUES ($1, $2)
            `, project.ID, address)
            if err != nil {
                return err
            }
        }
    }
    
    return tx.Commit()
}

func (s *PostgresStore) GetUserProjects(userID string) ([]*models.Project, error) {
    query := `
        SELECT p.*, array_agg(ca.address) as contract_addresses
        FROM projects p
        LEFT JOIN contract_addresses ca ON p.id = ca.project_id
        WHERE p.user_id = $1
        GROUP BY p.id
        ORDER BY p.created_at DESC
    `
    
    rows, err := s.db.Queryx(query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var projects []*models.Project
    for rows.Next() {
        var project models.Project
        var addresses sql.NullString
        
        err := rows.Scan(
            &project.ID,
            &project.UserID,
            &project.Name,
            &project.Network,
            &project.WebhookURL,
            &project.CreatedAt,
            &project.UpdatedAt,
            &addresses,
        )
        if err != nil {
            return nil, err
        }
        
        if addresses.Valid {
            project.ContractAddresses = strings.Split(addresses.String, ",")
        }
        
        projects = append(projects, &project)
    }
    
    return projects, nil
}

// Security alert operations
func (s *PostgresStore) CreateSecurityAlert(alert *models.SecurityAlert) error {
    query := `
        INSERT INTO security_alerts 
        (project_id, alert_type, severity, title, description, transaction_hash)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at
    `
    return s.db.QueryRowx(
        query,
        alert.ProjectID,
        alert.Type,
        alert.Severity,
        alert.Title,
        alert.Description,
        alert.TransactionHash,
    ).Scan(&alert.ID, &alert.CreatedAt)
}

func (s *PostgresStore) GetProjectAlerts(projectID string, limit int) ([]*models.SecurityAlert, error) {
    query := `
        SELECT * FROM security_alerts
        WHERE project_id = $1
        ORDER BY created_at DESC
        LIMIT $2
    `
    
    var alerts []*models.SecurityAlert
    err := s.db.Select(&alerts, query, projectID, limit)
    return alerts, err
}