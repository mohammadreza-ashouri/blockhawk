# BlockHawk

This is a real-time security monitoring dashboard for Solana and Ripple blockchains, built with Go. Written by [@mohammadreza-ashouri](https://github.com/mohammadreza-ashouri).

This is my submission for the [PBW 2025 Paris](https://www.parisblockchainweek.com/hackathon-2025) challenge.


Discord: [@mo9990871](https://discord.gg/mo9990871)

Email: [ashouri7774@gmail.com](mailto:ashouri7774@gmail.com)

X: [@ashouri777](https://x.com/ashouri777)


## Introduction

BlockHawk is a real-time security monitoring dashboard for Solana and Ripple blockchains, built with Go. It provides a web-based interface to visualize blockchain data and detect anomalies.

## Features

- Concurrent monitoring of Solana and Ripple blockchains
- Anomaly detection for suspicious transactions and activities
- Real-time visualization dashboard
- Alert system for potential security threats

1. **Unusual Transaction Patterns**
   - High-value transactions exceeding normal thresholds
   - Rapid succession of transactions from same address
   - Abnormal transaction fee structures
   - Transaction timing anomalies

2. **Network Anomalies**
   - Unusual network load/TPS spikes
   - Degradation in transaction confirmation times
   - Validator/node participation changes
   - Consensus timing irregularities

3. **Smart Contract Vulnerabilities**
   - Function call patterns indicating exploitation attempts
   - Resource exhaustion attacks (on Solana programs)
   - Repeated failed contract calls

4. **Wallet Security Issues**
   - Unusual withdrawal patterns
   - Suspicious cross-chain transfers
   - New wallet addresses receiving large sums

## AI/ML Detection Methods

BlockHawk implements several lightweight AI/ML techniques:

1. **Statistical Anomaly Detection**
   - Z-score analysis on transaction volumes and fees
   - Moving average calculations with adaptive thresholds
   - Time-series analysis for pattern recognition

2. **Clustering Algorithms**
   - K-means clustering for transaction categorization
   - Identifying outlier transactions through distance metrics

3. **Rule-Based Systems with Adaptive Thresholds**
   - Dynamically adjusted thresholds based on historical data
   - Chain-specific rule sets for different security concerns

4. **Simple Prediction Models**
   - Linear regression for trend analysis
   - Basic forecasting to establish expected vs. actual behavior

## Tech Stack
- Go for backend and blockchain integration
- WebSockets for real-time updates
- Simple AI/ML for anomaly detection
- Web-based dashboard with D3.js charts

## Installation

```bash
# Clone the repository
git clone https://github.com/mohammadreza-ashouri/blockhawk.git
cd blockhawk

# Install dependencies
go mod tidy

# Run the server
go run cmd/server/main.go