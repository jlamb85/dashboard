// Modern Dashboard JavaScript v2.0
// Bootstrap 5.3 compatible

'use strict';

// ============================================
// Theme Management (Dark Mode)
// ============================================

const ThemeManager = {
    STORAGE_KEY: 'dashboard-theme',
    LIGHT: 'light',
    DARK: 'dark',

    init() {
        this.setup();
        this.attachEventListeners();
    },

    setup() {
        const saved = localStorage.getItem(this.STORAGE_KEY);
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        const theme = saved || (prefersDark ? this.DARK : this.LIGHT);
        this.set(theme);
    },

    set(theme) {
        document.documentElement.setAttribute('data-bs-theme', theme);
        localStorage.setItem(this.STORAGE_KEY, theme);
        this.updateToggleButton(theme);
    },

    toggle() {
        const current = document.documentElement.getAttribute('data-bs-theme');
        const next = current === this.LIGHT ? this.DARK : this.LIGHT;
        this.set(next);
    },

    updateToggleButton(theme) {
        const btn = document.getElementById('theme-toggle');
        if (btn) {
            const icon = btn.querySelector('i');
            if (icon) {
                icon.className = theme === this.DARK ? 'bi bi-sun' : 'bi bi-moon-stars';
            }
        }
    },

    attachEventListeners() {
        const btn = document.getElementById('theme-toggle');
        if (btn) {
            btn.addEventListener('click', () => this.toggle());
        }
    }
};

// ============================================
// Sidebar Collapse Management
// ============================================

const SidebarManager = {
    STORAGE_KEY: 'dashboard-sidebar-state',
    sidebar: null,
    mainContent: null,
    toggleBtn: null,

    init() {
        this.sidebar = document.getElementById('sidebar-nav');
        this.mainContent = document.getElementById('main-content');
        this.toggleBtn = document.getElementById('sidebar-toggle');

        if (this.sidebar && this.mainContent && this.toggleBtn) {
            this.setup();
            this.attachEventListeners();
        }
    },

    setup() {
        const saved = localStorage.getItem(this.STORAGE_KEY);
        if (saved === 'collapsed') {
            this.collapse(false); // false = no animation on page load
        }
    },

    toggle() {
        if (this.sidebar.classList.contains('collapsed')) {
            this.expand();
        } else {
            this.collapse(true);
        }
    },

    collapse(animate = true) {
        if (!animate) {
            this.sidebar.style.transition = 'none';
            this.mainContent.style.transition = 'none';
        }

        this.sidebar.classList.add('collapsed');
        this.mainContent.classList.add('expanded');
        localStorage.setItem(this.STORAGE_KEY, 'collapsed');
        this.updateToggleButton(true);

        if (!animate) {
            // Force reflow then restore transitions
            void this.sidebar.offsetHeight;
            this.sidebar.style.transition = '';
            this.mainContent.style.transition = '';
        }
    },

    expand() {
        this.sidebar.classList.remove('collapsed');
        this.mainContent.classList.remove('expanded');
        localStorage.setItem(this.STORAGE_KEY, 'expanded');
        this.updateToggleButton(false);
    },

    updateToggleButton(isCollapsed) {
        const icon = this.toggleBtn.querySelector('i');
        if (icon) {
            icon.className = isCollapsed ? 'bi bi-layout-sidebar-inset-reverse' : 'bi bi-layout-sidebar-inset';
        }
        this.toggleBtn.title = isCollapsed ? 'Show sidebar' : 'Hide sidebar';
    },

    attachEventListeners() {
        this.toggleBtn.addEventListener('click', () => this.toggle());
    }
};

// ============================================
// Navigation Highlighting
// ============================================

const Navigation = {
    init() {
        this.highlightCurrentPage();
        this.attachLinkListeners();
    },

    highlightCurrentPage() {
        const currentPath = window.location.pathname;
        document.querySelectorAll('.nav-link').forEach(link => {
            const href = link.getAttribute('href');
            if ((href === '/' && currentPath === '/') || 
                (href !== '/' && currentPath.startsWith(href))) {
                link.classList.add('active');
            } else {
                link.classList.remove('active');
            }
        });
    },

    attachLinkListeners() {
        // Close mobile menu on link click
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', () => {
                const navbar = document.querySelector('.navbar-collapse');
                if (navbar && navbar.classList.contains('show')) {
                    document.querySelector('.navbar-toggler').click();
                }
            });
        });
    }
};

// ============================================
// Status Indicator
// ============================================

const StatusMonitor = {
    CHECK_INTERVAL: 30000, // 30 seconds
    indicator: null,

    init() {
        this.indicator = document.getElementById('status-indicator');
        if (this.indicator) {
            this.checkHealth();
            setInterval(() => this.checkHealth(), this.CHECK_INTERVAL);
        }
    },

    checkHealth() {
        fetch('/health')
            .then(response => {
                if (response.ok) {
                    this.setStatus('Active', 'success');
                } else {
                    this.setStatus('Degraded', 'warning');
                }
            })
            .catch(() => {
                this.setStatus('Offline', 'danger');
            });
    },

    setStatus(text, type) {
        if (this.indicator) {
            this.indicator.textContent = text;
            this.indicator.className = `badge bg-${type}`;
        }
    }
};

// ============================================
// Clickable Rows
// ============================================

const TableInteraction = {
    init() {
        this.makeRowsClickable();
    },

    makeRowsClickable() {
        document.querySelectorAll('[data-href]').forEach(row => {
            row.style.cursor = 'pointer';
            row.addEventListener('click', (e) => {
                // Don't navigate if clicking a button or link inside the row
                if (!e.target.closest('button, a')) {
                    window.location.href = row.getAttribute('data-href');
                }
            });
            row.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    window.location.href = row.getAttribute('data-href');
                }
            });
        });
    }
};

// ============================================
// Toast Notifications
// ============================================

const Notifications = {
    show(message, type = 'info', duration = 3000) {
        const container = document.getElementById('notification-container') || this.createContainer();
        
        const toast = document.createElement('div');
        toast.className = `alert alert-${type} alert-dismissible fade show`;
        toast.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        
        container.appendChild(toast);
        
        if (duration) {
            setTimeout(() => toast.remove(), duration);
        }
    },

    createContainer() {
        const container = document.createElement('div');
        container.id = 'notification-container';
        container.style.cssText = 'position: fixed; top: 80px; right: 20px; z-index: 1050; width: 350px; max-width: calc(100% - 20px);';
        document.body.appendChild(container);
        return container;
    }
};

// ============================================
// Loading States
// ============================================

const LoadingState = {
    show(element) {
        element.classList.add('loading');
    },

    hide(element) {
        element.classList.remove('loading');
    }
};

// ============================================
// Monitoring Control
// ============================================

const MonitoringControl = {
    init() {
        this.attachEventListeners();
        this.updateStatus();
        // Auto-refresh status every 30 seconds
        setInterval(() => this.updateStatus(), 30000);
    },

    async updateStatus() {
        try {
            const response = await fetch('/api/monitoring/status');
            if (!response.ok) {
                throw new Error('Failed to fetch monitoring status');
            }
            const data = await response.json();
            this.updateUI(data.active);
        } catch (error) {
            console.error('Error fetching monitoring status:', error);
            this.updateUI(null); // Unknown status
        }
    },

    updateUI(isActive) {
        const statusBadge = document.getElementById('monitoring-status');
        const startBtn = document.getElementById('start-monitoring');
        const stopBtn = document.getElementById('stop-monitoring');
        const restartBtn = document.getElementById('restart-monitoring');
        
        // Dashboard card elements
        const dashboardBadge = document.getElementById('dashboard-monitoring-status');
        const dashboardText = document.getElementById('dashboard-monitoring-text');
        const dashboardDesc = document.getElementById('dashboard-monitoring-desc');

        if (!statusBadge) return;

        // Restore button HTML if needed
        if (startBtn && !startBtn.innerHTML.includes('Start')) {
            startBtn.innerHTML = '<i class="bi bi-play-fill"></i> Start';
        }
        if (stopBtn && !stopBtn.innerHTML.includes('Stop')) {
            stopBtn.innerHTML = '<i class="bi bi-stop-fill"></i> Stop';
        }
        if (restartBtn && !restartBtn.innerHTML.includes('Restart')) {
            restartBtn.innerHTML = '<i class="bi bi-arrow-clockwise"></i> Restart';
        }

        if (isActive === null) {
            statusBadge.textContent = 'Unknown';
            statusBadge.className = 'badge bg-secondary';
            if (dashboardBadge) dashboardBadge.className = 'badge bg-secondary';
            if (dashboardText) dashboardText.textContent = 'Unknown';
            if (dashboardDesc) dashboardDesc.textContent = 'Status unavailable';
            if (startBtn) startBtn.disabled = true;
            if (stopBtn) stopBtn.disabled = true;
            if (restartBtn) restartBtn.disabled = true;
        } else if (isActive) {
            statusBadge.textContent = 'Running';
            statusBadge.className = 'badge bg-success';
            if (dashboardBadge) dashboardBadge.className = 'badge bg-success';
            if (dashboardText) dashboardText.textContent = 'Active';
            if (dashboardDesc) dashboardDesc.textContent = 'Real-time monitoring enabled';
            if (startBtn) startBtn.disabled = true;
            if (stopBtn) stopBtn.disabled = false;
            if (restartBtn) restartBtn.disabled = false;
        } else {
            statusBadge.textContent = 'Stopped';
            statusBadge.className = 'badge bg-danger';
            if (dashboardBadge) dashboardBadge.className = 'badge bg-danger';
            if (dashboardText) dashboardText.textContent = 'Stopped';
            if (dashboardDesc) dashboardDesc.textContent = 'Monitoring is currently disabled';
            if (startBtn) startBtn.disabled = false;
            if (stopBtn) stopBtn.disabled = true;
            if (restartBtn) restartBtn.disabled = true;
        }
    },

    async performAction(action, actionName) {
        try {
            const response = await fetch(`/api/monitoring/${action}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`Failed to ${action} monitoring`);
            }

            const data = await response.json();
            
            if (data.success) {
                Notifications.show(data.message, 'success');
            } else {
                Notifications.show(data.message, 'danger');
            }
            
            // Update UI with the new state
            await this.updateStatus();
        } catch (error) {
            console.error(`Error ${actionName} monitoring:`, error);
            Notifications.show(`Failed to ${action} monitoring`, 'danger');
            await this.updateStatus(); // Refresh status even on error
        }
    },

    attachEventListeners() {
        const startBtn = document.getElementById('start-monitoring');
        const stopBtn = document.getElementById('stop-monitoring');
        const restartBtn = document.getElementById('restart-monitoring');

        if (startBtn) {
            startBtn.addEventListener('click', () => this.performAction('start', 'Starting'));
        }
        if (stopBtn) {
            stopBtn.addEventListener('click', () => this.performAction('stop', 'Stopping'));
        }
        if (restartBtn) {
            restartBtn.addEventListener('click', () => this.performAction('restart', 'Restarting'));
        }
    }
};

// ============================================
// Monitoring Features Collapse Persistence
// ============================================

const MonitoringFeaturesToggle = {
    STORAGE_KEY: 'monitoring-features-collapse',

    init() {
        const collapseEl = document.getElementById('monitoring-features');
        const toggleBtn = document.querySelector('[data-bs-target="#monitoring-features"]');
        if (!collapseEl || !toggleBtn || !window.bootstrap) return;

        // Apply saved state
        const saved = (() => {
            try { return localStorage.getItem(this.STORAGE_KEY); } catch (_) { return null; }
        })();
        const collapse = new bootstrap.Collapse(collapseEl, { toggle: false });
        if (saved === 'collapsed') {
            collapse.hide();
            toggleBtn.setAttribute('aria-expanded', 'false');
        } else {
            collapse.show();
            toggleBtn.setAttribute('aria-expanded', 'true');
        }

        collapseEl.addEventListener('hidden.bs.collapse', () => {
            try { localStorage.setItem(this.STORAGE_KEY, 'collapsed'); } catch (_) {}
            toggleBtn.setAttribute('aria-expanded', 'false');
        });
        collapseEl.addEventListener('shown.bs.collapse', () => {
            try { localStorage.setItem(this.STORAGE_KEY, 'expanded'); } catch (_) {}
            toggleBtn.setAttribute('aria-expanded', 'true');
        });
    }
};

// ============================================
// Initialization
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    ThemeManager.init();
    SidebarManager.init();
    Navigation.init();
    StatusMonitor.init();
    TableInteraction.init();
    MonitoringControl.init();
    MonitoringFeaturesToggle.init();
    
    // Global error handler
    window.addEventListener('error', (e) => {
        console.error('Error:', e.error);
        Notifications.show('An error occurred. Please try again.', 'danger');
    });
});

// Export for external use
window.Dashboard = {
    Notifications,
    LoadingState,
    ThemeManager,
    SidebarManager,
    MonitoringControl
};