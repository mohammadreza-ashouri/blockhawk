// BlockHawk - Cross-Chain Security Monitor
// Author: Mohammad Reza Ashouri (@mohammadreza-ashouri)
// Paris Blockchain Week Hackathon 2025

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mohammadreza-ashouri/blockhawk/internal/analyzer"
	"github.com/mohammadreza-ashouri/blockhawk/internal/api" // Add this import
	"github.com/mohammadreza-ashouri/blockhawk/internal/config"
	"github.com/mohammadreza-ashouri/blockhawk/internal/database"
	"github.com/mohammadreza-ashouri/blockhawk/internal/models"
	"github.com/mohammadreza-ashouri/blockhawk/internal/ripple"
	"github.com/mohammadreza-ashouri/blockhawk/internal/solana"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow any origin for hackathon
		},
	}

	// Recent alerts storage
	recentAlerts     = make([]models.SecurityAlert, 0, 100)
	recentAlertMutex sync.RWMutex

	// Chain status storage
	chainStatus      = make(map[models.BlockchainType]models.ChainStatus)
	chainStatusMutex sync.RWMutex

	// Database store (changed from PostgresStore to a more generic variable)
	dbStore    interface{}
	apiHandler *api.API
	useSQLite  = true // Flag to determine which database to use
)

func main() {
	log.Println("Starting BlockHawk - Cross-Chain Security Monitor")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config, using defaults: %v", err)
	}

	// Create data directory if it doesn't exist
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize database - using SQLite
	if useSQLite {
		dbPath := filepath.Join(dataDir, "blockhawk.db")
		sqliteStore, err := database.NewSQLiteStore(dbPath)
		if err != nil {
			log.Printf("Warning: Failed to connect to SQLite database, running without user features: %v", err)
			// Continue without database features
		} else {
			dbStore = sqliteStore
			apiHandler = api.NewAPI(sqliteStore)
			log.Println("Successfully connected to SQLite database")
		}
	} else {
		// Original PostgreSQL code (kept for reference/fallback)
		postgresStore, err := database.NewPostgresStore(cfg.DatabaseURL)
		if err != nil {
			log.Printf("Warning: Failed to connect to Postgres database, running without user features: %v", err)
			// Continue without database features
		} else {
			dbStore = postgresStore
			apiHandler = api.NewAPI(postgresStore)
			log.Println("Successfully connected to PostgreSQL database")
		}
	}

	// Create and start blockchain monitors
	solanaMonitor := solana.NewMonitor()
	rippleMonitor := ripple.NewMonitor()
	securityAnalyzer := analyzer.NewAnalyzer()

	// Start all components
	solanaMonitor.Start()
	rippleMonitor.Start()
	securityAnalyzer.Start()

	// Handle clean shutdown
	defer func() {
		solanaMonitor.Stop()
		rippleMonitor.Stop()
		securityAnalyzer.Stop()

		// Close database connection if using SQLite
		if useSQLite {
			if store, ok := dbStore.(*database.SQLiteStore); ok && store != nil {
				store.Close()
			}
		}
	}()

	// Connect transaction streams to analyzer
	go forwardTransactions(solanaMonitor.GetTransactionChannel(), securityAnalyzer)
	go forwardTransactions(rippleMonitor.GetTransactionChannel(), securityAnalyzer)

	// Process alerts
	go processAlerts(securityAnalyzer.GetAlertChannel())

	// Track chain status
	go trackChainStatus(solanaMonitor.GetStatusChannel())
	go trackChainStatus(rippleMonitor.GetStatusChannel())

	// Set up HTTP server
	setupHTTPServer()

	log.Println("BlockHawk is running. Press Ctrl+C to exit.")
	select {} // Run forever
}

// forwardTransactions sends blockchain transactions to the analyzer
func forwardTransactions(txChan <-chan models.Transaction, analyzer *analyzer.Analyzer) {
	for tx := range txChan {
		analyzer.ProcessTransaction(tx)
	}
}

// processAlerts handles security alerts from the analyzer
func processAlerts(alertChan <-chan models.SecurityAlert) {
	for alert := range alertChan {
		// Store recent alerts (keep last 100)
		recentAlertMutex.Lock()
		recentAlerts = append(recentAlerts, alert)
		if len(recentAlerts) > 100 {
			recentAlerts = recentAlerts[len(recentAlerts)-100:]
		}
		recentAlertMutex.Unlock()

		// Log alert
		log.Printf("ALERT [%s] %s: %s",
			alert.Severity,
			alert.Title,
			alert.Description)
	}
}

// trackChainStatus updates the status of each blockchain
func trackChainStatus(statusChan <-chan models.ChainStatus) {
	for status := range statusChan {
		chainStatusMutex.Lock()
		chainStatus[status.Chain] = status
		chainStatusMutex.Unlock()
	}
}

// Helper functions for authentication middleware
func withAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if apiHandler == nil {
			http.Error(w, "Feature not available", http.StatusServiceUnavailable)
			return
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}

		var user *models.User
		var err error

		if useSQLite {
			store, ok := dbStore.(*database.SQLiteStore)
			if !ok || store == nil {
				http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
				return
			}
			user, err = store.GetUserByAPIKey(apiKey)
		} else {
			store, ok := dbStore.(*database.PostgresStore)
			if !ok || store == nil {
				http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
				return
			}
			user, err = store.GetUserByAPIKey(apiKey)
		}

		if err != nil || user == nil {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// Add user to context (optional, if you need it in handlers)
		next(w, r)
	}
}

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if apiHandler == nil {
			// Serve regular dashboard if no auth available
			http.ServeFile(w, r, "./web/templates/index.html")
			return
		}

		// Check for session cookie or API key
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Redirect to login page
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		next(w, r)
	}
}

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/dashboard.html")
}

// setupHTTPServer configures the web server
func setupHTTPServer() {
	// Serve static files
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve docs files
	docsFs := http.FileServer(http.Dir("./web/docs"))
	http.Handle("/docs/", http.StripPrefix("/docs/", docsFs))

	// API endpoints
	http.HandleFunc("/api/alerts", getAlertsHandler)
	http.HandleFunc("/api/status", getStatusHandler)

	// WebSocket endpoint
	http.HandleFunc("/ws", wsHandler)

	// Main dashboard page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/templates/index.html")
	})

	// Add authentication and project management endpoints if database is available
	if apiHandler != nil {
		// User authentication endpoints
		http.HandleFunc("/api/v1/register", apiHandler.RegisterUser)
		http.HandleFunc("/api/v1/login", apiHandler.LoginUser)

		// Serve login and register pages
		http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./web/templates/login.html")
		})
		http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./web/templates/register.html")
		})

		// Project management endpoints
		http.HandleFunc("/api/v1/projects", withAPIKey(apiHandler.CreateProject))
		http.HandleFunc("/api/v1/projects/list", withAPIKey(apiHandler.GetProjects))
		http.HandleFunc("/api/v1/projects/", withAPIKey(apiHandler.GetProject))
		http.HandleFunc("/api/v1/alerts", withAPIKey(apiHandler.GetAlerts))

		// Dashboard with authentication
		http.HandleFunc("/dashboard", withAuth(serveDashboard))
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Println("Starting web server on http://localhost:8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
}

// getAlertsHandler returns recent security alerts
func getAlertsHandler(w http.ResponseWriter, r *http.Request) {
	recentAlertMutex.RLock()
	defer recentAlertMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recentAlerts)
}

// getStatusHandler returns the status of all blockchains
func getStatusHandler(w http.ResponseWriter, r *http.Request) {
	chainStatusMutex.RLock()
	defer chainStatusMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chainStatus)
}

// wsHandler handles WebSocket connections for real-time updates
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Client connected
	log.Println("New WebSocket client connected")

	// Send initial data
	sendInitialData(conn)

	// Keep connection alive and send updates
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send updates every 2 seconds
			if err := sendUpdates(conn); err != nil {
				log.Printf("WebSocket error: %v", err)
				return
			}
		}
	}
}

// sendInitialData sends initial data to a new WebSocket client
func sendInitialData(conn *websocket.Conn) {
	// Get chain status
	chainStatusMutex.RLock()
	statusData := chainStatus
	chainStatusMutex.RUnlock()

	// Get recent alerts
	recentAlertMutex.RLock()
	alertData := recentAlerts
	recentAlertMutex.RUnlock()

	// Create initial data packet
	initialData := map[string]interface{}{
		"type":   "initial",
		"status": statusData,
		"alerts": alertData,
	}

	// Send to client
	if err := conn.WriteJSON(initialData); err != nil {
		log.Printf("WebSocket write error: %v", err)
	}
}

// sendUpdates sends periodic updates to WebSocket clients
func sendUpdates(conn *websocket.Conn) error {
	// Get latest chain status
	chainStatusMutex.RLock()
	statusData := chainStatus
	chainStatusMutex.RUnlock()

	// Get recent alerts (last 5)
	recentAlertMutex.RLock()
	var alertData []models.SecurityAlert
	if len(recentAlerts) > 0 {
		start := 0
		if len(recentAlerts) > 5 {
			start = len(recentAlerts) - 5
		}
		alertData = recentAlerts[start:]
	}
	recentAlertMutex.RUnlock()

	// Create update packet
	updateData := map[string]interface{}{
		"type":   "update",
		"status": statusData,
		"alerts": alertData,
		"time":   time.Now(),
	}

	// Send to client
	return conn.WriteJSON(updateData)
}
