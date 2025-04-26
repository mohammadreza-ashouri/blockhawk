// internal/models/models.go
package models

import (
    "time"
    "database/sql"
)

type User struct {
    ID           string         `db:"id" json:"id"`
    Email        string         `db:"email" json:"email"`
    PasswordHash string         `db:"password_hash" json:"-"`
    APIKey       string         `db:"api_key" json:"api_key"`
    CreatedAt    time.Time      `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}

type Project struct {
    ID                string         `db:"id" json:"id"`
    UserID            string         `db:"user_id" json:"user_id"`
    Name              string         `db:"name" json:"name"`
    Network           string         `db:"network" json:"network"`
    WebhookURL        sql.NullString `db:"webhook_url" json:"webhook_url,omitempty"`
    ContractAddresses []string       `json:"contract_addresses"`
    CreatedAt         time.Time      `db:"created_at" json:"created_at"`
    UpdatedAt         time.Time      `db:"updated_at" json:"updated_at"`
}

type SecurityAlert struct {
    ID              string    `db:"id" json:"id"`
    ProjectID       string    `db:"project_id" json:"project_id"`
    Type            string    `db:"alert_type" json:"type"`
    Severity        string    `db:"severity" json:"severity"`
    Title           string    `db:"title" json:"title"`
    Description     string    `db:"description" json:"description"`
    TransactionHash string    `db:"transaction_hash" json:"transaction_hash,omitempty"`
    IsResolved      bool      `db:"is_resolved" json:"is_resolved"`
    CreatedAt       time.Time `db:"created_at" json:"created_at"`
}