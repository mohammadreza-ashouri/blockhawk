// BlockHawk - Cross-Chain Security Monitor
// Author: Mohammad Reza Ashouri (@mohammadreza-ashouri)
// Paris Blockchain Week Hackathon 2025

package ripple

import (
	"context"
	"log"
	"math/rand" // For demo purposes
	"time"

	"github.com/mohammadreza-ashouri/blockhawk/internal/models"
)

// Monitor represents a Ripple blockchain monitor
type Monitor struct {
	txChan     chan models.Transaction
	statusChan chan models.ChainStatus
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewMonitor creates a new Ripple monitor
func NewMonitor() *Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		txChan:     make(chan models.Transaction, 100),
		statusChan: make(chan models.ChainStatus, 10),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins monitoring the Ripple blockchain
func (m *Monitor) Start() {
	log.Println("Starting Ripple blockchain monitor...")

	// Start transaction monitoring goroutine
	go m.monitorTransactions()

	// Start status updates goroutine
	go m.updateStatus()
}

// Stop halts all monitoring activities
func (m *Monitor) Stop() {
	log.Println("Stopping Ripple blockchain monitor...")
	m.cancel()

	// Close channels
	close(m.txChan)
	close(m.statusChan)
}

// GetTransactionChannel returns the channel for receiving transactions
func (m *Monitor) GetTransactionChannel() <-chan models.Transaction {
	return m.txChan
}

// GetStatusChannel returns the channel for receiving chain status updates
func (m *Monitor) GetStatusChannel() <-chan models.ChainStatus {
	return m.statusChan
}

// monitorTransactions monitors transactions on the Ripple blockchain
func (m *Monitor) monitorTransactions() {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// For hackathon purposes, generate some mock transactions
			tx := generateMockRippleTransaction()

			select {
			case m.txChan <- tx:
				// Transaction sent successfully
			default:
				// Channel full, logging and continuing
				log.Println("Ripple transaction channel full, dropping transaction")
			}
		}
	}
}

// updateStatus updates the overall Ripple blockchain status
func (m *Monitor) updateStatus() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			status := models.ChainStatus{
				Chain:          models.RippleChain,
				IsConnected:    true,
				LastBlockTime:  time.Now().Add(-time.Duration(rand.Intn(4)) * time.Second),
				TPS:            1500 + rand.Float64()*300,
				PendingTxCount: rand.Intn(500),
				SecurityScore:  85 + rand.Float64()*15,
			}

			select {
			case m.statusChan <- status:
				// Status sent successfully
			default:
				// Channel full, logging and continuing
				log.Println("Ripple status channel full, dropping status update")
			}
		}
	}
}

// Helper function to generate mock Ripple transactions
func generateMockRippleTransaction() models.Transaction {
	wallets := []string{
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"rDsbeomae4FXwgQTJp9Rs64Qg9vDiTCdBv",
		"rUAi7YnAJLxqCHEy9C5cWmd7jJbfuCajgh",
		"rLHzPsX6oXkzU2qL12kHCH8G8cnZv1rBJh",
		"r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	}

	txTypes := []string{"payment", "escrow_create", "escrow_finish", "trust_set", "account_set"}

	// Occasionally generate high-value or high-fee transactions for detection
	amount := rand.Float64() * 100
	if rand.Intn(10) == 0 {
		amount = 500 + rand.Float64()*1000 // High value
	}

	fee := rand.Float64() * 0.0001
	if rand.Intn(20) == 0 {
		fee = 0.001 + rand.Float64()*0.005 // High fee
	}

	return models.Transaction{
		ID:        randString(32),
		Chain:     models.RippleChain,
		Timestamp: time.Now(),
		From:      wallets[rand.Intn(len(wallets))],
		To:        wallets[rand.Intn(len(wallets))],
		Amount:    amount,
		Fee:       fee,
		Status:    "validated",
		Data: map[string]interface{}{
			"type":         txTypes[rand.Intn(len(txTypes))],
			"ledger_index": rand.Intn(10000000) + 70000000,
			"validated":    true,
		},
	}
}

// Helper function to generate random strings
func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
