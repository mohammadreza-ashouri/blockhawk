// internal/models/user.go
package models

// Add only new types here that aren't in models.go
// For example, you might add:
type UserSettings struct {
	UserID         string `json:"user_id"`
	EmailAlerts    bool   `json:"email_alerts"`
	WebhookEnabled bool   `json:"webhook_enabled"`
}
