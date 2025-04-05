// BlockHawk - Cross-Chain Security Monitor
// Author: Mohammad Reza Ashouri (@mohammadreza-ashouri)
// Paris Blockchain Week Hackathon 2025

package solana

import (
	"context"
	"log"
	"math/rand" // For demo purposes
	"time"

	"github.com/mohammadreza-ashouri/blockhawk/internal/models"
)

// Monitor represents a Solana blockchain monitor
type Monitor struct {
	txChan     chan models.Transaction
	statusChan chan models.ChainStatus
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewMonitor creates a new Solana monitor
func NewMonitor() *Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		txChan:     make(chan models.Transaction, 100),
		statusChan: make(chan models.ChainStatus, 10),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins monitoring the Solana blockchain
func (m *Monitor) Start() {
	log.Println("Starting Solana blockchain monitor...")

	// Start transaction monitoring goroutine
	go m.monitorTransactions()

	// Start status updates goroutine
	go m.updateStatus()
}

// Stop halts all monitoring activities
func (m *Monitor) Stop() {
	log.Println("Stopping Solana blockchain monitor...")
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

// monitorTransactions monitors transactions on the Solana blockchain
// In a real implementation, this would connect to Solana RPC
func (m *Monitor) monitorTransactions() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// For hackathon purposes, generate some mock transactions
			tx := generateMockSolanaTransaction()

			select {
			case m.txChan <- tx:
				// Transaction sent successfully
			default:
				// Channel full, logging and continuing
				log.Println("Solana transaction channel full, dropping transaction")
			}
		}
	}
}

// updateStatus updates the overall Solana blockchain status
func (m *Monitor) updateStatus() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			status := models.ChainStatus{
				Chain:          models.SolanaChain,
				IsConnected:    true,
				LastBlockTime:  time.Now().Add(-time.Duration(rand.Intn(5)) * time.Second),
				TPS:            400 + rand.Float64()*100,
				PendingTxCount: rand.Intn(1000),
				SecurityScore:  70 + rand.Float64()*30,
			}

			select {
			case m.statusChan <- status:
				// Status sent successfully
			default:
				// Channel full, logging and continuing
				log.Println("Solana status channel full, dropping status update")
			}
		}
	}
}

// Helper function to generate mock Solana transactions
func generateMockSolanaTransaction() models.Transaction {
	wallets := []string{
		"4Qkev8aNZcqFNSRhQzwyLMFSsi94jHqE8WNVTJzTP99F",
		"vines1vzrYbzLMRdu58ou5XTby4qAqVRLmqo36NKPTg",
		"GpYnVDj3dmN9ZP8tWManpT5pYpJBQzWuuhyWmziNZoFJ",
		"2q7pyhPwAwZ3QMfZrnAbDhnh9mDUqycszcpf8bnT7i8J",
		"71bhKKL89U7Snh2SfkTna3nxmNYGVBTuzyYPQzTPSgQ3",
	}

	txTypes := []string{"transfer", "swap", "stake", "unstake", "create_account"}

	// Occasionally generate high-value or high-fee transactions for detection
	amount := rand.Float64() * 10
	if rand.Intn(10) == 0 {
		amount = 50 + rand.Float64()*100 // High value
	}

	fee := rand.Float64() * 0.001
	if rand.Intn(20) == 0 {
		fee = 0.01 + rand.Float64()*0.05 // High fee
	}

	return models.Transaction{
		ID:        randString(32),
		Chain:     models.SolanaChain,
		Timestamp: time.Now(),
		From:      wallets[rand.Intn(len(wallets))],
		To:        wallets[rand.Intn(len(wallets))],
		Amount:    amount,
		Fee:       fee,
		Status:    "confirmed",
		Data: map[string]interface{}{
			"type":          txTypes[rand.Intn(len(txTypes))],
			"compute_units": rand.Intn(200000),
			"slot":          rand.Intn(1000000),
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
