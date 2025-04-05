// BlockHawk - Cross-Chain Security Monitor
// Author: Mohammad Reza Ashouri (@mohammadreza-ashouri)
// Paris Blockchain Week Hackathon 2025

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mohammadreza-ashouri/blockhawk/internal/analyzer"
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
)

func main() {
	log.Println("Starting BlockHawk - Cross-Chain Security Monitor")

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
