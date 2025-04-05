// BlockHawk - Cross-Chain Security Monitor
// Author: Mohammad Reza Ashouri (@mohammadreza-ashouri)
// Paris Blockchain Week Hackathon 2025

package analyzer

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mohammadreza-ashouri/blockhawk/internal/models"
)

// Analyzer implements security analysis features
type Analyzer struct {
	txBuffer        map[models.BlockchainType][]models.Transaction
	txSenderHistory map[string][]time.Time // Track sender transaction times
	alertChan       chan models.SecurityAlert
	bufferMutex     sync.RWMutex
	historyMutex    sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewAnalyzer creates a new security analyzer
func NewAnalyzer() *Analyzer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Analyzer{
		txBuffer:        make(map[models.BlockchainType][]models.Transaction),
		txSenderHistory: make(map[string][]time.Time),
		alertChan:       make(chan models.SecurityAlert, 100),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start begins the analysis process
func (a *Analyzer) Start() {
	log.Println("Starting security analyzer...")

	// Initialize buffers for each chain
	a.bufferMutex.Lock()
	a.txBuffer[models.SolanaChain] = make([]models.Transaction, 0, 100)
	a.txBuffer[models.RippleChain] = make([]models.Transaction, 0, 100)
	a.bufferMutex.Unlock()

	// Start cleanup goroutine
	go a.periodicCleanup()
}

// Stop halts all analysis activities
func (a *Analyzer) Stop() {
	log.Println("Stopping security analyzer...")
	a.cancel()
	close(a.alertChan)
}

// GetAlertChannel returns the channel for receiving security alerts
func (a *Analyzer) GetAlertChannel() <-chan models.SecurityAlert {
	return a.alertChan
}

// ProcessTransaction analyzes a transaction for security issues
func (a *Analyzer) ProcessTransaction(tx models.Transaction) {
	// Add transaction to buffer
	a.bufferMutex.Lock()
	a.txBuffer[tx.Chain] = append(a.txBuffer[tx.Chain], tx)

	// Keep buffer at reasonable size (last 100 transactions)
	if len(a.txBuffer[tx.Chain]) > 100 {
		a.txBuffer[tx.Chain] = a.txBuffer[tx.Chain][len(a.txBuffer[tx.Chain])-100:]
	}
	a.bufferMutex.Unlock()

	// Check for immediate security concerns
	a.detectLargeTransactions(tx)
	a.detectHighFees(tx)

	// Update transaction history for sender
	a.updateSenderHistory(tx)

	// Check for rapid transaction sequences
	a.detectRapidTransactions(tx)
}

// createAlert creates and sends a new security alert
func (a *Analyzer) createAlert(alert models.SecurityAlert) {
	select {
	case a.alertChan <- alert:
		// Alert sent successfully
		log.Printf("Security alert generated: %s (%s)", alert.Title, alert.Severity)
	default:
		// Channel full, logging alert
		log.Printf("Alert channel full, dropping alert: %s", alert.Title)
	}
}

// detectLargeTransactions identifies unusually large transactions
func (a *Analyzer) detectLargeTransactions(tx models.Transaction) {
	var threshold float64

	// Set threshold based on blockchain
	switch tx.Chain {
	case models.SolanaChain:
		threshold = 20.0 // 20 SOL
	case models.RippleChain:
		threshold = 200.0 // 200 XRP
	default:
		return
	}

	if tx.Amount > threshold {
		// Calculate how many times above threshold
		magnitude := tx.Amount / threshold

		// Determine severity based on magnitude
		severity := models.LowSeverity
		if magnitude > 5 {
			severity = models.MediumSeverity
		}
		if magnitude > 20 {
			severity = models.HighSeverity
		}

		a.createAlert(models.SecurityAlert{
			ID:       uuid.New().String(),
			Chain:    tx.Chain,
			Type:     "large_transaction",
			Severity: severity,
			Title:    "Large Transaction Detected",
			Description: fmt.Sprintf("Transaction of %.4f on %s chain exceeds normal threshold by %.1fx",
				tx.Amount, tx.Chain, magnitude),
			Timestamp:   time.Now(),
			RelatedData: tx,
		})
	}
}

// detectHighFees identifies transactions with unusually high fees
func (a *Analyzer) detectHighFees(tx models.Transaction) {
	var threshold float64

	// Set threshold based on blockchain
	switch tx.Chain {
	case models.SolanaChain:
		threshold = 0.01 // 0.01 SOL
	case models.RippleChain:
		threshold = 0.001 // 0.001 XRP
	default:
		return
	}

	if tx.Fee > threshold {
		a.createAlert(models.SecurityAlert{
			ID:       uuid.New().String(),
			Chain:    tx.Chain,
			Type:     "high_fee",
			Severity: models.MediumSeverity,
			Title:    "Abnormally High Fee Detected",
			Description: fmt.Sprintf("Transaction fee of %.6f on %s chain is unusually high",
				tx.Fee, tx.Chain),
			Timestamp:   time.Now(),
			RelatedData: tx,
		})
	}
}

// updateSenderHistory tracks transaction times for senders
func (a *Analyzer) updateSenderHistory(tx models.Transaction) {
	key := fmt.Sprintf("%s-%s", tx.Chain, tx.From)

	a.historyMutex.Lock()
	defer a.historyMutex.Unlock()

	// Initialize if not exists
	if _, exists := a.txSenderHistory[key]; !exists {
		a.txSenderHistory[key] = make([]time.Time, 0, 10)
	}

	// Add current timestamp
	a.txSenderHistory[key] = append(a.txSenderHistory[key], tx.Timestamp)

	// Keep only last 10 transactions
	if len(a.txSenderHistory[key]) > 10 {
		a.txSenderHistory[key] = a.txSenderHistory[key][len(a.txSenderHistory[key])-10:]
	}
}

// detectRapidTransactions identifies unusually rapid transaction sequences
func (a *Analyzer) detectRapidTransactions(tx models.Transaction) {
	key := fmt.Sprintf("%s-%s", tx.Chain, tx.From)

	a.historyMutex.RLock()
	defer a.historyMutex.RUnlock()

	history, exists := a.txSenderHistory[key]
	if !exists || len(history) < 5 {
		return // Not enough history
	}

	// Check for 5+ transactions in 1 minute
	count := 0
	oneMinuteAgo := tx.Timestamp.Add(-1 * time.Minute)

	for _, txTime := range history {
		if txTime.After(oneMinuteAgo) {
			count++
		}
	}

	if count >= 5 {
		a.createAlert(models.SecurityAlert{
			ID:       uuid.New().String(),
			Chain:    tx.Chain,
			Type:     "rapid_transactions",
			Severity: models.MediumSeverity,
			Title:    "Rapid Transaction Sequence",
			Description: fmt.Sprintf("Address %s performed %d transactions in 1 minute",
				tx.From, count),
			Timestamp:   time.Now(),
			RelatedData: tx.From,
		})
	}
}

// calculateMovingAverage computes simple moving average for a slice of values
func (a *Analyzer) calculateMovingAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStandardDeviation computes standard deviation for a slice of values
func (a *Analyzer) calculateStandardDeviation(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := a.calculateMovingAverage(values)
	sumSquaredDiffs := 0.0

	for _, v := range values {
		diff := v - mean
		sumSquaredDiffs += diff * diff
	}

	return math.Sqrt(sumSquaredDiffs / float64(len(values)))
}

// periodicCleanup removes old data from memory
func (a *Analyzer) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.cleanupOldData()
		}
	}
}

// cleanupOldData removes old transaction history
func (a *Analyzer) cleanupOldData() {
	twoHoursAgo := time.Now().Add(-2 * time.Hour)

	a.historyMutex.Lock()
	defer a.historyMutex.Unlock()

	// Remove transactions older than 2 hours
	for key, times := range a.txSenderHistory {
		newTimes := make([]time.Time, 0, len(times))

		for _, t := range times {
			if t.After(twoHoursAgo) {
				newTimes = append(newTimes, t)
			}
		}

		if len(newTimes) > 0 {
			a.txSenderHistory[key] = newTimes
		} else {
			delete(a.txSenderHistory, key)
		}
	}
}
