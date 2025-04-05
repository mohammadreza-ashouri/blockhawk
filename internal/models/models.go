
// BlockHawk - Cross-Chain Security Monitor
// Author: Mohammad Reza Ashouri (@mohammadreza-ashouri)
// Paris Blockchain Week Hackathon 2025

package models

import (
	"time"
)

// BlockchainType represents the type of blockchain
type BlockchainType string

const (
	// SolanaChain represents the Solana blockchain
	SolanaChain BlockchainType = "solana"
	// RippleChain represents the Ripple (XRP) blockchain
	RippleChain BlockchainType = "ripple"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	ID        string        `json:"id"`
	Chain     BlockchainType `json:"chain"`
	Timestamp time.Time     `json:"timestamp"`
	From      string        `json:"from"`
	To        string        `json:"to"`
	Amount    float64       `json:"amount"`
	Fee       float64       `json:"fee"`
	Status    string        `json:"status"`
	Data      interface{}   `json:"data,omitempty"` // Chain-specific data
}

// AlertSeverity represents the severity level of a security alert
type AlertSeverity string

const (
	// LowSeverity represents a low severity alert
	LowSeverity AlertSeverity = "low"
	// MediumSeverity represents a medium severity alert
	MediumSeverity AlertSeverity = "medium"
	// HighSeverity represents a high severity alert
	HighSeverity AlertSeverity = "high"
)

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string        `json:"id"`
	Chain       BlockchainType `json:"chain"`
	Type        string        `json:"type"`
	Severity    AlertSeverity `json:"severity"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Timestamp   time.Time     `json:"timestamp"`
	RelatedData interface{}   `json:"related_data,omitempty"`
}

// ChainStatus represents the overall status of a blockchain
type ChainStatus struct {
	Chain            BlockchainType  `json:"chain"`
	IsConnected      bool           `json:"is_connected"`
	LastBlockTime    time.Time      `json:"last_block_time"`
	TPS              float64        `json:"tps"` // Transactions per second
	PendingTxCount   int            `json:"pending_tx_count"`
	SecurityScore    float64        `json:"security_score"` // 0-100 scale
}
