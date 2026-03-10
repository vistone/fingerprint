// Fingerprint Gateway Admin Console - Main App

// Initialize i18n when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Apply translations to static HTML elements
    if (typeof I18N !== 'undefined') {
        I18N.applyToDOM();
    }
});

// Helper function for table cell rendering with i18n-aware text
function td(text) {
    return `<td>${text}</td>`;
}

function th(i18nKey) {
    return `<th>${t(i18nKey)}</th>`;
}

// Utility Functions
function formatHeaderName(key) {
    // 将驼峰命名转换为 HTTP Header 格式
    // 例如: AcceptLanguage -> Accept-Language
    const headerMap = {
        'Accept': 'Accept',
        'AcceptLanguage': 'Accept-Language',
        'AcceptEncoding': 'Accept-Encoding',
        'UserAgent': 'User-Agent',
        'SecFetchSite': 'Sec-Fetch-Site',
        'SecFetchMode': 'Sec-Fetch-Mode',
        'SecFetchUser': 'Sec-Fetch-User',
        'SecFetchDest': 'Sec-Fetch-Dest',
        'SecCHUA': 'Sec-CH-UA',
        'SecCHUAMobile': 'Sec-CH-UA-Mobile',
        'SecCHUAPlatform': 'Sec-CH-UA-Platform',
        'UpgradeInsecureRequests': 'Upgrade-Insecure-Requests'
    };
    return headerMap[key] || key;
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// State management
const state = {
    currentPage: 'dashboard',
    profiles: [],
    filteredProfiles: null,
    stats: null,
    isLoading: false,
    profilesPage: 1,
    profilesPerPage: 20,
    totalProfiles: 0
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
        case 'agent':
            loadAgentData();
            break;
        case 'config':
            loadConfig();
            break;
        case 'client-test':
            loadClientTestProfiles();
            break;
        case 'analyze':
            loadAnalyzePage();
            break;
        case 'ml-engine':
            loadMLPage();
            break;
        case 'defense':
            loadDefensePage();
            break;
        case 'antidetect':
            loadAntiDetectPage();
            break;
        case 'plugins':
            loadPluginsPage();
            break;
        case 'tools':
            loadToolsPage();
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
    document.getElementById('requestsPerSec').textContent = state.stats.requestsPerSec ?? '-';
    document.getElementById('avgLatency').textContent = state.stats.avgLatency ? `${state.stats.avgLatency}ms` : '-';
    document.getElementById('successRate').textContent = state.stats.successRate ? `${state.stats.successRate.toFixed(1)}%` : '-';

    // Update uptime
    const uptimeEl = document.getElementById('uptime');
    if (uptimeEl && state.stats.uptime) {
        uptimeEl.textContent = `uptime: ${state.stats.uptime}`;
    }

    // Update recent classifications
    updateRecentClassifications(state.stats.recentClassifications || []);

    // Update system status dynamically
    updateSystemStatus(state.stats);
}

function updateRecentClassifications(classifications) {
    const tbody = document.getElementById('recentClassifications');

    if (classifications.length === 0) {
        tbody.innerHTML = `<tr><td colspan="5" class="loading">${t('loading')}</td></tr>`;
        return;
    }

    tbody.innerHTML = classifications.map(c => `
        <tr>
            <td>${formatTime(c.timestamp)}</td>
            <td><code>${c.ja3Hash?.substring(0, 16)}...</code></td>
            <td>${c.browser}</td>
            <td>${(c.confidence * 100).toFixed(1)}%</td>
            <td><span class="status-badge online">${t('status.running')}</span></td>
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
    // Clear filtered profiles when rendering all profiles
    state.filteredProfiles = null;
    renderProfileList(state.profiles);
}

function renderProfileList(profiles) {
    const tbody = document.getElementById('profilesTableBody');
    const count = document.getElementById('profileCount');

    if (profiles.length === 0) {
        tbody.innerHTML = `<tr><td colspan="10" class="loading">${t('profiles.noProfiles')}</td></tr>`;
        count.textContent = t('profiles.count', 0);
        return;
    }

    state.totalProfiles = profiles.length;
    const start = (state.profilesPage - 1) * state.profilesPerPage;
    const end = start + state.profilesPerPage;
    const pageProfiles = profiles.slice(start, end);

    count.textContent = `${profiles.length} ${t('profiles.count', profiles.length).split('{n} ')[1]}`;

    tbody.innerHTML = pageProfiles.map(p => `
        <tr>
            <td><code>${p.id}</code></td>
            <td>${p.name}</td>
            <td>${getBrowserIcon(p.browserType)} ${p.browserType}</td>
            <td>${p.browserVersion}</td>
            <td>${p.os}</td>
            <td>${formatTLSVersion(p.tlsVersion)}</td>
            <td><span class="badge">${p.cipherSuites || 0}</span></td>
            <td><span class="badge">${p.extensions || 0}</span></td>
            <td>${p.tcpip ? `<span class="badge badge-success" title="TTL:${p.tcpip.ttl}, Win:${p.tcpip.windowSize}">✓</span>` : '<span class="badge">-</span>'}</td>
            <td>
                <button class="btn btn-sm" onclick="viewProfile('${p.id}')">${t('profiles.view')}</button>
            </td>
        </tr>
    `).join('');

    updatePaginationButtons();
}

function updatePaginationButtons() {
    const prevBtn = document.querySelector('.page-buttons button:first-child');
    const nextBtn = document.querySelector('.page-buttons button:last-child');

    const profiles = state.filteredProfiles || state.profiles;
    const maxPage = Math.ceil(profiles.length / state.profilesPerPage);

    if (prevBtn) {
        prevBtn.disabled = state.profilesPage <= 1;
        prevBtn.onclick = () => {
            if (state.profilesPage > 1) {
                state.profilesPage--;
                if (state.filteredProfiles) {
                    renderProfileList(state.filteredProfiles);
                } else {
                    renderProfileList(state.profiles);
                }
            }
        };
    }

    if (nextBtn) {
        nextBtn.disabled = state.profilesPage >= maxPage;
        nextBtn.onclick = () => {
            if (state.profilesPage < maxPage) {
                state.profilesPage++;
                if (state.filteredProfiles) {
                    renderProfileList(state.filteredProfiles);
                } else {
                    renderProfileList(state.profiles);
                }
            }
        };
    }
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

    state.filteredProfiles = filtered;
    state.profilesPage = 1; // Reset to first page when filtering
    renderFilteredProfiles();
}

function renderFilteredProfiles() {
    renderProfileList(state.filteredProfiles || state.profiles);
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

    // Update traffic chart
    if (data.trafficData) {
        renderTrafficChart(data.trafficData);
    }

    // Update TCP/IP chart
    if (data.tcpipDistribution) {
        renderTCPIPChart(data.tcpipDistribution);
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
        tbody.innerHTML = `<tr><td colspan="8" class="loading">${t('loading')}</td></tr>`;
        return;
    }

    tbody.innerHTML = requests.map(r => `
        <tr>
            <td>${formatTime(r.timestamp)}</td>
            <td>${r.ip}</td>
            <td><span class="method-badge ${r.method}">${r.method}</span></td>
            <td>${r.path}</td>
            <td>${r.ja3 ? `<code>${r.ja3.substring(0, 16)}...</code>` : '-'}</td>
            <td>${r.classification || '-'}</td>
            <td>${r.latency > 0 ? r.latency + 'ms' : '<1ms'}</td>
            <td><span class="status-badge ${r.status < 400 ? 'online' : 'offline'}">${r.status}</span></td>
        </tr>
    `).join('');
}

// Logs Functions
// Log state
let logStreamSource = null;
let logEntries = [];

async function loadLogs() {
    try {
        const level = document.getElementById('logLevelFilter')?.value || 'all';
        const url = level && level !== 'all' ? `/api/admin/logs?level=${level}` : '/api/admin/logs';
        const response = await fetch(url);
        const data = await response.json();
        logEntries = data.logs || [];
        renderLogs();
    } catch (error) {
        console.error('Failed to load logs:', error);
        document.getElementById('logOutput').textContent = 'Failed to load logs';
    }
}

function renderLogs() {
    const output = document.getElementById('logOutput');
    if (!output) return;

    if (logEntries.length === 0) {
        output.innerHTML = `<span style="color:var(--gray-500);">${t('loading.logs')}</span>`;
        return;
    }

    const levelColors = {
        'ERROR': '#ef4444', 'WARN': '#f59e0b',
        'INFO': '#3b82f6', 'DEBUG': '#6b7280'
    };

    output.innerHTML = logEntries.map(e => {
        const ts = e.timestamp ? new Date(e.timestamp).toLocaleString() : '';
        const color = levelColors[e.level] || '#6b7280';
        const src = e.source ? `[${escapeHtml(e.source)}]` : '';
        return `<span style="color:${color}">[${ts}] ${escapeHtml(e.level)} ${src} ${escapeHtml(e.message)}</span>`;
    }).join('\n');

    document.getElementById('logCount').textContent = t('logs.entries', logEntries.length);

    // Auto-scroll
    if (document.getElementById('logAutoScroll')?.checked) {
        output.scrollTop = output.scrollHeight;
    }
}

function filterLogs() {
    loadLogs();
}

function clearLogs() {
    logEntries = [];
    renderLogs();
}

function toggleLogStream() {
    const toggle = document.getElementById('logStreamToggle');
    if (toggle.checked) {
        startLogStream();
    } else {
        stopLogStream();
    }
}

function startLogStream() {
    if (logStreamSource) return;
    logStreamSource = new EventSource('/api/admin/logs/stream');
    logStreamSource.onopen = () => {
        document.getElementById('logStreamStatus').textContent = t('logs.streamConnected');
    };
    logStreamSource.onmessage = (event) => {
        try {
            const entry = JSON.parse(event.data);
            if (entry.type === 'connected') return;
            logEntries.push(entry);
            // Keep max 500 entries in UI
            if (logEntries.length > 500) logEntries = logEntries.slice(-500);
            renderLogs();
        } catch (e) { /* ignore parse errors */ }
    };
    logStreamSource.onerror = () => {
        document.getElementById('logStreamStatus').textContent = t('logs.streamDisconnected');
        stopLogStream();
        document.getElementById('logStreamToggle').checked = false;
    };
}

function stopLogStream() {
    if (logStreamSource) {
        logStreamSource.close();
        logStreamSource = null;
    }
    document.getElementById('logStreamStatus').textContent = t('logs.streamDisconnected');
}

// ==================== System Status ====================
function updateSystemStatus(stats) {
    const list = document.getElementById('systemStatusList');
    if (!list) return;

    const ss = stats.systemStatus || {};
    const agentInfo = stats.agent || {};

    const items = [
        { label: t('status.apiServer'), online: true, detail: t('status.running') },
        { label: t('status.mlClassifier'), online: true, detail: t('status.active') },
        { label: t('status.cache'), online: ss.cache !== false, detail: ss.cache !== false ? t('config.enabled') : t('config.server') },
        { label: t('status.antidetect'), online: ss.antiDetectEnabled, detail: ss.antiDetectEnabled ? t('status.active') : t('status.error') },
        { label: '🤖 ' + t('status.agent'), online: agentInfo.enabled, detail: agentInfo.enabled ? `${t('status.active')} (${agentInfo.activeSessions || 0} sessions)` : t('status.error') },
        { label: '🔍 ' + t('status.scanner'), online: ss.scanner, detail: ss.scanner ? t('ct.send') : 'Regex Only' },
    ];

    list.innerHTML = items.map(i => `
        <div class="status-item">
            <span class="status-label">${i.label}</span>
            <span class="status-badge ${i.online ? 'online' : 'offline'}">${i.detail}</span>
        </div>
    `).join('');
}

// ==================== Config Functions ====================
async function loadConfig() {
    try {
        const response = await fetch('/api/admin/config');
        const cfg = await response.json();

        // Server
        setVal('cfg-server-endpoint', cfg.server?.endpoint);
        setVal('cfg-server-port', cfg.server?.port);
        // Rate Limit
        setVal('cfg-rl-rps', cfg.rateLimit?.rps);
        setVal('cfg-rl-burst', cfg.rateLimit?.burstSize);
        // Cache
        setChecked('cfg-cache-enabled', cfg.cache?.enabled);
        setVal('cfg-cache-size', cfg.cache?.size);
        setVal('cfg-cache-ttl', cfg.cache?.ttl);
        // ML
        setVal('cfg-ml-risk', cfg.ml?.riskThreshold);
        // Anti-Detection
        setChecked('cfg-p3-enabled', cfg.p3?.enabled);
        setVal('cfg-p3-profileId', cfg.p3?.profileId);
        setVal('cfg-p3-proxyTarget', cfg.p3?.proxyTarget);
        setChecked('cfg-p3-directProxy', cfg.p3?.directProxy);
        setChecked('cfg-p3-injectConsist', cfg.p3?.injectConsist);
        // Scanner
        setChecked('cfg-scanner-useBrowser', cfg.scanner?.useBrowser);
        setVal('cfg-scanner-browserWS', cfg.scanner?.browserWS);
        setVal('cfg-scanner-timeout', cfg.scanner?.browserTimeout);
        // Agent
        setChecked('cfg-agent-enabled', cfg.agent?.enabled);
        setVal('cfg-agent-sessionWindow', cfg.agent?.sessionWindow);
        setVal('cfg-agent-fpThreshold', cfg.agent?.fpSwitchRateThreshold);
        setVal('cfg-agent-consThreshold', cfg.agent?.consistencyThreshold);
    } catch (error) {
        console.error('Failed to load config:', error);
    }
}

async function saveConfig() {
    const cfg = {
        rateLimit: {
            rps: getNumVal('cfg-rl-rps'),
            burstSize: getNumVal('cfg-rl-burst'),
        },
        cache: {
            enabled: getChecked('cfg-cache-enabled'),
            size: getNumVal('cfg-cache-size'),
            ttl: getNumVal('cfg-cache-ttl'),
        },
        ml: {
            riskThreshold: getNumVal('cfg-ml-risk'),
        },
        p3: {
            enabled: getChecked('cfg-p3-enabled'),
            profileId: getVal('cfg-p3-profileId'),
            proxyTarget: getVal('cfg-p3-proxyTarget'),
            directProxy: getChecked('cfg-p3-directProxy'),
            injectConsist: getChecked('cfg-p3-injectConsist'),
        },
        scanner: {
            useBrowser: getChecked('cfg-scanner-useBrowser'),
            browserWS: getVal('cfg-scanner-browserWS'),
            browserTimeout: getNumVal('cfg-scanner-timeout'),
        },
        agent: {
            enabled: getChecked('cfg-agent-enabled'),
        },
    };

    try {
        const response = await fetch('/api/admin/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(cfg),
        });
        const result = await response.json();
        if (result.status === 'success') {
            showNotification(t('config.saved'), 'success');
        } else {
            showNotification(t('config.saveFailed'), 'error');
        }
    } catch (error) {
        console.error('Failed to save config:', error);
        showNotification(t('config.saveFailed'), 'error');
    }
}

// Config helpers
function setVal(id, val) { const el = document.getElementById(id); if (el && val !== undefined) el.value = val; }
function setChecked(id, val) { const el = document.getElementById(id); if (el) el.checked = !!val; }
function getVal(id) { return document.getElementById(id)?.value || ''; }
function getNumVal(id) { return parseFloat(document.getElementById(id)?.value) || 0; }
function getChecked(id) { return document.getElementById(id)?.checked || false; }

// ==================== Agent Functions ====================
async function loadAgentData() {
    try {
        const [statusResp, strategiesResp, knowledgeResp] = await Promise.all([
            fetch('/api/admin/agent/status'),
            fetch('/api/admin/agent/strategies'),
            fetch('/api/admin/agent/knowledge'),
        ]);
        const status = await statusResp.json();
        const strategies = await strategiesResp.ok ? await strategiesResp.json() : { strategies: [] };
        const knowledge = await knowledgeResp.ok ? await knowledgeResp.json() : null;

        renderAgentStatus(status);
        renderAgentStrategies(strategies.strategies || []);
        if (knowledge) renderKnowledgeBase(knowledge);
    } catch (error) {
        console.error('Failed to load agent data:', error);
        document.getElementById('agentStatus').textContent = 'Error';
    }
}

function renderAgentStatus(status) {
    const statusEl = document.getElementById('agentStatus');
    const detailEl = document.getElementById('agentStatusDetail');

    if (!status.enabled) {
        statusEl.textContent = 'Disabled';
        statusEl.style.color = 'var(--gray-500)';
        detailEl.textContent = 'Agent is not enabled';
        return;
    }

    statusEl.textContent = 'Active';
    statusEl.style.color = '#10b981';
    detailEl.textContent = status.status;

    const stats = status.stats || {};
    document.getElementById('agentObservations').textContent = stats.total_observations || 0;
    document.getElementById('agentSessions').textContent = stats.active_sessions || 0;
    document.getElementById('agentStrategies').textContent = stats.active_strategies || 0;
    document.getElementById('agentLearnedPatterns').textContent = `${stats.learned_patterns || 0} learned`;
}

function renderAgentStrategies(strategies) {
    const tbody = document.getElementById('agentStrategiesTable');
    if (!strategies || strategies.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="loading">No active strategies</td></tr>';
        return;
    }

    const actionColors = {
        'allow': '#10b981', 'monitor': '#3b82f6',
        'challenge': '#f59e0b', 'throttle': '#f97316', 'block': '#ef4444'
    };

    tbody.innerHTML = strategies.map(s => `
        <tr>
            <td><strong>${escapeHtml(s.name)}</strong><br><small style="color:var(--gray-500)">${escapeHtml(s.description || '')}</small></td>
            <td><span class="badge">${escapeHtml(s.threat_class)}</span></td>
            <td><span class="badge" style="background:${actionColors[s.action] || '#6b7280'};color:white">${s.action}</span></td>
            <td>${s.hit_count || 0}</td>
            <td>${s.learned ? '🧠 Learned' : '📋 Built-in'}</td>
            <td><span class="status-badge ${s.enabled ? 'online' : 'offline'}">${s.enabled ? 'Active' : 'Disabled'}</span></td>
        </tr>
    `).join('');
}

function renderKnowledgeBase(knowledge) {
    const summary = document.getElementById('knowledgeBaseSummary');
    if (!knowledge) {
        summary.innerHTML = '<div class="loading">Knowledge base unavailable</div>';
        return;
    }

    const stats = knowledge.stats || {};
    const families = knowledge.families || [];

    summary.innerHTML = `
        <div class="stats-grid" style="margin-bottom: 16px;">
            <div style="text-align:center;padding:12px;background:var(--gray-100);border-radius:8px;">
                <div style="font-size:24px;font-weight:700;color:var(--primary);">${stats.TotalKnownBrowsers || 0}</div>
                <div style="font-size:12px;color:var(--gray-500);">Browser Families</div>
            </div>
            <div style="text-align:center;padding:12px;background:var(--gray-100);border-radius:8px;">
                <div style="font-size:24px;font-weight:700;color:var(--primary);">${stats.TotalKnownVersions || 0}</div>
                <div style="font-size:12px;color:var(--gray-500);">Known Versions</div>
            </div>
            <div style="text-align:center;padding:12px;background:var(--gray-100);border-radius:8px;">
                <div style="font-size:24px;font-weight:700;color:var(--primary);">${stats.TotalKnownProfiles || 0}</div>
                <div style="font-size:12px;color:var(--gray-500);">Known Profiles</div>
            </div>
        </div>
        <div style="display:flex;flex-wrap:wrap;gap:8px;">
            ${families.map(f => `
                <button class="btn" onclick="showKnowledgeDetail('${escapeHtml(f.family)}')" style="display:flex;align-items:center;gap:6px;">
                    ${getBrowserIcon(f.family)} <strong>${escapeHtml(f.family)}</strong>
                    <small style="color:var(--gray-500);">${f.versions?.length || 0} versions · ${(f.marketShare * 100).toFixed(0)}%</small>
                </button>
            `).join('')}
        </div>
    `;

    // Store for detail view
    state.knowledgeFamilies = families;
}

function showKnowledgeDetail(family) {
    const detail = document.getElementById('knowledgeBaseDetail');
    const families = state.knowledgeFamilies || [];
    const f = families.find(fam => fam.family === family);
    if (!f) {
        detail.innerHTML = '<div>Family not found</div>';
        return;
    }

    const versions = f.versions || [];
    detail.innerHTML = `
        <h4>${getBrowserIcon(family)} ${family} — ${versions.length} Known Versions</h4>
        <div style="margin-top:8px;font-size:13px;color:var(--gray-600);">
            Market Share: ${(f.marketShare * 100).toFixed(1)}% · Common Cipher Suites: ${f.cipherSuites} · Common Extensions: ${f.extensions}
        </div>
        <table class="data-table" style="margin-top:12px;">
            <thead>
                <tr>
                    <th>Version</th>
                    <th>Major</th>
                    <th>TLS</th>
                    <th>Ciphers</th>
                    <th>Extensions</th>
                    <th>H2 Window</th>
                    <th>Released</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                ${versions.map(v => `
                    <tr>
                        <td><strong>${escapeHtml(v.version)}</strong></td>
                        <td>${v.versionMajor}</td>
                        <td>${formatTLSVersion(v.tlsVersion)}</td>
                        <td>${v.cipherSuites}</td>
                        <td>${v.extensions}</td>
                        <td>${v.h2WindowSize ? v.h2WindowSize.toLocaleString() : '-'}</td>
                        <td>${v.releasedYear || '-'}</td>
                        <td>${v.deprecated ? '<span class="badge badge-danger">Deprecated</span>' : '<span class="badge badge-success">Active</span>'}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
}

// Notification
function showNotification(message, type) {
    const existing = document.querySelector('.notification-toast');
    if (existing) existing.remove();

    const colors = { success: '#10b981', error: '#ef4444', info: '#3b82f6' };
    const toast = document.createElement('div');
    toast.className = 'notification-toast';
    toast.style.cssText = `position:fixed;top:20px;right:20px;padding:12px 20px;border-radius:8px;color:white;font-size:14px;z-index:10000;background:${colors[type] || colors.info};box-shadow:0 4px 12px rgba(0,0,0,0.15);`;
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
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

// Profile Detail Modal
window.viewProfile = async function(id) {
    const modal = document.getElementById('profileModal');
    const title = document.getElementById('profileModalTitle');
    const body = document.getElementById('profileModalBody');

    modal.classList.add('active');
    body.innerHTML = '<div class="loading">Loading profile details...</div>';

    try {
        const response = await fetch(`/api/admin/profiles/${id}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const profile = await response.json();

        title.textContent = profile.name || 'Profile Detail';

        body.innerHTML = `
            <div class="profile-detail-section">
                <h4>Basic Information</h4>
                <div class="profile-detail-grid">
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">ID</span>
                        <span class="profile-detail-value"><code>${profile.id}</code></span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Name</span>
                        <span class="profile-detail-value">${profile.name}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Browser</span>
                        <span class="profile-detail-value">${getBrowserIcon(profile.browserType)} ${profile.browserType} ${profile.browserVersion}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Operating System</span>
                        <span class="profile-detail-value">${profile.os} ${profile.osVersion} (${profile.osArch} ${profile.osBitness})</span>
                    </div>
                    ${profile.description ? `
                    <div class="profile-detail-item full-width">
                        <span class="profile-detail-label">Description</span>
                        <span class="profile-detail-value">${profile.description}</span>
                    </div>
                    ` : ''}
                </div>
            </div>

            <div class="profile-detail-section">
                <h4>TLS Configuration</h4>
                <div class="profile-detail-grid">
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">TLS Version</span>
                        <span class="profile-detail-value">${formatTLSVersion(profile.tlsVersion)} (0x${profile.tlsVersion?.toString(16).padStart(4, '0')})</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Cipher Suites</span>
                        <span class="profile-detail-value">${profile.cipherSuites?.length || 0} suites</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Extensions</span>
                        <span class="profile-detail-value">${profile.extensions?.length || 0} extensions</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Supported Curves</span>
                        <span class="profile-detail-value">${profile.supportedCurves?.length || 0} curves</span>
                    </div>
                </div>

                ${profile.cipherSuites?.length ? `
                <div class="profile-detail-section" style="margin-top: 16px;">
                    <h4>Cipher Suites (${profile.cipherSuites.length})</h4>
                    <div class="profile-detail-tags">
                        ${profile.cipherSuites.map(cs => `<span class="tag">0x${cs.toString(16).toUpperCase().padStart(4, '0')}</span>`).join('')}
                    </div>
                </div>
                ` : ''}

                ${profile.extensions?.length ? `
                <div class="profile-detail-section" style="margin-top: 16px;">
                    <h4>TLS Extensions (${profile.extensions.length})</h4>
                    <div class="profile-detail-tags">
                        ${profile.extensions.map(ext => `<span class="tag" title="Type: ${ext.Type}">${getExtensionName(ext.Type)}</span>`).join('')}
                    </div>
                </div>
                ` : ''}
            </div>

            ${profile.tcpip ? `
            <div class="profile-detail-section">
                <h4>TCP/IP Fingerprint</h4>
                <div class="profile-detail-grid">
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">TTL</span>
                        <span class="profile-detail-value">${profile.tcpip.ttl}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Window Size</span>
                        <span class="profile-detail-value">${profile.tcpip.windowSize?.toLocaleString()}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">MSS</span>
                        <span class="profile-detail-value">${profile.tcpip.mss}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Window Scale</span>
                        <span class="profile-detail-value">${profile.tcpip.windowScale || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">DF Flag</span>
                        <span class="profile-detail-value">${profile.tcpip.df ? 'Set' : 'Not Set'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">SACK Permitted</span>
                        <span class="profile-detail-value">${profile.tcpip.sackPermitted ? 'Yes' : 'No'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Timestamps</span>
                        <span class="profile-detail-value">${profile.tcpip.timestamps ? 'Enabled' : 'Disabled'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">NOP Count</span>
                        <span class="profile-detail-value">${profile.tcpip.noOperation}</span>
                    </div>
                    <div class="profile-detail-item full-width">
                        <span class="profile-detail-label">TCP Options Signature</span>
                        <span class="profile-detail-value"><code>${profile.tcpip.optionsSignature}</code></span>
                    </div>
                    ${profile.tcpip.ja4t ? `
                    <div class="profile-detail-item full-width">
                        <span class="profile-detail-label">JA4T Fingerprint</span>
                        <span class="profile-detail-value"><code>${profile.tcpip.ja4t}</code></span>
                    </div>
                    ` : ''}
                </div>
            </div>
            ` : ''}

            <div class="profile-detail-section">
                <h4>HTTP Protocol Support</h4>
                <div class="profile-detail-grid">
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">HTTP/1.1</span>
                        <span class="profile-detail-value"><span class="status-badge online">Supported</span></span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">HTTP/2</span>
                        <span class="profile-detail-value">${profile.http2Settings?.headerTableSize > 0 ? '<span class="status-badge online">Supported</span>' : '<span class="status-badge offline">Not Configured</span>'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">HTTP/3 (QUIC)</span>
                        <span class="profile-detail-value">${profile.http3Supported ? '<span class="status-badge online">Supported</span>' : '<span class="status-badge offline">Not Configured</span>'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">ALPN</span>
                        <span class="profile-detail-value">${profile.http3Supported ? 'h3, h2, http/1.1' : (profile.http2Settings?.headerTableSize > 0 ? 'h2, http/1.1' : 'http/1.1')}</span>
                    </div>
                </div>
            </div>

            <div class="profile-detail-section">
                <h4>HTTP/2 Settings</h4>
                <div class="profile-detail-grid">
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Header Table Size</span>
                        <span class="profile-detail-value">${profile.http2Settings?.headerTableSize?.toLocaleString() || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Max Concurrent Streams</span>
                        <span class="profile-detail-value">${profile.http2Settings?.maxConcurrentStreams || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Initial Window Size</span>
                        <span class="profile-detail-value">${profile.http2Settings?.initialWindowSize?.toLocaleString() || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Max Frame Size</span>
                        <span class="profile-detail-value">${profile.http2Settings?.maxFrameSize?.toLocaleString() || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Enable Push</span>
                        <span class="profile-detail-value">${profile.http2Settings?.enablePush === 0 ? 'No' : 'Yes'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Connection Flow</span>
                        <span class="profile-detail-value">${profile.connectionFlow?.toLocaleString() || 'N/A'}</span>
                    </div>
                </div>
            </div>

            ${profile.http3Supported ? `
            <div class="profile-detail-section">
                <h4>HTTP/3 (QUIC) Settings</h4>
                <div class="profile-detail-grid">
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">QUIC Version</span>
                        <span class="profile-detail-value">${profile.http3Settings?.quicVersion === 1 ? 'RFC 9000 (v1)' : '0x' + profile.http3Settings?.quicVersion?.toString(16)}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Initial Max Data</span>
                        <span class="profile-detail-value">${profile.http3Settings?.initialMaxData?.toLocaleString() || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Initial Max Stream Data</span>
                        <span class="profile-detail-value">${profile.http3Settings?.initialMaxStreamData?.toLocaleString() || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Max Bidi Streams</span>
                        <span class="profile-detail-value">${profile.http3Settings?.initialMaxStreamsBidi || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Max Uni Streams</span>
                        <span class="profile-detail-value">${profile.http3Settings?.initialMaxStreamsUni || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Max UDP Payload</span>
                        <span class="profile-detail-value">${profile.http3Settings?.maxUDPPayloadSize?.toLocaleString() || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">ACK Delay Exponent</span>
                        <span class="profile-detail-value">${profile.http3Settings?.ackDelayExponent || 'N/A'}</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Max ACK Delay</span>
                        <span class="profile-detail-value">${profile.http3Settings?.maxAckDelay || 'N/A'} ms</span>
                    </div>
                    <div class="profile-detail-item">
                        <span class="profile-detail-label">Active Migration</span>
                        <span class="profile-detail-value">${profile.http3Settings?.disableActiveMigration ? 'Disabled' : 'Enabled'}</span>
                    </div>
                </div>
            </div>
            ` : ''}

            ${profile.headers ? `
            <div class="profile-detail-section">
                <h4>HTTP Headers</h4>
                <pre class="profile-detail-raw">${JSON.stringify(profile.headers, null, 2)}</pre>
            </div>
            ` : ''}

            <div class="profile-detail-section">
                <h4>Raw Profile Data</h4>
                <pre class="profile-detail-raw">${JSON.stringify(profile, null, 2)}</pre>
            </div>
        `;
    } catch (error) {
        console.error('Failed to load profile:', error);
        body.innerHTML = `<div class="error">Failed to load profile: ${error.message}</div>`;
    }
};

window.closeProfileModal = function() {
    const modal = document.getElementById('profileModal');
    modal.classList.remove('active');
};

// Close modal on overlay click
document.addEventListener('click', function(e) {
    const modal = document.getElementById('profileModal');
    if (e.target === modal) {
        closeProfileModal();
    }
});

// Close modal on Escape key
document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') {
        closeProfileModal();
    }
});

// Charts (placeholder - would integrate with Chart.js)
function renderBrowserChart(data) {
    const container = document.getElementById('browserChart');
    if (!container) return;

    // 计算总数和百分比
    const total = Object.values(data).reduce((a, b) => a + b, 0);
    if (total === 0) {
        container.innerHTML = '<div class="loading">No data available</div>';
        return;
    }

    // 转换为百分比并排序
    const sortedData = Object.entries(data)
        .map(([browser, count]) => ({ browser, count, percentage: (count / total * 100).toFixed(1) }))
        .sort((a, b) => b.count - a.count);

    const maxCount = sortedData[0].count;

    const colors = {
        'chrome': '#4285f4',
        'firefox': '#ff7139',
        'safari': '#00d8ff',
        'edge': '#0078d7',
        'opera': '#ff1b2d',
        'brave': '#fb542b',
        'samsung': '#1428a0'
    };

    container.innerHTML = `
        <div style="display: flex; flex-direction: column; gap: 12px; width: 100%; padding: 10px;">
            ${sortedData.map(({ browser, count, percentage }) => `
                <div style="display: flex; align-items: center; gap: 12px;">
                    <span style="width: 90px; font-size: 13px; font-weight: 500; color: var(--gray-700);">${getBrowserIcon(browser)} ${capitalize(browser)}</span>
                    <div style="flex: 1; background: var(--gray-200); height: 24px; border-radius: 4px; overflow: hidden;">
                        <div style="width: ${(count / maxCount * 100)}%; background: ${colors[browser] || 'var(--primary)'}; height: 100%; border-radius: 4px; transition: width 0.5s ease;"></div>
                    </div>
                    <span style="width: 60px; text-align: right; font-size: 12px; color: var(--gray-600);">${count} (${percentage}%)</span>
                </div>
            `).join('')}
        </div>
    `;
}

function renderOSChart(data) {
    const container = document.getElementById('osChart');
    if (!container) return;

    // 计算总数
    const total = Object.values(data).reduce((a, b) => a + b, 0);
    if (total === 0) {
        container.innerHTML = '<div class="loading">No data available</div>';
        return;
    }

    // 合并和简化 OS 名称
    const osGroups = {};
    Object.entries(data).forEach(([os, count]) => {
        let group = 'Other';
        if (os.includes('Windows')) group = 'Windows';
        else if (os.includes('Mac OS') || os.includes('Macintosh')) group = 'macOS';
        else if (os.includes('Linux') && !os.includes('Android')) group = 'Linux';
        else if (os.includes('Android')) group = 'Android';
        else if (os.includes('iPhone') || os.includes('iPad') || os.includes('iOS')) group = 'iOS';
        osGroups[group] = (osGroups[group] || 0) + count;
    });

    // 排序
    const sortedData = Object.entries(osGroups)
        .map(([os, count]) => ({ os, count, percentage: (count / total * 100).toFixed(1) }))
        .sort((a, b) => b.count - a.count);

    let currentAngle = 0;
    const colors = {
        'Windows': '#0078d7',
        'macOS': '#555555',
        'Linux': '#fcc624',
        'Android': '#3ddc84',
        'iOS': '#007aff',
        'Other': '#9ca3af'
    };

    container.innerHTML = `
        <div style="display: flex; align-items: center; gap: 24px; padding: 10px;">
            <div style="width: 160px; height: 160px; border-radius: 50%; background: conic-gradient(
                ${sortedData.map(({ os, count }) => {
                    const angle = (count / total) * 360;
                    const gradient = `${colors[os] || '#9ca3af'} ${currentAngle}deg ${currentAngle + angle}deg`;
                    currentAngle += angle;
                    return gradient;
                }).join(', ')}
            ); box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1);"></div>
            <div style="display: flex; flex-direction: column; gap: 8px;">
                ${sortedData.map(({ os, count, percentage }) => `
                    <div style="display: flex; align-items: center; gap: 10px;">
                        <div style="width: 14px; height: 14px; background: ${colors[os] || '#9ca3af'}; border-radius: 3px;"></div>
                        <span style="font-size: 13px; color: var(--gray-700); min-width: 70px;">${os}</span>
                        <span style="font-size: 12px; color: var(--gray-500);">${count} (${percentage}%)</span>
                    </div>
                `).join('')}
            </div>
        </div>
    `;
}

// Traffic Chart
function renderTrafficChart(data) {
    const container = document.getElementById('trafficChart');
    if (!container) return;

    if (!data || !data.labels || !data.data || data.labels.length === 0) {
        container.innerHTML = '<div class="loading">No traffic data available</div>';
        return;
    }

    const maxValue = Math.max(...data.data);
    const labels = data.labels;
    const values = data.data;

    container.innerHTML = `
        <div style="display: flex; align-items: flex-end; justify-content: space-around; height: 100%; padding: 20px; gap: 8px;">
            ${values.map((val, i) => {
                const height = maxValue > 0 ? (val / maxValue * 100) : 0;
                return `
                    <div style="display: flex; flex-direction: column; align-items: center; gap: 8px; flex: 1;">
                        <div style="font-size: 11px; color: var(--gray-600); font-weight: 500;">${val}</div>
                        <div style="width: 100%; max-width: 60px; height: ${Math.max(height, 5)}px;
                                    background: linear-gradient(to top, var(--primary), #818cf8);
                                    border-radius: 4px 4px 0 0; transition: height 0.5s ease;"></div>
                        <span style="font-size: 11px; color: var(--gray-500);">${labels[i] || ''}</span>
                    </div>
                `;
            }).join('')}
        </div>
    `;
}

// Helper function to capitalize first letter
function capitalize(str) {
    if (!str) return '';
    return str.charAt(0).toUpperCase() + str.slice(1);
}

// TCP/IP Fingerprint Chart
function renderTCPIPChart(data) {
    const container = document.getElementById('tcpipChart');
    if (!container) return;

    // 计算总数
    const total = Object.values(data).reduce((a, b) => a + b, 0);
    if (total === 0) {
        container.innerHTML = '<div class="loading">No data available</div>';
        return;
    }

    // 排序
    const sortedData = Object.entries(data)
        .map(([os, count]) => ({ os, count, percentage: (count / total * 100).toFixed(1) }))
        .sort((a, b) => b.count - a.count);

    const maxCount = sortedData[0].count;

    const colors = {
        'Windows': '#0078d7',
        'macOS/iOS': '#555555',
        'Linux': '#fcc624',
        'Unknown': '#9ca3af',
        'Other': '#6b7280'
    };

    container.innerHTML = `
        <div style="display: flex; flex-direction: column; gap: 12px; width: 100%; padding: 10px;">
            ${sortedData.map(({ os, count, percentage }) => `
                <div style="display: flex; align-items: center; gap: 12px;">
                    <span style="width: 100px; font-size: 13px; font-weight: 500; color: var(--gray-700);">${os}</span>
                    <div style="flex: 1; background: var(--gray-200); height: 24px; border-radius: 4px; overflow: hidden;">
                        <div style="width: ${(count / maxCount * 100)}%; background: ${colors[os] || 'var(--primary)'}; height: 100%; border-radius: 4px; transition: width 0.5s ease;"></div>
                    </div>
                    <span style="width: 80px; text-align: right; font-size: 12px; color: var(--gray-600);">${count} (${percentage}%)</span>
                </div>
            `).join('')}
        </div>
    `;
}

// TLS Extension names
function getExtensionName(type) {
    const names = {
        0: 'server_name',
        1: 'max_fragment_length',
        5: 'status_request',
        10: 'supported_groups',
        11: 'ec_point_formats',
        13: 'signature_algorithms',
        14: 'use_srtp',
        15: 'heartbeat',
        16: 'ALPN',
        17: 'status_request_v2',
        18: 'signed_certificate_timestamp',
        19: 'client_certificate_type',
        20: 'server_certificate_type',
        21: 'padding',
        22: 'encrypt_then_mac',
        23: 'extended_master_secret',
        24: 'token_binding',
        25: 'cached_info',
        35: 'session_ticket',
        41: 'pre_shared_key',
        42: 'early_data',
        43: 'supported_versions',
        44: 'cookie',
        45: 'psk_key_exchange_modes',
        47: 'certificate_authorities',
        48: 'oid_filters',
        49: 'post_handshake_auth',
        50: 'signature_algorithms_cert',
        51: 'key_share',
        57: 'quic_transport_params',
        127: 'renegotiation_info',
        17513: 'application_settings',
        65281: 'renegotiation_info'
    };
    return names[type] || `ext_${type}`;
}


// ==================== Client Test Functions ====================

// 加载 Client Test 页面的 Profile 列表
function loadClientTestProfiles() {
    const select = document.getElementById('testProfileSelect');
    if (!select) return;

    if (state.profiles.length === 0) {
        select.innerHTML = '<option value="">No profiles available</option>';
        return;
    }

    select.innerHTML = state.profiles.map(p =>
        `<option value="${p.id}">${p.name} (${p.browserType})</option>`
    ).join('');
}

// 运行客户端测试
async function runClientTest() {
    const profileSelect = document.getElementById('testProfileSelect');
    const urlInput = document.getElementById('testUrl');
    const methodSelect = document.getElementById('testMethod');
    const bodyInput = document.getElementById('testBody');
    const resultDiv = document.getElementById('testResult');

    const profileId = profileSelect.value;
    const url = urlInput.value.trim();
    const method = methodSelect.value;
    const body = bodyInput.value.trim();

    if (!profileId) {
        alert('Please select a profile');
        return;
    }
    if (!url) {
        alert('Please enter a URL');
        return;
    }

    // 显示加载状态
    resultDiv.style.display = 'block';
    document.getElementById('requestTraceSection').style.display = 'none';
    document.getElementById('responseTraceSection').style.display = 'none';

    try {
        const response = await fetch('/api/admin/client/test', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                profileId: profileId,
                url: url,
                method: method,
                body: body || undefined
            })
        });

        const data = await response.json();

        // 始终显示请求指纹（即使请求失败）
        if (data.requestTrace) {
            displayRequestTrace(data.requestTrace);
        }

        if (data.success && data.responseTrace) {
            // 检查是否启用扫描模式
            const enableScanning = document.getElementById('testEnableScanning')?.checked || false;
            if (enableScanning && data.responseTrace.bodyPreview) {
                // 步骤1: 先显示快速的客户端扫描结果（正则表达式）
                const quickScanResults = scanFingerprintDetectionCode(data.responseTrace.bodyPreview);
                data.scanResults = quickScanResults;

                // 步骤2: 异步调用后端V8扫描器进行深度分析
                callV8Scanner(data.responseTrace.bodyPreview, url).then(v8Results => {
                    data.v8Results = v8Results;
                    // 更新扫描结果显示，包含V8结果
                    if (document.getElementById('scanner-content')?.style.display !== 'none') {
                        displayScannerResults(quickScanResults, v8Results);
                    }
                }).catch(err => {
                    console.error('V8扫描失败:', err);
                    // 即使V8扫描失败，也继续显示快速扫描结果
                });
            }
            // 显示响应信息
            displayResponseTrace(data.responseTrace, data.scanResults);
        } else {
            // 显示错误
            document.getElementById('responseTraceSection').style.display = 'block';
            document.querySelector('#responseTraceSection .result-status').innerHTML =
                `<span class="badge badge-danger">✗ Failed</span> <span class="badge">${data.error || 'Error'}</span>`;
            document.querySelector('#responseTraceSection .result-body').textContent =
                data.error || 'Unknown error';
        }

    } catch (error) {
        console.error('Client test error:', error);
        const respSection = document.getElementById('responseTraceSection');
        if (respSection) {
            respSection.style.display = 'block';
            const statusDiv = respSection.querySelector('.result-status');
            if (statusDiv) {
                statusDiv.innerHTML = '<span class="badge badge-danger">✗ Error</span>';
            }
            const bodyPre = respSection.querySelector('.result-body');
            if (bodyPre) {
                bodyPre.textContent = error.message || 'Unknown error';
            }
        }
    }
}

// 显示请求追踪信息
function displayRequestTrace(trace) {
    const section = document.getElementById('requestTraceSection');
    if (!section) {
        console.error('requestTraceSection not found');
        return;
    }
    section.style.display = 'block';

    // TCP/IP 信息
    if (trace.tcpip) {
        const tcpipTable = document.getElementById('tcpip-table');
        tcpipTable.innerHTML = `
            <tr><td>TTL (Time To Live)</td><td>${trace.tcpip.ttl}</td></tr>
            <tr><td>Window Size</td><td>${trace.tcpip.windowSize} bytes</td></tr>
            <tr><td>MSS (Max Segment Size)</td><td>${trace.tcpip.mss} bytes</td></tr>
            <tr><td>Window Scale</td><td>${trace.tcpip.windowScale}</td></tr>
            <tr><td>Don't Fragment (DF)</td><td>${trace.tcpip.df ? 'Yes' : 'No'}</td></tr>
            <tr><td>SACK Permitted</td><td>${trace.tcpip.sackPermitted ? 'Yes' : 'No'}</td></tr>
            <tr><td>TCP Timestamps</td><td>${trace.tcpip.timestamps ? 'Enabled' : 'Disabled'}</td></tr>
            <tr><td>JA4T Fingerprint</td><td><code>${trace.tcpip.ja4t || 'N/A'}</code></td></tr>
        `;
    }

    // TLS 信息
    if (trace.tls) {
        const tlsTable = document.getElementById('tls-table');
        const cipherSuites = trace.tls.cipherSuites ? trace.tls.cipherSuites.slice(0, 10).map(cs =>
            `<span class="trace-array-item">${cs}</span>`
        ).join('') : '';
        const extensions = trace.tls.extensions ? trace.tls.extensions.slice(0, 10).map(ext =>
            `<span class="trace-array-item">${ext}</span>`
        ).join('') : '';

        tlsTable.innerHTML = `
            <tr><td>TLS Version</td><td><span class="tls-version tls13">${trace.tls.version}</span></td></tr>
            <tr><td>ClientHello ID</td><td>${trace.tls.clientHelloId}</td></tr>
            <tr><td>Cipher Suites (${trace.tls.cipherSuites?.length || 0})</td><td><div class="trace-array">${cipherSuites}</div></td></tr>
            <tr><td>Extensions (${trace.tls.extensions?.length || 0})</td><td><div class="trace-array">${extensions}</div></td></tr>
            <tr><td>Supported Groups</td><td>${(trace.tls.supportedGroups || []).join(', ')}</td></tr>
            <tr><td>ALPN Protocols</td><td>${(trace.tls.alpnProtocols || []).join(', ')}</td></tr>
        `;

        // JA3 指纹
        document.getElementById('ja3-string').textContent = trace.tls.ja3 || 'N/A';
        document.getElementById('ja3-hash').textContent = trace.tls.ja3Hash || 'N/A';
    }

    // HTTP 信息
    if (trace.http) {
        // HTTP Headers
        const headersTable = document.getElementById('http-headers-table');
        let headersHtml = '';
        for (const [key, value] of Object.entries(trace.http.headers || {})) {
            if (value) {
                headersHtml += `<tr><td>${key}</td><td>${value}</td></tr>`;
            }
        }
        headersTable.innerHTML = headersHtml || '<tr><td colspan="2">No headers</td></tr>';

        // HTTP/2 Settings
        if (trace.http.http2Settings) {
            const h2 = trace.http.http2Settings;
            document.getElementById('http2-settings-table').innerHTML = `
                <tr><td>Header Table Size</td><td>${h2.headerTableSize}</td></tr>
                <tr><td>Enable Push</td><td>${h2.enablePush}</td></tr>
                <tr><td>Max Concurrent Streams</td><td>${h2.maxConcurrentStreams}</td></tr>
                <tr><td>Initial Window Size</td><td>${h2.initialWindowSize}</td></tr>
                <tr><td>Max Frame Size</td><td>${h2.maxFrameSize}</td></tr>
                <tr><td>Max Header List Size</td><td>${h2.maxHeaderListSize}</td></tr>
                <tr><td>Connection Flow</td><td>${h2.connectionFlow}</td></tr>
                <tr><td>Priority Frames</td><td>${h2.priorityFrames}</td></tr>
            `;
        }
    }
}

// 显示响应追踪信息
function displayResponseTrace(trace, scanResults) {
    document.getElementById('responseTraceSection').style.display = 'block';

    // 状态
    const statusClass = trace.statusCode >= 200 && trace.statusCode < 300 ? 'badge-success' :
                        trace.statusCode >= 400 ? 'badge-danger' : 'badge-warning';
    document.querySelector('#responseTraceSection .result-status').innerHTML =
        `<span class="badge ${statusClass}">${trace.statusCode} ${trace.status}</span> ` +
        `<span class="badge">${trace.bodyLength} bytes</span> ` +
        `<span class="badge">${trace.responseTime}</span>`;

    // 响应头
    const respHeadersTable = document.getElementById('response-headers-table');
    let respHeadersHtml = '';
    for (const [key, value] of Object.entries(trace.headers || {})) {
        respHeadersHtml += `<tr><td>${key}</td><td>${value}</td></tr>`;
    }
    respHeadersTable.innerHTML = respHeadersHtml || '<tr><td colspan="2">No headers</td></tr>';

    // 响应体
    document.querySelector('#responseTraceSection .result-body').textContent = trace.bodyPreview || '';

    // 显示扫描结果
    if (scanResults && scanResults.detected.length > 0) {
        displayScannerResults(scanResults);
    }
}

// 切换追踪标签页
function switchTraceTab(tab) {
    // 更新按钮状态
    document.querySelectorAll('.trace-tab').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');

    // 显示对应内容
    document.querySelectorAll('.trace-content').forEach(content => {
        content.style.display = 'none';
    });
    document.getElementById(tab + '-content').style.display = 'block';
}

// 清除测试结果
function clearTestResult() {
    const resultDiv = document.getElementById('testResult');
    if (resultDiv) {
        resultDiv.style.display = 'none';
    }
    document.getElementById('testBody').value = '';

    // 重置追踪显示
    const requestSection = document.getElementById('requestTraceSection');
    const responseSection = document.getElementById('responseTraceSection');
    if (requestSection) requestSection.style.display = 'none';
    if (responseSection) responseSection.style.display = 'none';
}

// 当切换到 Client Test 页面时加载 Profile 列表
document.addEventListener('DOMContentLoaded', () => {
    // 监听页面切换事件
    const navItems = document.querySelectorAll('.nav-item');
    navItems.forEach(item => {
        item.addEventListener('click', () => {
            const page = item.getAttribute('data-page');
            if (page === 'client-test') {
                loadClientTestProfiles();
            }
        });
    });
});

// 将反检测代码注入到 HTML 中
function injectAntiDetectCodeIntoHTML(htmlContent, adCode) {
    if (!htmlContent || !adCode) {
        return htmlContent;
    }

    // 创建 script 标签
    const scriptTag = `<script>\n// Anti-Detection Code Injected\n${adCode}\n</script>`;

    // 尝试在 <head> 后注入
    const headMatch = htmlContent.match(/<head[^>]*>/i);
    if (headMatch) {
        return htmlContent.substring(0, headMatch[0].length) + '\n' + scriptTag + '\n' + htmlContent.substring(headMatch[0].length);
    }

    // 尝试在 <html> 后注入
    const htmlMatch = htmlContent.match(/<html[^>]*>/i);
    if (htmlMatch) {
        return htmlContent.substring(0, htmlMatch[0].length) + '\n' + scriptTag + '\n' + htmlContent.substring(htmlMatch[0].length);
    }

    // 在最前面注入
    return scriptTag + '\n' + htmlContent;
}

// 扫描指纹检测代码
function scanFingerprintDetectionCode(htmlContent) {
    const detected = [];
    const notDetected = [];

    // 定义检测模式
    const patterns = {
        'WebGPU Detection': {
            regex: /navigator\.gpu|GPUAdapter|requestAdapter|createRenderPipeline|wgpu/gi,
            severity: 'high',
            description: '检测 WebGPU 支持'
        },
        'Automation Detection': {
            regex: /navigator\.webdriver|window\.document\.documentElement\.getAttribute.*webdriver|chrome\.runtime|__proto__\.__proto__/gi,
            severity: 'high',
            description: '检测自动化工具特征'
        },
        'MediaDevices Detection': {
            regex: /navigator\.mediaDevices|enumerateDevices|getUserMedia|getDisplayMedia/gi,
            severity: 'medium',
            description: '检测摄像头/麦克风设备'
        },
        'Permissions Detection': {
            regex: /navigator\.permissions|\.query\('camera'\)|\.query\('microphone'\)|\.query\('geolocation'\)/gi,
            severity: 'medium',
            description: '检测权限状态'
        },
        'Plugin Detection': {
            regex: /navigator\.plugins|mimeTypes|ActiveXObject/gi,
            severity: 'low',
            description: '检测浏览器插件'
        },
        'Canvas Fingerprinting': {
            regex: /canvas|getContext.*2d|fillText|measureText|toDataURL/gi,
            severity: 'medium',
            description: '可能的 Canvas 指纹识别'
        },
        'WebGL Fingerprinting': {
            regex: /webgl|getParameter|RENDERER|VENDOR/gi,
            severity: 'medium',
            description: '可能的 WebGL 指纹识别'
        },
        'User Agent Detection': {
            regex: /navigator\.userAgent|User-Agent|headless|HeadlessChrome/gi,
            severity: 'low',
            description: '检测 User-Agent 特征'
        }
    };

    // 执行扫描
    for (const [name, pattern] of Object.entries(patterns)) {
        const matches = htmlContent.match(pattern.regex);
        if (matches && matches.length > 0) {
            detected.push({
                name: name,
                severity: pattern.severity,
                description: pattern.description,
                count: matches.length,
                samples: matches.slice(0, 3)
            });
        } else {
            notDetected.push({
                name: name,
                description: pattern.description
            });
        }
    }

    return {
        detected: detected,
        notDetected: notDetected,
        totalDetected: detected.length,
        totalNotDetected: notDetected.length
    };
}

// 显示扫描结果（支持客户端和服务端结果合并）
function displayScannerResults(quickResults, v8Results) {
    const container = document.getElementById('scannerResults');
    if (!container) return;

    let html = '';

    // 确定显示哪个结果（优先显示V8结果，如果没有则显示快速扫描结果）
    const primaryResults = v8Results || quickResults;

    if (v8Results) {
        html += '<div style="background: #0066ff20; border: 1px solid #0066ff; padding: 8px; border-radius: 4px; margin-bottom: 12px;">';
        html += '<strong>🔬 动态V8分析结果（深度扫描）</strong>';
        html += '<p style="color: #0066; font-size: 12px; margin: 4px 0;">基于V8引擎实时执行分析</p>';
        html += '</div>';
    } else {
        html += '<div style="background: #ffaa0020; border: 1px solid #ffaa00; padding: 8px; border-radius: 4px; margin-bottom: 12px;">';
        html += '<strong>⚡ 快速静态扫描结果（正则分析）</strong>';
        html += '<p style="color: #ff8800; font-size: 12px; margin: 4px 0;">基于HTML源代码的快速模式匹配</p>';
        html += '</div>';
    }

    // 显示检测摘要
    const highCount = primaryResults.detected.filter(d => d.severity === 'high').length;
    const mediumCount = primaryResults.detected.filter(d => d.severity === 'medium').length;
    const lowCount = primaryResults.detected.filter(d => d.severity === 'low').length;

    const summaryColor = highCount > 0 ? '#ff4444' : mediumCount > 0 ? '#ffaa00' : '#44aa44';
    html += `<div style="background: ${summaryColor}20; border-left: 4px solid ${summaryColor}; padding: 12px; margin-bottom: 12px; border-radius: 4px;">
        <strong>📊 检测摘要</strong><br>
        <span style="color: #ff4444;">🔴 高风险: ${highCount}</span> |
        <span style="color: #ffaa00;">🟡 中风险: ${mediumCount}</span> |
        <span style="color: #44aa44;">🟢 低风险: ${lowCount}</span>
        ${v8Results && v8Results.dynamicFeaturesFound ? '<br><span style="color: #0066ff;">✓ 发现动态指纹检测</span>' : ''}
    </div>`;

    // 显示检测到的项目
    if (primaryResults.detected.length > 0) {
        html += '<h5 style="margin-top: 12px; color: #ff4444;">⚠️ 检测到的指纹检测代码</h5>';
        html += '<table class="trace-table">';
        primaryResults.detected.forEach(item => {
            const severityIcon = item.severity === 'high' ? '🔴' : item.severity === 'medium' ? '🟡' : '🟢';
            const samplesHtml = (item.samples || []).slice(0, 2).map(s => {
                // 截断长样本
                const displayStr = typeof s === 'string' ? s.substring(0, 60) : String(s).substring(0, 60);
                return `<code style="background: #f0f0f0; padding: 2px 4px; border-radius: 2px;">${displayStr}</code>`;
            }).join('<br>');

            html += `
                <tr>
                    <td style="width: 40%;">
                        <strong>${severityIcon} ${item.name}</strong><br>
                        <small style="color: #888;">${item.description}</small>
                    </td>
                    <td>
                        <strong>次数:</strong> ${item.count || 1}<br>
                        <strong>样本:</strong> ${samplesHtml || '<code>-</code>'}
                    </td>
                </tr>
            `;
        });
        html += '</table>';
    }

    // 显示未检测到的项目
    if (primaryResults.notDetected && primaryResults.notDetected.length > 0) {
        html += '<h5 style="margin-top: 12px; color: #44aa44;">✓ 未检测到</h5>';
        html += '<ul style="list-style: none; padding: 0;">';
        primaryResults.notDetected.forEach(item => {
            html += `<li style="margin: 4px 0; color: #44aa44;">✓ ${typeof item === 'string' ? item : (item.name || item)}</li>`;
        });
        html += '</ul>';
    }

    // 如果是V8结果，显示执行详情
    if (v8Results && v8Results.executionDetails) {
        html += '<h5 style="margin-top: 12px;">📋 脚本执行详情</h5>';
        html += '<pre style="background: #f5f5f5; padding: 8px; border-radius: 4px; overflow-x: auto; font-size: 11px;">';
        html += escapeHtml(v8Results.executionDetails);
        html += '</pre>';
    }

    container.innerHTML = html;

    // 显示 Scanner 标签页
    const scannerTab = Array.from(document.querySelectorAll('.trace-tab')).find(tab =>
        tab.textContent.includes('Scanner')
    );
    if (scannerTab) {
        // 自动切换到 Scanner 标签页
        document.querySelectorAll('.trace-tab').forEach(btn => btn.classList.remove('active'));
        scannerTab.classList.add('active');
        document.querySelectorAll('.trace-content').forEach(content => {
            content.style.display = 'none';
        });
        document.getElementById('scanner-content').style.display = 'block';
    }
}

// HTML转义函数
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 调用后端V8扫描器
async function callV8Scanner(htmlContent, url) {
    try {
        const gatewayEndpoint = getGatewayUrl();
        const response = await fetch(gatewayEndpoint + '/scan', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                html: htmlContent,
                url: url,
                followRedirects: true,
                maxRedirects: 10,
                executeJs: true,
                waitMs: 3000
            })
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`扫描失败: ${response.status} - ${errorText}`);
        }

        const data = await response.json();
        return data.result || null;
    } catch (error) {
        console.error('V8扫描接口调用失败:', error);
        throw error;
    }
}

// 获取网关地址
function getGatewayUrl() {
    // 从当前URL推断网关地址
    const urlObj = new URL(window.location.href);
    return `${urlObj.protocol}//${urlObj.hostname}${urlObj.port ? ':' + urlObj.port : ''}`;
}

// =====================================================================
//  HELPER: 填充 Profile 下拉列表
// =====================================================================
function populateProfileSelect(selectId) {
    const sel = document.getElementById(selectId);
    if (!sel) return;
    if (state.profiles.length === 0) {
        sel.innerHTML = '<option value="">No profiles</option>';
        return;
    }
    sel.innerHTML = state.profiles.map(p =>
        `<option value="${p.id}">${p.name} (${p.browserType})</option>`
    ).join('');
}

function riskColor(score) {
    if (score >= 0.9) return '#9c27b0';
    if (score >= 0.7) return '#f44336';
    if (score >= 0.4) return '#ff9800';
    if (score >= 0.2) return '#8bc34a';
    return '#4caf50';
}

// =====================================================================
//  ANALYZE PAGE
// =====================================================================
function loadAnalyzePage() {
    populateProfileSelect('analyzeProfileSelect');
}

async function runAnalyzeProfile() {
    const id = document.getElementById('analyzeProfileSelect').value;
    if (!id) return alert('请选择 Profile');
    try {
        const r = await API.analyzeProfile(id);
        const section = document.getElementById('analyzeResult');
        section.style.display = 'block';

        // 状态卡片
        document.getElementById('analyzeStats').innerHTML = `
            <div class="stat-card"><div class="stat-icon">🔑</div><div class="stat-info"><h3>Hash</h3><p class="stat-value" style="font-size:12px;word-break:break-all">${r.fingerprintHash || '-'}</p></div></div>
            <div class="stat-card"><div class="stat-icon">💾</div><div class="stat-info"><h3>Cached</h3><p class="stat-value">${r.cached ? 'Yes' : 'No'}</p></div></div>
            <div class="stat-card"><div class="stat-icon">⏱️</div><div class="stat-info"><h3>处理时间</h3><p class="stat-value">${r.processingTimeMs || 0} ms</p></div></div>
        `;

        // 分类
        const c = r.classification || {};
        document.getElementById('analyzeClassification').innerHTML = `<table class="kvtable">
            <tr><td>Protocol</td><td>${c.protocol || '-'}</td></tr>
            <tr><td>Family</td><td>${c.family || '-'}</td></tr>
            <tr><td>Version</td><td>${c.version || '-'}</td></tr>
            <tr><td>Confidence</td><td>${(c.confidence * 100).toFixed(1)}%</td></tr>
            <tr><td>Protocol Conf</td><td>${(c.protocolConfidence * 100).toFixed(1)}%</td></tr>
            <tr><td>Family Conf</td><td>${(c.familyConfidence * 100).toFixed(1)}%</td></tr>
            <tr><td>Version Conf</td><td>${(c.versionConfidence * 100).toFixed(1)}%</td></tr>
        </table>`;

        // 风险
        const ra = r.riskAssessment || {};
        const rScore = ra.score || 0;
        document.getElementById('analyzeRisk').innerHTML = `
            <div style="margin-bottom:12px;">
                <span style="font-size:20px;font-weight:700;color:${riskColor(rScore)}">${ra.level || '-'}</span>
                <span style="margin-left:12px;font-size:14px;">${(rScore * 100).toFixed(1)}%</span>
                <div class="risk-bar"><div class="risk-bar-fill" style="width:${rScore*100}%;background:${riskColor(rScore)}"></div></div>
            </div>
            ${(ra.factors || []).map(f => `<div style="font-size:12px;padding:4px 0;"><strong>${f.name}</strong> (${(f.weight*100).toFixed(0)}%) — ${f.description}</div>`).join('')}
            ${(ra.suggestions || []).length ? '<hr style="border-color:var(--border);margin:8px 0">' + ra.suggestions.map(s => `<div style="font-size:12px;color:var(--gray-400)">💡 ${s}</div>`).join('') : ''}
        `;

        // JA3/JA4
        const ja3 = r.ja3 || {};
        const ja4 = r.ja4 || {};
        const ja4h = r.ja4h || {};
        document.getElementById('analyzeFingerprints').innerHTML = `<table class="kvtable">
            <tr><td>JA3 Hash</td><td style="font-family:monospace;font-size:11px;word-break:break-all">${ja3.hash || '-'}</td></tr>
            <tr><td>JA3 Raw</td><td style="font-family:monospace;font-size:10px;word-break:break-all">${ja3.raw || '-'}</td></tr>
            <tr><td>JA4</td><td style="font-family:monospace;font-size:11px;word-break:break-all">${ja4.fingerprint || '-'}</td></tr>
            <tr><td>JA4H</td><td style="font-family:monospace;font-size:11px;word-break:break-all">${ja4h.fingerprint || '-'}</td></tr>
        </table>`;

        // 检测发现
        const findings = r.findings || [];
        const hints = r.defenseHints || [];
        document.getElementById('analyzeFindings').innerHTML = (findings.length ?
            findings.map(f => `<div style="padding:4px 0;font-size:12px;"><span class="badge badge-danger">${f.rule}</span> ${f.description} <span style="color:${riskColor(f.riskScore)}">(${(f.riskScore*100).toFixed(0)}%)</span></div>`).join('')
            : '<div style="color:var(--gray-400)">无威胁发现 ✅</div>'
        ) + (hints.length ?
            '<hr style="border-color:var(--border);margin:8px 0">' + hints.map(h => `<div style="font-size:12px;">🛡️ ${h}</div>`).join('')
            : '');

        // Agent
        const ag = r.agentDecision;
        const agCard = document.getElementById('analyzeAgentCard');
        if (ag) {
            agCard.style.display = 'block';
            document.getElementById('analyzeAgent').innerHTML = `<table class="kvtable">
                <tr><td>Action</td><td><span class="badge">${ag.action || '-'}</span></td></tr>
                <tr><td>Reason</td><td>${ag.reason || '-'}</td></tr>
                <tr><td>Confidence</td><td>${((ag.confidence || 0) * 100).toFixed(1)}%</td></tr>
            </table>`;
        } else {
            agCard.style.display = 'none';
        }
    } catch (e) {
        alert('分析失败: ' + e.message);
    }
}

// =====================================================================
//  ML ENGINE PAGE
// =====================================================================
async function loadMLPage() {
    populateProfileSelect('mlProfileSelect');
    try {
        const info = await API.getMLInfo();
        const layers = info.layers || [];
        const features = info.featureTypes || [];
        document.getElementById('mlModelInfo').innerHTML = `
            <div class="card"><div class="card-header"><h4>${info.architecture} <span class="badge badge-success">${info.status}</span></h4></div>
            <div class="card-body">
                <div class="arch-pipeline">
                    ${layers.map((l, i) =>
                        `<div class="arch-stage"><h5>${l.name}</h5><p>Level ${l.level}｜阈值 ${(l.threshold*100).toFixed(0)}%｜权重 ${(l.weight*100).toFixed(0)}%</p><p style="margin-top:4px;font-size:10px;">${l.description}</p></div>` +
                        (i < layers.length - 1 ? '<div class="arch-arrow">→</div>' : '')
                    ).join('')}
                </div>
                <h5 style="margin:16px 0 8px;">特征维度 (${features.length})</h5>
                <div class="feature-grid">
                    ${features.map(f => `<div class="feature-item"><span class="feature-name">${f.name}</span><span class="badge badge-info">${f.category}</span></div>`).join('')}
                </div>
            </div></div>
        `;
    } catch (e) {
        document.getElementById('mlModelInfo').innerHTML = '<div class="card"><div class="card-body" style="color:var(--gray-400)">无法加载模型信息</div></div>';
    }
}

async function runMLExtract() {
    const id = document.getElementById('mlProfileSelect').value;
    if (!id) return alert('请选择 Profile');
    try {
        const r = await API.mlExtract(id);
        const el = document.getElementById('mlExtractResult');
        el.style.display = 'block';
        document.getElementById('mlFeaturesList').innerHTML = `
            <p style="margin-bottom:8px;">${r.profileName}  —  <strong>${r.totalCount}</strong> 维特征</p>
            <div class="feature-grid">
                ${(r.features || []).map(f => `<div class="feature-item"><span class="feature-name">${f.name}</span><span class="feature-val">${typeof f.value === 'number' ? f.value.toFixed(4) : f.value}</span></div>`).join('')}
            </div>
        `;
    } catch (e) {
        alert('提取失败: ' + e.message);
    }
}

async function runMLClassify() {
    const id = document.getElementById('mlProfileSelect').value;
    if (!id) return alert('请选择 Profile');
    try {
        const r = await API.mlClassify(id);
        const el = document.getElementById('mlClassifyResult');
        el.style.display = 'block';
        const c = r.classification || {};
        document.getElementById('mlClassifyDetail').innerHTML = `
            <div style="display:flex;align-items:center;gap:16px;margin-bottom:12px;">
                <span style="font-size:20px;font-weight:700;">${c.family || '?'}</span>
                <span style="font-size:16px;color:var(--gray-400)">${c.version || '?'}</span>
                <span class="badge ${r.isHighConfidence ? 'badge-success' : 'badge-warning'}">${r.isHighConfidence ? 'High Confidence' : 'Low Confidence'}</span>
            </div>
            <table class="kvtable">
                <tr><td>Profile</td><td>${r.profileName} (${r.browser} ${r.version})</td></tr>
                <tr><td>Protocol</td><td>${c.protocol}</td></tr>
                <tr><td>Family</td><td>${c.family}</td></tr>
                <tr><td>Version</td><td>${c.version}</td></tr>
                <tr><td>Overall Confidence</td><td>${(c.confidence * 100).toFixed(1)}%</td></tr>
                <tr><td>Protocol Conf</td><td>${(c.protocolConfidence * 100).toFixed(1)}%</td></tr>
                <tr><td>Family Conf</td><td>${(c.familyConfidence * 100).toFixed(1)}%</td></tr>
                <tr><td>Version Conf</td><td>${(c.versionConfidence * 100).toFixed(1)}%</td></tr>
            </table>
        `;
    } catch (e) {
        alert('分类失败: ' + e.message);
    }
}

async function runMLBatch() {
    try {
        const r = await API.mlBatch();
        const el = document.getElementById('mlBatchResult');
        el.style.display = 'block';
        const results = r.results || [];
        const protoDist = r.protocolDistribution || {};
        const famDist = r.familyDistribution || {};
        document.getElementById('mlBatchDetail').innerHTML = `
            <div class="stats-grid" style="margin-bottom:16px;">
                <div class="stat-card"><div class="stat-icon">📊</div><div class="stat-info"><h3>Total</h3><p class="stat-value">${r.total}</p></div></div>
                <div class="stat-card"><div class="stat-icon">✅</div><div class="stat-info"><h3>High Conf Rate</h3><p class="stat-value">${r.highConfidenceRate.toFixed(1)}%</p></div></div>
            </div>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;margin-bottom:16px;">
                <div><h5>Protocol Distribution</h5>
                    ${Object.entries(protoDist).map(([k,v]) => `<div class="feature-item"><span class="feature-name">${k}</span><span class="feature-val">${v}</span></div>`).join('')}
                </div>
                <div><h5>Family Distribution</h5>
                    ${Object.entries(famDist).map(([k,v]) => `<div class="feature-item"><span class="feature-name">${k}</span><span class="feature-val">${v}</span></div>`).join('')}
                </div>
            </div>
            <table class="data-table"><thead><tr>
                <th>ID</th><th>Name</th><th>Browser</th><th>Protocol</th><th>Family</th><th>ML Version</th><th>Confidence</th><th>High</th>
            </tr></thead><tbody>
                ${results.map(r => `<tr>
                    <td style="font-size:11px">${r.id}</td><td>${r.name}</td><td>${r.browser} ${r.version}</td>
                    <td>${r.protocol}</td><td>${r.family}</td><td>${r.mlVersion}</td>
                    <td><span style="color:${riskColor(1-r.confidence)}">${(r.confidence*100).toFixed(1)}%</span></td>
                    <td>${r.highConf ? '✅' : '⚠️'}</td>
                </tr>`).join('')}
            </tbody></table>
        `;
    } catch (e) {
        alert('批量分类失败: ' + e.message);
    }
}

// =====================================================================
//  DEFENSE PAGE
// =====================================================================
async function loadDefensePage() {
    populateProfileSelect('defenseProfileSelect');
    try {
        const r = await API.getDefenseRules();
        const rules = r.detectionRules || [];
        const strategies = r.agentStrategies || [];
        const levels = r.riskLevels || [];
        document.getElementById('defenseRulesSection').innerHTML = `
            <div class="card"><div class="card-header"><h4>检测规则 (${rules.length})</h4></div>
            <div class="card-body">
                <table class="data-table"><thead><tr>
                    <th>规则</th><th>描述</th><th>特征</th><th>阈值</th><th>风险</th><th>严重度</th><th>类别</th>
                </tr></thead><tbody>
                    ${rules.map(r => `<tr>
                        <td><strong>${r.name}</strong></td><td style="font-size:12px;">${r.description}</td>
                        <td><code>${r.feature}</code></td><td>${r.threshold}</td>
                        <td style="color:${riskColor(r.riskScore)}">${(r.riskScore*100).toFixed(0)}%</td>
                        <td><span class="badge badge-${r.severity === 'critical' ? 'danger' : r.severity === 'high' ? 'warning' : 'info'}">${r.severity}</span></td>
                        <td>${r.category}</td>
                    </tr>`).join('')}
                </tbody></table>
            </div></div>
            <div class="card" style="margin-top:16px;"><div class="card-header"><h4>风险等级体系</h4></div>
            <div class="card-body" style="display:flex;gap:16px;flex-wrap:wrap;">
                ${levels.map(l => `<div style="display:flex;align-items:center;gap:8px;">
                    <div style="width:16px;height:16px;border-radius:50%;background:${l.color}"></div>
                    <span><strong>${l.level}</strong> ≥ ${(l.threshold*100).toFixed(0)}%</span>
                </div>`).join('')}
            </div></div>
            ${strategies.length ? `<div class="card" style="margin-top:16px;"><div class="card-header"><h4>🤖 Agent 策略 (${strategies.length})</h4></div>
            <div class="card-body"><table class="data-table"><thead><tr>
                <th>Name</th><th>Action</th><th>Threat</th><th>Priority</th><th>Learned</th>
            </tr></thead><tbody>
                ${strategies.map(s => `<tr><td>${s.name}</td><td><span class="badge">${s.action}</span></td><td>${s.threat}</td><td>${s.priority}</td><td>${s.learned ? '🧠' : '📜'}</td></tr>`).join('')}
            </tbody></table></div></div>` : ''}
        `;
    } catch (e) {
        document.getElementById('defenseRulesSection').innerHTML = '<div class="card"><div class="card-body" style="color:var(--gray-400)">无法加载规则</div></div>';
    }
}

async function runDefenseDetect() {
    const id = document.getElementById('defenseProfileSelect').value;
    if (!id) return alert('请选择 Profile');
    try {
        const r = await API.defenseDetect(id);
        const section = document.getElementById('defenseDetectResult');
        section.style.display = 'block';

        const d = r.detection || {};
        document.getElementById('defenseDetection').innerHTML = `
            <div style="margin-bottom:12px;">
                <span style="font-size:18px;font-weight:700;color:${d.isThreat ? '#f44336' : '#4caf50'}">${d.isThreat ? '⚠️ THREAT' : '✅ SAFE'}</span>
                <span style="margin-left:12px;">Risk: <strong style="color:${riskColor(d.riskScore)}">${d.riskLevel} (${(d.riskScore*100).toFixed(1)}%)</strong></span>
                <div class="risk-bar"><div class="risk-bar-fill" style="width:${d.riskScore*100}%;background:${riskColor(d.riskScore)}"></div></div>
            </div>
            <h5>Findings (${(d.findings||[]).length})</h5>
            ${(d.findings||[]).length ?
                (d.findings||[]).map(f => `<div style="padding:4px 0;font-size:12px;"><span class="badge badge-danger">${f.rule}</span> ${f.description} <span style="color:${riskColor(f.riskScore)}">(${(f.riskScore*100).toFixed(0)}%)</span></div>`).join('')
                : '<div style="color:var(--gray-400)">无发现</div>'
            }
        `;

        const ra = r.riskAssessment || {};
        const rScore = ra.score || 0;
        document.getElementById('defenseRiskAssess').innerHTML = `
            <div style="margin-bottom:12px;">
                <span style="font-size:18px;font-weight:700;color:${riskColor(rScore)}">${ra.level || '-'}</span>
                <span style="margin-left:8px;">${(rScore*100).toFixed(1)}%</span>
                <div class="risk-bar"><div class="risk-bar-fill" style="width:${rScore*100}%;background:${riskColor(rScore)}"></div></div>
            </div>
            ${(ra.factors || []).map(f => `<div style="font-size:12px;padding:2px 0;"><strong>${f.name}</strong> (${(f.weight*100).toFixed(0)}%) — ${f.description}</div>`).join('')}
            ${(ra.suggestions || []).length ? '<hr style="border-color:var(--border);margin:8px 0">' + ra.suggestions.map(s => `<div style="font-size:12px;">💡 ${s}</div>`).join('') : ''}
        `;

        const adv = r.defenseAdvice || {};
        const rec = adv.recommended || {};
        document.getElementById('defenseAdvice').innerHTML = `<table class="kvtable">
            <tr><td>Canvas Noise</td><td>${rec.EnableCanvasNoise ? '✅' : '❌'}</td></tr>
            <tr><td>Audio Noise</td><td>${rec.EnableAudioNoise ? '✅' : '❌'}</td></tr>
            <tr><td>WebGL Noise</td><td>${rec.EnableWebGLNoise ? '✅' : '❌'}</td></tr>
            <tr><td>Timing Noise</td><td>${rec.EnableTimingNoise ? '✅' : '❌'}</td></tr>
            <tr><td>Noise Level</td><td>${(rec.NoiseLevel*100).toFixed(0)}%</td></tr>
        </table>
        ${(adv.protection && adv.protection.AppliedMeasures) ?
            '<h5 style="margin-top:12px;">Applied Measures</h5>' + adv.protection.AppliedMeasures.map(m => `<div style="font-size:12px;">🔧 ${m}</div>`).join('')
            : ''}
        `;
    } catch (e) {
        alert('检测失败: ' + e.message);
    }
}

// =====================================================================
//  ANTI-DETECTION ENGINE PAGE
// =====================================================================
let _adSDKData = null;

async function loadAntiDetectPage() {
    populateProfileSelect('adProfileSelect');
    try {
        const s = await API.getAntiDetectStatus();
        const gens = s.generators || [];
        document.getElementById('adStatusSection').innerHTML = `
            <div class="stats-grid">
                <div class="stat-card"><div class="stat-icon">🔒</div><div class="stat-info"><h3>Enabled</h3><p class="stat-value">${s.enabled ? 'Yes' : 'No'}</p></div></div>
                <div class="stat-card"><div class="stat-icon">📋</div><div class="stat-info"><h3>Profiles</h3><p class="stat-value">${s.profileCount || 0}</p></div></div>
                <div class="stat-card"><div class="stat-icon">💉</div><div class="stat-info"><h3>Injector</h3><p class="stat-value">${s.injectorReady ? 'Ready' : 'Off'}</p></div></div>
                <div class="stat-card"><div class="stat-icon">🔗</div><div class="stat-info"><h3>Proxy</h3><p class="stat-value">${s.directProxy ? 'Direct' : (s.proxyTarget || 'None')}</p></div></div>
            </div>
            <div class="card" style="margin-top:16px;"><div class="card-header"><h4>代码生成器 (${gens.length})</h4></div>
            <div class="card-body">
                ${gens.map(g => `<div class="plugin-type-card"><h5>${g.name}</h5><p>${g.description}</p></div>`).join('')}
            </div></div>
        `;
    } catch (e) {
        document.getElementById('adStatusSection').innerHTML = '<div class="card"><div class="card-body" style="color:var(--gray-400)">无法加载反检测状态</div></div>';
    }
}

async function runAntiDetectPreview() {
    const profileId = document.getElementById('adProfileSelect').value;
    const generator = document.getElementById('adGeneratorSelect').value;
    try {
        const r = await API.antiDetectPreview(profileId, generator);
        document.getElementById('adCodePreview').style.display = 'block';
        document.getElementById('adPreviewTitle').textContent = r.generator + ' — ' + (r.profileName || '');
        document.getElementById('adCodeLength').textContent = r.codeLength + ' bytes';
        document.getElementById('adCodeContent').textContent = r.code || '// empty';
    } catch (e) {
        alert('预览失败: ' + e.message);
    }
}

async function runAntiDetectInjectTest() {
    const profileId = document.getElementById('adProfileSelect').value;
    try {
        const r = await API.antiDetectInject('', profileId);
        document.getElementById('adInjectResult').style.display = 'block';
        document.getElementById('adInjectStats').innerHTML = `
            <div style="display:flex;gap:16px;margin-bottom:8px;">
                <span>Original: <strong>${r.originalLength}</strong> bytes</span>
                <span>Injected: <strong>${r.injectedLength}</strong> bytes</span>
                <span>Delta: <strong style="color:#4caf50">+${r.deltaBytes}</strong> bytes</span>
            </div>
        `;
        document.getElementById('adInjectContent').textContent = r.injected || '';
    } catch (e) {
        alert('注入测试失败: ' + e.message);
    }
}

async function runAntiDetectSDKPreview() {
    try {
        _adSDKData = await API.getAntiDetectSDK();
        document.getElementById('adSDKResult').style.display = 'block';
        showSDKTab('core', document.querySelector('#adSDKResult .btn-sm'));
    } catch (e) {
        alert('SDK 预览失败: ' + e.message);
    }
}

function showSDKTab(tab, btn) {
    if (!_adSDKData) return;
    // 切换按钮状态
    document.querySelectorAll('#adSDKResult .btn-sm').forEach(b => b.classList.remove('active'));
    if (btn) btn.classList.add('active');
    const data = tab === 'core' ? _adSDKData.coreJS : _adSDKData.injectorJS;
    document.getElementById('adSDKContent').textContent = (data && data.code) || '// not available';
}

// =====================================================================
//  PLUGINS PAGE
// =====================================================================
async function loadPluginsPage() {
    try {
        const info = await API.getPluginsInfo();
        const types = info.pluginTypes || [];
        const arch = info.extensionArchitecture || {};
        const ifaces = arch.interfaces || [];
        const regAPI = info.registrationAPI || [];
        document.getElementById('pluginsInfoSection').innerHTML = `
            <div class="card"><div class="card-header"><h4>注册表状态</h4></div>
            <div class="card-body"><pre class="code-block">${JSON.stringify(info.registry || {}, null, 2)}</pre></div></div>

            <div class="card" style="margin-top:16px;"><div class="card-header"><h4>插件类型 (${types.length})</h4></div>
            <div class="card-body" style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
                ${types.map(t => `<div class="plugin-type-card"><h5>${t.icon} ${t.name} <span class="badge badge-info">${t.type}</span></h5><p>${t.description}</p></div>`).join('')}
            </div></div>

            <div class="card" style="margin-top:16px;"><div class="card-header"><h4>扩展管道架构</h4></div>
            <div class="card-body">
                <p style="color:var(--gray-400);margin-bottom:12px;">${arch.description || ''}</p>
                <div class="arch-pipeline">
                    ${ifaces.map((iface, i) =>
                        `<div class="arch-stage"><h5>${iface.name}</h5><p style="font-family:monospace;font-size:10px;">${iface.method}</p><p>${iface.description}</p></div>` +
                        (i < ifaces.length - 1 ? '<div class="arch-arrow">→</div>' : '')
                    ).join('')}
                </div>
            </div></div>

            <div class="card" style="margin-top:16px;"><div class="card-header"><h4>注册 API</h4></div>
            <div class="card-body">
                <table class="data-table"><thead><tr><th>Function</th><th>Description</th></tr></thead><tbody>
                    ${regAPI.map(a => `<tr><td style="font-family:monospace;font-size:12px;">${a.function}</td><td>${a.description}</td></tr>`).join('')}
                </tbody></table>
            </div></div>
        `;
    } catch (e) {
        document.getElementById('pluginsInfoSection').innerHTML = '<div class="card"><div class="card-body" style="color:var(--gray-400)">无法加载插件信息</div></div>';
    }
}

// =====================================================================
//  TOOLS PAGE
// =====================================================================
function loadToolsPage() {
    populateProfileSelect('toolsJA3ProfileSelect');
    populateProfileSelect('toolsValidateProfileSelect');
    populateProfileSelect('toolsCompareA');
    populateProfileSelect('toolsCompareB');
}

async function runToolsJA3() {
    const id = document.getElementById('toolsJA3ProfileSelect').value;
    if (!id) return alert('请选择 Profile');
    try {
        const r = await API.toolsJA3(id);
        const ja3 = r.ja3 || {};
        const ja4 = r.ja4 || {};
        document.getElementById('toolsJA3Result').innerHTML = `<table class="kvtable">
            <tr><td>JA3 Hash</td><td style="font-family:monospace;font-size:11px;word-break:break-all">${ja3.hash || '-'}</td></tr>
            <tr><td>JA3 Raw</td><td style="font-family:monospace;font-size:10px;word-break:break-all">${ja3.raw || '-'}</td></tr>
            <tr><td>JA4</td><td style="font-family:monospace;font-size:11px;word-break:break-all">${ja4.fingerprint || '-'}</td></tr>
            <tr><td>TLS Version</td><td>${r.input ? r.input.tlsVersion : '-'}</td></tr>
            <tr><td>Cipher Suites</td><td>${r.input ? r.input.cipherSuites : '-'}</td></tr>
            <tr><td>Extensions</td><td>${r.input ? r.input.extensions : '-'}</td></tr>
            <tr><td>Curves</td><td>${r.input ? r.input.curves : '-'}</td></tr>
        </table>`;
    } catch (e) {
        alert('计算失败: ' + e.message);
    }
}

async function runToolsValidate() {
    const id = document.getElementById('toolsValidateProfileSelect').value;
    if (!id) return alert('请选择 Profile');
    try {
        const r = await API.toolsValidate(id);
        const v = r.validation || {};
        document.getElementById('toolsValidateResult').innerHTML = `
            <div style="margin-bottom:8px;">
                <span style="font-size:16px;font-weight:700;color:${v.valid ? '#4caf50' : '#f44336'}">${v.valid ? '✅ VALID' : '❌ INVALID'}</span>
                <span style="margin-left:8px;">${r.profileName}</span>
            </div>
            ${(v.errors||[]).length ? '<div style="margin-bottom:6px;"><strong>Errors:</strong></div>' + v.errors.map(e => `<div style="font-size:12px;color:#f44336;">❌ ${e}</div>`).join('') : ''}
            ${(v.warnings||[]).length ? '<div style="margin:6px 0;"><strong>Warnings:</strong></div>' + v.warnings.map(w => `<div style="font-size:12px;color:#ff9800;">⚠️ ${w}</div>`).join('') : ''}
            ${(v.missingFields||[]).length ? '<div style="margin:6px 0;"><strong>Missing Fields:</strong></div>' + v.missingFields.map(m => `<div style="font-size:12px;color:var(--gray-400);">📋 ${m}</div>`).join('') : ''}
            ${r.tcpipValidation ? `<div style="margin-top:6px;font-size:12px;">TCP/IP: ${r.tcpipValidation || '✅ OK'}</div>` : ''}
        `;
    } catch (e) {
        alert('验证失败: ' + e.message);
    }
}

async function runToolsCompare() {
    const a = document.getElementById('toolsCompareA').value;
    const b = document.getElementById('toolsCompareB').value;
    if (!a || !b) return alert('请选择两个 Profile');
    if (a === b) return alert('请选择不同的 Profile');
    try {
        const r = await API.toolsCompare(a, b);
        const diffs = r.diffs || [];
        document.getElementById('toolsCompareResult').innerHTML = `
            <div style="display:flex;gap:16px;margin-bottom:12px;">
                <span style="font-size:18px;font-weight:700;">相似度: <span style="color:${riskColor(1-r.similarity)}">${(r.similarity*100).toFixed(1)}%</span></span>
                <div class="risk-bar" style="flex:1;align-self:center;"><div class="risk-bar-fill" style="width:${r.similarity*100}%;background:${riskColor(1-r.similarity)}"></div></div>
            </div>
            <table class="data-table"><thead><tr>
                <th>Field</th><th>${r.a ? r.a.name : 'A'}</th><th>${r.b ? r.b.name : 'B'}</th>
            </tr></thead><tbody>
                <tr><td>Browser</td><td>${r.a?r.a.browser:''} ${r.a?r.a.version:''}</td><td>${r.b?r.b.browser:''} ${r.b?r.b.version:''}</td></tr>
                <tr><td>OS</td><td>${r.a?r.a.os:''}</td><td>${r.b?r.b.os:''}</td></tr>
                <tr><td>TLS Version</td><td>${r.a?r.a.tlsVersion:''}</td><td>${r.b?r.b.tlsVersion:''}</td></tr>
                <tr><td>Cipher Suites</td><td>${r.a?r.a.ciphers:''}</td><td>${r.b?r.b.ciphers:''}</td></tr>
                <tr><td>Extensions</td><td>${r.a?r.a.extensions:''}</td><td>${r.b?r.b.extensions:''}</td></tr>
            </tbody></table>
            ${diffs.length ? `<h5 style="margin:12px 0 6px;">差异 (${diffs.length})</h5>
            <table class="data-table"><thead><tr><th>Field</th><th>A</th><th>B</th></tr></thead><tbody>
                ${diffs.map(d => `<tr><td>${d.field}</td><td>${d.a}</td><td>${d.b}</td></tr>`).join('')}
            </tbody></table>` : '<div style="margin-top:8px;color:#4caf50">无差异</div>'}
        `;
    } catch (e) {
        alert('对比失败: ' + e.message);
    }
}
