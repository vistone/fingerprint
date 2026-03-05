// Fingerprint Gateway Admin Console - Main App

// State management
const state = {
    currentPage: 'dashboard',
    profiles: [],
    stats: null,
    isLoading: false
};

// DOM Elements
document.addEventListener('DOMContentLoaded', () => {
    initNavigation();
    initDashboard();
    initProfiles();
    loadInitialData();
});

// Navigation
function initNavigation() {
    const navItems = document.querySelectorAll('.nav-item');
    const pages = document.querySelectorAll('.page');
    const pageTitle = document.getElementById('pageTitle');
    const menuToggle = document.getElementById('menuToggle');
    const sidebar = document.querySelector('.sidebar');

    navItems.forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const targetPage = item.dataset.page;
            
            // Update active states
            navItems.forEach(nav => nav.classList.remove('active'));
            item.classList.add('active');
            
            // Show target page
            pages.forEach(page => page.classList.remove('active'));
            document.getElementById(targetPage).classList.add('active');
            
            // Update title
            pageTitle.textContent = item.querySelector('span:last-child').textContent;
            
            // Update state
            state.currentPage = targetPage;
            
            // Close mobile menu
            sidebar.classList.remove('open');
            
            // Load page-specific data
            loadPageData(targetPage);
        });
    });

    // Mobile menu toggle
    menuToggle?.addEventListener('click', () => {
        sidebar.classList.toggle('open');
    });
}

// Load initial data
async function loadInitialData() {
    try {
        await Promise.all([
            loadStats(),
            loadProfiles()
        ]);
        updateDashboard();
    } catch (error) {
        console.error('Failed to load initial data:', error);
        showNotification('Failed to load data', 'error');
    }
}

// Load page-specific data
function loadPageData(page) {
    switch (page) {
        case 'dashboard':
            loadStats();
            break;
        case 'profiles':
            loadProfiles();
            break;
        case 'analytics':
            loadAnalytics();
            break;
        case 'requests':
            loadRequests();
            break;
        case 'logs':
            loadLogs();
            break;
    }
}

// Dashboard Functions
function initDashboard() {
    // Auto-refresh stats every 30 seconds
    setInterval(() => {
        if (state.currentPage === 'dashboard') {
            loadStats();
        }
    }, 30000);
}

async function loadStats() {
    try {
        const response = await fetch('/api/admin/stats');
        const data = await response.json();
        state.stats = data;
        updateDashboard();
    } catch (error) {
        console.error('Failed to load stats:', error);
    }
}

function updateDashboard() {
    if (!state.stats) return;

    // Update stat cards
    document.getElementById('totalProfiles').textContent = state.stats.totalProfiles || '-';
    document.getElementById('requestsPerSec').textContent = state.stats.requestsPerSec || '-';
    document.getElementById('avgLatency').textContent = state.stats.avgLatency ? `${state.stats.avgLatency}ms` : '-';
    document.getElementById('successRate').textContent = state.stats.successRate ? `${state.stats.successRate}%` : '-';

    // Update recent classifications
    updateRecentClassifications(state.stats.recentClassifications || []);
}

function updateRecentClassifications(classifications) {
    const tbody = document.getElementById('recentClassifications');
    
    if (classifications.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="loading">No recent classifications</td></tr>';
        return;
    }

    tbody.innerHTML = classifications.map(c => `
        <tr>
            <td>${formatTime(c.timestamp)}</td>
            <td><code>${c.ja3Hash?.substring(0, 16)}...</code></td>
            <td>${c.browser}</td>
            <td>${(c.confidence * 100).toFixed(1)}%</td>
            <td><span class="status-badge online">Success</span></td>
        </tr>
    `).join('');
}

// Profiles Functions
function initProfiles() {
    const searchInput = document.getElementById('profileSearch');
    const browserFilter = document.getElementById('browserFilter');
    const osFilter = document.getElementById('osFilter');

    searchInput?.addEventListener('input', debounce(() => {
        filterProfiles();
    }, 300));

    browserFilter?.addEventListener('change', filterProfiles);
    osFilter?.addEventListener('change', filterProfiles);
}

async function loadProfiles() {
    try {
        const response = await fetch('/api/admin/profiles');
        const data = await response.json();
        state.profiles = data.profiles || [];
        renderProfiles();
    } catch (error) {
        console.error('Failed to load profiles:', error);
    }
}

function renderProfiles() {
    const tbody = document.getElementById('profilesTableBody');
    const count = document.getElementById('profileCount');
    
    const profiles = state.profiles;
    
    if (profiles.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="loading">No profiles found</td></tr>';
        count.textContent = '0 profiles';
        return;
    }

    count.textContent = `${profiles.length} profiles`;

    tbody.innerHTML = profiles.slice(0, 20).map(p => `
        <tr>
            <td><code>${p.id}</code></td>
            <td>${p.name}</td>
            <td>${getBrowserIcon(p.browserType)} ${p.browserType}</td>
            <td>${p.browserVersion}</td>
            <td>${p.os}</td>
            <td>${formatTLSVersion(p.tlsVersion)}</td>
            <td>
                <button class="btn btn-sm" onclick="viewProfile('${p.id}')">View</button>
            </td>
        </tr>
    `).join('');
}

function filterProfiles() {
    const search = document.getElementById('profileSearch')?.value.toLowerCase() || '';
    const browser = document.getElementById('browserFilter')?.value || '';
    const os = document.getElementById('osFilter')?.value || '';

    let filtered = state.profiles;

    if (search) {
        filtered = filtered.filter(p => 
            p.name?.toLowerCase().includes(search) ||
            p.id?.toLowerCase().includes(search)
        );
    }

    if (browser) {
        filtered = filtered.filter(p => 
            p.browserType?.toLowerCase() === browser
        );
    }

    if (os) {
        filtered = filtered.filter(p => 
            p.os?.toLowerCase().includes(os)
        );
    }

    renderFilteredProfiles(filtered);
}

function renderFilteredProfiles(profiles) {
    const tbody = document.getElementById('profilesTableBody');
    const count = document.getElementById('profileCount');

    count.textContent = `${profiles.length} profiles`;

    if (profiles.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="loading">No matching profiles</td></tr>';
        return;
    }

    tbody.innerHTML = profiles.slice(0, 20).map(p => `
        <tr>
            <td><code>${p.id}</code></td>
            <td>${p.name}</td>
            <td>${getBrowserIcon(p.browserType)} ${p.browserType}</td>
            <td>${p.browserVersion}</td>
            <td>${p.os}</td>
            <td>${formatTLSVersion(p.tlsVersion)}</td>
            <td>
                <button class="btn btn-sm" onclick="viewProfile('${p.id}')">View</button>
            </td>
        </tr>
    `).join('');
}

// Analytics Functions
async function loadAnalytics() {
    try {
        const response = await fetch('/api/admin/analytics');
        const data = await response.json();
        updateAnalytics(data);
    } catch (error) {
        console.error('Failed to load analytics:', error);
    }
}

function updateAnalytics(data) {
    // Update browser chart
    if (data.browserDistribution) {
        renderBrowserChart(data.browserDistribution);
    }

    // Update OS chart
    if (data.osDistribution) {
        renderOSChart(data.osDistribution);
    }

    // Update top fingerprints
    if (data.topFingerprints) {
        const tbody = document.getElementById('topFingerprints');
        tbody.innerHTML = data.topFingerprints.map(f => `
            <tr>
                <td><code>${f.hash.substring(0, 20)}...</code></td>
                <td>${f.count.toLocaleString()}</td>
                <td>${f.percentage}%</td>
            </tr>
        `).join('');
    }
}

// Requests Functions
async function loadRequests() {
    try {
        const response = await fetch('/api/admin/requests');
        const data = await response.json();
        updateRequests(data.requests || []);
    } catch (error) {
        console.error('Failed to load requests:', error);
    }
}

function updateRequests(requests) {
    const tbody = document.getElementById('requestsTable');
    
    if (requests.length === 0) {
        tbody.innerHTML = '<tr><td colspan="8" class="loading">No requests found</td></tr>';
        return;
    }

    tbody.innerHTML = requests.map(r => `
        <tr>
            <td>${formatTime(r.timestamp)}</td>
            <td>${r.ip}</td>
            <td>${r.method}</td>
            <td>${r.path}</td>
            <td><code>${r.ja3?.substring(0, 16)}...</code></td>
            <td>${r.classification || '-'}</td>
            <td>${r.latency}ms</td>
            <td><span class="status-badge ${r.status < 400 ? 'online' : 'offline'}">${r.status}</span></td>
        </tr>
    `).join('');
}

// Logs Functions
async function loadLogs() {
    try {
        const response = await fetch('/api/admin/logs');
        const data = await response.json();
        document.getElementById('logOutput').textContent = data.logs || 'No logs available';
    } catch (error) {
        console.error('Failed to load logs:', error);
        document.getElementById('logOutput').textContent = 'Failed to load logs';
    }
}

// Utility Functions
function formatTime(timestamp) {
    if (!timestamp) return '-';
    const date = new Date(timestamp);
    return date.toLocaleString();
}

function formatTLSVersion(version) {
    if (!version) return '-';
    const versions = {
        0x0301: 'TLS 1.0',
        0x0302: 'TLS 1.1',
        0x0303: 'TLS 1.2',
        0x0304: 'TLS 1.3'
    };
    return versions[version] || `0x${version.toString(16)}`;
}

function getBrowserIcon(browserType) {
    const icons = {
        'chrome': '🌐',
        'firefox': '🦊',
        'safari': '🧭',
        'edge': '🌊',
        'opera': '⚪'
    };
    return icons[browserType?.toLowerCase()] || '🌐';
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

function showNotification(message, type = 'info') {
    // Simple notification - can be enhanced
    console.log(`[${type.toUpperCase()}] ${message}`);
}

// Global functions
window.viewProfile = function(id) {
    console.log('View profile:', id);
    // Implement profile detail view
};

// Charts (placeholder - would integrate with Chart.js)
function renderBrowserChart(data) {
    const container = document.getElementById('browserChart');
    if (!container) return;
    
    // Placeholder chart rendering
    container.innerHTML = `
        <div style="display: flex; flex-direction: column; gap: 8px; width: 100%;">
            ${Object.entries(data).map(([browser, count]) => `
                <div style="display: flex; align-items: center; gap: 8px;">
                    <span style="width: 80px;">${browser}</span>
                    <div style="flex: 1; background: #e5e7eb; height: 20px; border-radius: 4px; overflow: hidden;">
                        <div style="width: ${count}%; background: var(--primary); height: 100%;"></div>
                    </div>
                    <span style="width: 40px; text-align: right;">${count}%</span>
                </div>
            `).join('')}
        </div>
    `;
}

function renderOSChart(data) {
    const container = document.getElementById('osChart');
    if (!container) return;
    
    // Placeholder pie chart
    const total = Object.values(data).reduce((a, b) => a + b, 0);
    let currentAngle = 0;
    
    const colors = ['#4f46e5', '#10b981', '#f59e0b', '#ef4444', '#3b82f6'];
    
    container.innerHTML = `
        <div style="display: flex; align-items: center; gap: 20px;">
            <div style="width: 150px; height: 150px; border-radius: 50%; background: conic-gradient(
                ${Object.entries(data).map(([os, count], i) => {
                    const angle = (count / total) * 360;
                    const gradient = `${colors[i % colors.length]} ${currentAngle}deg ${currentAngle + angle}deg`;
                    currentAngle += angle;
                    return gradient;
                }).join(', ')}
            );"></div>
            <div style="display: flex; flex-direction: column; gap: 4px;">
                ${Object.entries(data).map(([os, count], i) => `
                    <div style="display: flex; align-items: center; gap: 8px;">
                        <div style="width: 12px; height: 12px; background: ${colors[i % colors.length]}; border-radius: 2px;"></div>
                        <span>${os}: ${count}%</span>
                    </div>
                `).join('')}
            </div>
        </div>
    `;
}
