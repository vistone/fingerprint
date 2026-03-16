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
    async getLogs(level) {
        const endpoint = level && level !== 'all' ? `/api/admin/logs?level=${level}` : '/api/admin/logs';
        return this.request(endpoint);
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

    // Agent API
    async getAgentStatus() {
        return this.request('/api/admin/agent/status');
    },

    async getAgentStrategies() {
        return this.request('/api/admin/agent/strategies');
    },

    async getAgentKnowledge() {
        return this.request('/api/admin/agent/knowledge');
    },

    async getCrawlerStatus() {
        return this.request('/api/admin/crawler/status');
    },

    async startCrawler() {
        return this.request('/api/admin/crawler/start', {
            method: 'POST',
            body: JSON.stringify({}),
        });
    },

    async crawlWithCrawler(url) {
        return this.request('/api/admin/crawler/crawl', {
            method: 'POST',
            body: JSON.stringify({ url }),
        });
    },

    async getWAFStatus() {
        return this.request('/api/admin/waf/status');
    },

    // SSE Log Stream
    subscribeToLogStream(callback) {
        const es = new EventSource('/api/admin/logs/stream');
        es.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                callback(data);
            } catch (e) { /* ignore */ }
        };
        es.onerror = () => {
            callback({ type: 'error' });
        };
        return {
            unsubscribe: () => es.close(),
        };
    },

    // ===== 分析引擎 API =====
    async analyzeProfile(profileId) {
        return this.request('/api/admin/analyze/profile', {
            method: 'POST',
            body: JSON.stringify({ profileId }),
        });
    },

    // ===== ML 引擎 API =====
    async getMLInfo() {
        return this.request('/api/admin/ml/info');
    },
    async mlExtract(profileId) {
        return this.request('/api/admin/ml/extract', {
            method: 'POST',
            body: JSON.stringify({ profileId }),
        });
    },
    async mlClassify(profileId) {
        return this.request('/api/admin/ml/classify', {
            method: 'POST',
            body: JSON.stringify({ profileId }),
        });
    },
    async mlBatch() {
        return this.request('/api/admin/ml/batch');
    },

    // ===== MLService — 中央 AI 服务 API =====
    async getMLServiceStats() {
        return this.request('/api/admin/ml/service/stats');
    },
    async getMLServiceHealth() {
        return this.request('/api/admin/ml/service/health');
    },
    async mlServiceInfer(profileId) {
        return this.request('/api/admin/ml/service/infer', {
            method: 'POST',
            body: JSON.stringify({ profileId }),
        });
    },
    async mlServiceValidate(profileId) {
        return this.request('/api/admin/ml/service/validate', {
            method: 'POST',
            body: JSON.stringify({ profileId }),
        });
    },
    async mlServiceGenerate(targetBrowser, targetOS, noiseIntensity) {
        return this.request('/api/admin/ml/service/generate', {
            method: 'POST',
            body: JSON.stringify({ targetBrowser, targetOS, noiseIntensity, maxAttempts: 10 }),
        });
    },
    async mlServiceEvolve() {
        return this.request('/api/admin/ml/service/evolve', {
            method: 'POST',
            body: JSON.stringify({}),
        });
    },
    async mlServiceTrain() {
        return this.request('/api/admin/ml/service/train', {
            method: 'POST',
            body: JSON.stringify({}),
        });
    },
    async mlServiceTrainingStatus() {
        return this.request('/api/admin/ml/service/training-status');
    },
    async mlServiceFeedback(profileId, label, reward) {
        return this.request('/api/admin/ml/service/feedback', {
            method: 'POST',
            body: JSON.stringify({ profileId, label, reward }),
        });
    },

    // ===== 防御系统 API =====
    async getDefenseRules() {
        return this.request('/api/admin/defense/rules');
    },
    async defenseDetect(profileId) {
        return this.request('/api/admin/defense/detect', {
            method: 'POST',
            body: JSON.stringify({ profileId }),
        });
    },

    // ===== 反检测引擎 API =====
    async getAntiDetectStatus() {
        return this.request('/api/admin/antidetect/status');
    },
    async antiDetectPreview(profileId, generator) {
        return this.request('/api/admin/antidetect/preview', {
            method: 'POST',
            body: JSON.stringify({ profileId, generator }),
        });
    },
    async antiDetectInject(html, profileId) {
        return this.request('/api/admin/antidetect/inject', {
            method: 'POST',
            body: JSON.stringify({ html, profileId }),
        });
    },
    async getAntiDetectSDK() {
        return this.request('/api/admin/antidetect/sdk');
    },

    // ===== 插件系统 API =====
    async getPluginsInfo() {
        return this.request('/api/admin/plugins/info');
    },

    // ===== 指纹工具 API =====
    async toolsJA3(profileId) {
        return this.request('/api/admin/tools/ja3', {
            method: 'POST',
            body: JSON.stringify({ profileId }),
        });
    },
    async toolsValidate(profileId) {
        return this.request('/api/admin/tools/validate', {
            method: 'POST',
            body: JSON.stringify({ profileId }),
        });
    },
    async toolsCompare(profileA, profileB) {
        return this.request('/api/admin/tools/compare', {
            method: 'POST',
            body: JSON.stringify({ profileA, profileB }),
        });
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
