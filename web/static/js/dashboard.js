document.addEventListener('DOMContentLoaded', function() {
    // Connect to WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    const socket = new WebSocket(wsUrl);
    
    // WebSocket event handlers
    socket.onopen = function() {
        console.log('WebSocket connection established');
    };
    
    socket.onclose = function() {
        console.log('WebSocket connection closed');
        // Try to reconnect after 3 seconds
        setTimeout(function() {
            window.location.reload();
        }, 3000);
    };
    
    socket.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
    
    socket.onmessage = function(event) {
        const data = JSON.parse(event.data);
        
        if (data.type === 'initial' || data.type === 'update') {
            // Update blockchain status
            updateBlockchainStatus(data.status);
            
            // Update alerts
            if (data.alerts && data.alerts.length > 0) {
                updateAlerts(data.alerts);
            }
        }
    };
    
    // Fetch initial data via REST API in case WebSocket is slow
    fetchInitialData();
});

// Fetch initial data from REST APIs
function fetchInitialData() {
    // Fetch blockchain status
    fetch('/api/status')
        .then(response => response.json())
        .then(data => {
            updateBlockchainStatus(data);
        })
        .catch(error => {
            console.error('Error fetching blockchain status:', error);
        });
    
    // Fetch alerts
    fetch('/api/alerts')
        .then(response => response.json())
        .then(data => {
            updateAlerts(data);
        })
        .catch(error => {
            console.error('Error fetching alerts:', error);
        });
}

// Update blockchain status display
function updateBlockchainStatus(statusData) {
    if (!statusData) return;
    
    // Update Solana status
    if (statusData.solana) {
        const solanaStatus = statusData.solana;
        const solanaEl = document.getElementById('solana-status');
        
        solanaEl.innerHTML = createStatusHTML(solanaStatus);
    }
    
    // Update Ripple status
    if (statusData.ripple) {
        const rippleStatus = statusData.ripple;
        const rippleEl = document.getElementById('ripple-status');
        
        rippleEl.innerHTML = createStatusHTML(rippleStatus);
    }
}

// Create HTML for status card
function createStatusHTML(status) {
    const securityScoreClass = getSecurityScoreClass(status.security_score);
    const connectedStatus = status.is_connected ? 
        '<span class="text-success">Connected</span>' : 
        '<span class="text-danger">Disconnected</span>';
    
    const lastBlockTime = new Date(status.last_block_time).toLocaleTimeString();
    
    return `
        <div class="text-center mb-3">
            <div class="security-score ${securityScoreClass}">${Math.round(status.security_score)}</div>
            <div class="text-muted">Security Score</div>
        </div>
        
        <div class="status-item d-flex justify-content-between">
            <span class="status-label">Connection:</span>
            <span>${connectedStatus}</span>
        </div>
        
        <div class="status-item d-flex justify-content-between">
            <span class="status-label">Last Block:</span>
            <span>${lastBlockTime}</span>
        </div>
        
        <div class="status-item d-flex justify-content-between">
            <span class="status-label">TPS:</span>
            <span>${Math.round(status.tps)}</span>
        </div>
        
        <div class="status-item d-flex justify-content-between">
            <span class="status-label">Pending Tx:</span>
            <span>${status.pending_tx_count}</span>
        </div>
    `;
}

// Get class for security score
function getSecurityScoreClass(score) {
    if (score >= 80) return 'text-success';
    if (score >= 60) return 'text-warning';
    return 'text-danger';
}

// Update alerts table
function updateAlerts(alerts) {
    if (!alerts || alerts.length === 0) return;
    
    const tableBody = document.getElementById('alerts-table');
    
    // Clear loading indicator if present
    if (tableBody.innerHTML.includes('spinner-border')) {
        tableBody.innerHTML = '';
    }
    
    // Add new alerts at the top
    alerts.forEach(alert => {
        // Check if alert already exists
        const existingRow = document.getElementById(`alert-${alert.id}`);
        if (existingRow) return;
        
        const row = document.createElement('tr');
        row.id = `alert-${alert.id}`;
        row.className = getSeverityClass(alert.severity);
        
        const time = new Date(alert.timestamp).toLocaleTimeString();
        
        row.innerHTML = `
            <td>${time}</td>
            <td>${alert.chain}</td>
            <td>${alert.severity}</td>
            <td>${alert.title}</td>
            <td>${alert.description}</td>
        `;
        
        tableBody.insertBefore(row, tableBody.firstChild);
    });
    
    // Limit to 20 rows for performance
    while (tableBody.childElementCount > 20) {
        tableBody.removeChild(tableBody.lastChild);
    }
}

// Get CSS class for severity
function getSeverityClass(severity) {
    switch (severity) {
        case 'high': return 'severity-high';
        case 'medium': return 'severity-medium';
        case 'low': return 'severity-low';
        default: return '';
    }
}
