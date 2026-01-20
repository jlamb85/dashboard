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
// Initialization
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    ThemeManager.init();
    SidebarManager.init();
    Navigation.init();
    StatusMonitor.init();
    TableInteraction.init();
    
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
    SidebarManager
};