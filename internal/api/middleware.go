package api

import (
	"context"
	"net/http"

	"github.com/mohammadreza-ashouri/blockhawk/internal/database"
	"github.com/mohammadreza-ashouri/blockhawk/internal/models"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Key for storing the authenticated user in request context
const userContextKey contextKey = "user"

// AuthMiddleware creates a middleware for API authentication
func AuthMiddleware(store database.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key from header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, "API key required", http.StatusUnauthorized)
				return
			}

			// Get user from API key
			user, err := store.GetUserByAPIKey(apiKey)
			if err != nil {
				http.Error(w, "Error authenticating request", http.StatusInternalServerError)
				return
			}

			if user == nil {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			// Add user to request context
			ctx := context.WithValue(r.Context(), userContextKey, user)

			// Call the next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth is a middleware that requires authentication
func RequireAuth(store database.Store) func(http.HandlerFunc) http.HandlerFunc {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get API key from header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, "API key required", http.StatusUnauthorized)
				return
			}

			// Get user from API key
			user, err := store.GetUserByAPIKey(apiKey)
			if err != nil {
				http.Error(w, "Error authenticating request", http.StatusInternalServerError)
				return
			}

			if user == nil {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			// Add user to request context
			ctx := context.WithValue(r.Context(), userContextKey, user)

			// Call the handler with updated context
			handler(w, r.WithContext(ctx))
		}
	}
}

// GetAuthenticatedUser extracts the authenticated user from request context
func GetAuthenticatedUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}
