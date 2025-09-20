// Dashboard JavaScript
class Dashboard {
    constructor() {
        this.refreshInterval = null;
        this.charts = new Map();
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadDashboardData();
        this.startAutoRefresh();
    }

    setupEventListeners() {
        // Refresh buttons
        document.querySelectorAll('.widget-refresh').forEach(button => {
            button.addEventListener('click', (e) => {
                const widget = e.target.closest('.widget');
                const widgetId = widget.dataset.widgetId;
                this.refreshWidget(widgetId);
            });
        });

        // Global refresh
        document.addEventListener('keydown', (e) => {
            if (e.key === 'r' && (e.ctrlKey || e.metaKey)) {
                e.preventDefault();
                this.loadDashboardData();
            }
        });
    }

    async loadDashboardData() {
        try {
            const response = await fetch('/api/dashboard');
            const data = await response.json();
            this.renderDashboard(data);
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            this.showError('Failed to load dashboard data');
        }
    }

    renderDashboard(data) {
        // Update page title
        if (data.title) {
            document.title = data.title;
        }

        // Render system stats
        this.renderSystemStats(data.stats);

        // Render widgets
        if (data.widgets) {
            data.widgets.forEach(widget => {
                this.renderWidget(widget);
            });
        }

        // Update last refresh time
        this.updateLastRefresh(data.last_update);
    }

    renderSystemStats(stats) {
        const statsContainer = document.getElementById('system-stats');
        if (!statsContainer || !stats) return;

        const statsHtml = `
            <div class="stat-card">
                <div class="stat-label">CPU Usage</div>
                <div class="stat-value">${Math.round(stats.cpu_usage)}%</div>
                <div class="stat-change neutral">Current</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Memory Usage</div>
                <div class="stat-value">${Math.round(stats.memory_usage)} MB</div>
                <div class="stat-change positive">+2.1%</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Disk Usage</div>
                <div class="stat-value">${Math.round(stats.disk_usage)}%</div>
                <div class="stat-change neutral">Stable</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Uptime</div>
                <div class="stat-value">${this.formatUptime(stats.uptime)}</div>
                <div class="stat-change positive">Running</div>
            </div>
        `;

        statsContainer.innerHTML = statsHtml;
    }

    renderWidget(widget) {
        const element = document.getElementById(`widget-${widget.id}`);
        if (!element) return;

        const content = element.querySelector('.widget-content');
        if (!content) return;

        switch (widget.type) {
            case 'chart':
                this.renderChart(widget, content);
                break;
            case 'table':
                this.renderTable(widget, content);
                break;
            case 'metric':
                this.renderMetric(widget, content);
                break;
            case 'activity':
                this.renderActivity(widget, content);
                break;
            case 'progress':
                this.renderProgress(widget, content);
                break;
            case 'text':
                this.renderText(widget, content);
                break;
        }
    }

    renderChart(widget, container) {
        const canvasId = `chart-${widget.id}`;
        container.innerHTML = `<canvas id="${canvasId}" class="chart-canvas"></canvas>`;

        const canvas = document.getElementById(canvasId);
        const ctx = canvas.getContext('2d');

        // Simple line chart implementation
        this.drawLineChart(ctx, widget.data);
    }

    drawLineChart(ctx, data) {
        const canvas = ctx.canvas;
        const width = canvas.width = canvas.offsetWidth;
        const height = canvas.height = canvas.offsetHeight;
        
        ctx.clearRect(0, 0, width, height);
        
        if (!data.datasets || !data.datasets[0] || !data.datasets[0].data) return;

        const dataset = data.datasets[0];
        const values = dataset.data;
        const max = Math.max(...values);
        const min = Math.min(...values);
        const range = max - min || 1;

        // Draw axes
        ctx.strokeStyle = '#e2e8f0';
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(40, height - 40);
        ctx.lineTo(width - 20, height - 40);
        ctx.moveTo(40, height - 40);
        ctx.lineTo(40, 20);
        ctx.stroke();

        // Draw data line
        ctx.strokeStyle = dataset.borderColor || '#3b82f6';
        ctx.lineWidth = 2;
        ctx.beginPath();

        values.forEach((value, index) => {
            const x = 40 + (index * (width - 60) / (values.length - 1));
            const y = height - 40 - ((value - min) / range) * (height - 60);
            
            if (index === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });

        ctx.stroke();

        // Draw data points
        ctx.fillStyle = dataset.borderColor || '#3b82f6';
        values.forEach((value, index) => {
            const x = 40 + (index * (width - 60) / (values.length - 1));
            const y = height - 40 - ((value - min) / range) * (height - 60);
            
            ctx.beginPath();
            ctx.arc(x, y, 3, 0, 2 * Math.PI);
            ctx.fill();
        });
    }

    renderTable(widget, container) {
        const data = widget.data;
        if (!data.headers || !data.rows) return;

        const tableHtml = `
            <table class="data-table">
                <thead>
                    <tr>
                        ${data.headers.map(header => `<th>${header}</th>`).join('')}
                    </tr>
                </thead>
                <tbody>
                    ${data.rows.map(row => 
                        `<tr>${row.map(cell => `<td>${cell}</td>`).join('')}</tr>`
                    ).join('')}
                </tbody>
            </table>
        `;

        container.innerHTML = tableHtml;
    }

    renderMetric(widget, container) {
        const data = widget.data;
        const trendClass = data.trend === 'up' ? 'positive' : 
                          data.trend === 'down' ? 'negative' : 'neutral';

        const metricHtml = `
            <div class="metric">
                <div class="metric-value">
                    ${data.value}
                    ${data.unit ? `<span class="metric-unit">${data.unit}</span>` : ''}
                </div>
                ${data.description ? `<div class="metric-label">${data.description}</div>` : ''}
                ${data.change ? `<div class="metric-trend ${trendClass}">${data.change > 0 ? '+' : ''}${data.change}%</div>` : ''}
            </div>
        `;

        container.innerHTML = metricHtml;
    }

    renderActivity(widget, container) {
        const data = widget.data;
        if (!data.items) return;

        const activitiesHtml = data.items.map(item => `
            <div class="activity-item">
                <div class="activity-icon ${item.type}">
                    ${this.getActivityIcon(item.type)}
                </div>
                <div class="activity-content">
                    <div class="activity-title">${item.title}</div>
                    ${item.description ? `<div class="activity-description">${item.description}</div>` : ''}
                    <div class="activity-time">${this.formatTime(item.timestamp)}</div>
                </div>
            </div>
        `).join('');

        container.innerHTML = `<div class="activity-feed">${activitiesHtml}</div>`;
    }

    renderProgress(widget, container) {
        const data = widget.data;
        const percentage = (data.value / data.max) * 100;

        const progressHtml = `
            <div class="progress-label">
                <span>${data.label}</span>
                <span>${Math.round(percentage)}%</span>
            </div>
            <div class="progress">
                <div class="progress-bar" style="width: ${percentage}%"></div>
            </div>
            ${data.description ? `<div class="progress-description">${data.description}</div>` : ''}
        `;

        container.innerHTML = progressHtml;
    }

    renderText(widget, container) {
        const data = widget.data;
        
        const textHtml = Object.entries(data).map(([key, value]) => `
            <div style="margin-bottom: 0.5rem;">
                <strong>${key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}:</strong> ${value}
            </div>
        `).join('');

        container.innerHTML = textHtml;
    }

    getActivityIcon(type) {
        const icons = {
            success: '✓',
            warning: '⚠',
            error: '✕',
            info: 'ℹ'
        };
        return icons[type] || '•';
    }

    formatTime(timestamp) {
        const date = new Date(timestamp);
        const now = new Date();
        const diff = now - date;

        if (diff < 60000) return 'Just now';
        if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
        if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
        return `${Math.floor(diff / 86400000)}d ago`;
    }

    formatUptime(seconds) {
        const days = Math.floor(seconds / 86400);
        const hours = Math.floor((seconds % 86400) / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);

        if (days > 0) return `${days}d ${hours}h`;
        if (hours > 0) return `${hours}h ${minutes}m`;
        return `${minutes}m`;
    }

    updateLastRefresh(timestamp) {
        const element = document.getElementById('last-refresh');
        if (element) {
            const date = new Date(timestamp);
            element.textContent = `Last updated: ${date.toLocaleTimeString()}`;
        }
    }

    startAutoRefresh() {
        // Refresh every 30 seconds
        this.refreshInterval = setInterval(() => {
            this.loadDashboardData();
        }, 30000);
    }

    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    refreshWidget(widgetId) {
        // Refresh a specific widget
        const widget = document.getElementById(`widget-${widgetId}`);
        if (widget) {
            widget.classList.add('refreshing');
            setTimeout(() => {
                widget.classList.remove('refreshing');
                this.loadDashboardData();
            }, 500);
        }
    }

    showError(message) {
        // Simple error notification
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-notification';
        errorDiv.textContent = message;
        errorDiv.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: #fee2e2;
            color: #dc2626;
            padding: 1rem;
            border-radius: 6px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            z-index: 1000;
        `;
        
        document.body.appendChild(errorDiv);
        
        setTimeout(() => {
            errorDiv.remove();
        }, 5000);
    }
}

// Initialize dashboard when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    if (document.getElementById('dashboard')) {
        window.dashboard = new Dashboard();
    }
});