// Fingerprint Gateway Admin Console - API Client

const API = {
    baseURL: '',
    
    // Set base URL for API requests
    setBaseURL(url) {
        this.baseURL = url;
    },
    
    // Generic request method
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
            },
        };
        
        const response = await fetch(url, { ...defaultOptions, ...options });
        
        if (!response.ok) {
            throw new Error(`API error: ${response.status} ${response.statusText}`);
        }
        
        return response.json();
    },
    
    // Stats API
    async getStats() {
        return this.request('/api/admin/stats');
    },
    
    // Profiles API
    async getProfiles(filters = {}) {
        const params = new URLSearchParams();
        
        if (filters.q) params.append('q', filters.q);
        if (filters.browser) params.append('browser', filters.browser);
        if (filters.os) params.append('os', filters.os);
        
        const query = params.toString();
        const endpoint = query ? `/api/admin/profiles?${query}` : '/api/admin/profiles';
        
        return this.request(endpoint);
    },
    
    // Analytics API
    async getAnalytics() {
        return this.request('/api/admin/analytics');
    },
    
    // Requests API
    async getRequests() {
        return this.request('/api/admin/requests');
    },
    
    // Logs API
    async getLogs() {
        return this.request('/api/admin/logs');
    },
    
    // Config API
    async getConfig() {
        return this.request('/api/admin/config');
    },
    
    async updateConfig(config) {
        return this.request('/api/admin/config', {
            method: 'POST',
            body: JSON.stringify(config),
        });
    },
    
    // Real-time updates
    subscribeToUpdates(callback) {
        // WebSocket connection for real-time updates
        // Placeholder implementation
        const ws = new WebSocket(`ws://${window.location.host}/api/admin/ws`);
        
        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            callback(data);
        };
        
        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
        
        return {
            unsubscribe: () => ws.close(),
        };
    },
};

// Auto-initialize API base URL
if (typeof window !== 'undefined') {
    API.setBaseURL('');
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { API };
}
