// ConfigEditor - Caddy configuration editor
class ConfigEditor {
    constructor() {
        this.instanceId = this.getInstanceId();
        this.currentFormat = 'json';
        this.originalConfig = '';
        this.unsavedChanges = false;

        if (!this.instanceId) {
            this.showToast('No instance specified', 'error');
            return;
        }

        this.init();
    }

    getInstanceId() {
        // Try to get from URL path
        const pathMatch = window.location.pathname.match(/\/caddy\/instances\/([^/]+)\/config/);
        if (pathMatch) return pathMatch[1];

        // Check for data attribute
        const elem = document.getElementById('config-editor');
        if (elem && elem.dataset.instanceId) return elem.dataset.instanceId;

        return null;
    }

    async init() {
        this.setupEditor();
        await this.loadInstance();
        await this.loadConfig();
        await this.loadSites();
        this.setupAutoRefresh();
    }

    setupEditor() {
        const editor = document.getElementById('config-editor');
        const lineNumbers = document.getElementById('line-numbers');

        // Tab key support
        editor.addEventListener('keydown', (e) => {
            if (e.key === 'Tab') {
                e.preventDefault();
                const start = editor.selectionStart;
                const end = editor.selectionEnd;
                editor.value = editor.value.substring(0, start) + '    ' + editor.value.substring(end);
                editor.selectionStart = editor.selectionEnd = start + 4;
                this.updateLineNumbers();
                this.markUnsaved();
            }
        });

        // Detect unsaved changes
        editor.addEventListener('input', () => {
            this.markUnsaved();
        });

        // Update line numbers on scroll
        editor.addEventListener('scroll', () => {
            lineNumbers.scrollTop = editor.scrollTop;
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                this.saveConfig();
            }
        });

        this.updateLineNumbers();
    }

    updateLineNumbers() {
        const editor = document.getElementById('config-editor');
        const lineNumbers = document.getElementById('line-numbers');
        const lines = editor.value.split('\n').length;
        lineNumbers.innerHTML = Array.from({ length: lines }, (_, i) => i + 1).join('\n');
    }

    async loadInstance() {
        try {
            const response = await fetch(`/api/caddy/instances/${this.instanceId}`);
            const data = await response.json();

            if (data.instance) {
                document.getElementById('instance-name').textContent = data.instance.name;

                const statusDot = document.querySelector('.status-dot');
                const statusText = document.getElementById('instance-status');

                statusDot.className = `status-dot ${data.instance.status || 'unknown'}`;
                statusText.textContent = data.instance.status === 'online' ? 'Online' :
                                         data.instance.status === 'offline' ? 'Offline' : 'Unknown';
            }
        } catch (error) {
            console.error('Failed to load instance:', error);
        }
    }

    async loadConfig(format = 'json') {
        try {
            const endpoint = format === 'caddyfile'
                ? `/api/caddy/instances/${this.instanceId}/config/caddyfile`
                : `/api/caddy/instances/${this.instanceId}/config`;

            const response = await fetch(endpoint);
            if (!response.ok) {
                throw new Error('Failed to load config');
            }

            const config = await response.text();
            this.originalConfig = config;
            document.getElementById('config-editor').value = config;
            this.updateLineNumbers();
            this.unsavedChanges = false;
            this.currentFormat = format;

            document.getElementById('config-format').value = format;
        } catch (error) {
            console.error('Failed to load config:', error);
            this.showToast('Failed to load configuration', 'error');
        }
    }

    async loadSites() {
        try {
            const response = await fetch(`/api/caddy/instances/${this.instanceId}/sites`);
            const sites = await response.json();

            const container = document.getElementById('sites-container');
            if (sites && sites.length > 0) {
                container.innerHTML = sites.map(site => `
                    <div class="site-item" onclick="navigateToSite('${site.name}')">
                        <div class="site-name">${this.escapeHtml(site.name)}</div>
                        <div class="site-address">${site.listen ? site.listen.join(', ') : 'No addresses'}</div>
                    </div>
                `).join('');
            } else {
                container.innerHTML = '<p style="color: #64748b; font-size: 0.9rem;">No sites configured</p>';
            }
        } catch (error) {
            console.error('Failed to load sites:', error);
        }
    }

    async saveConfig() {
        const config = document.getElementById('config-editor').value;

        try {
            const response = await fetch(`/api/caddy/instances/${this.instanceId}/reload`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ config })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Failed to save config');
            }

            this.originalConfig = config;
            this.unsavedChanges = false;
            this.showToast('Configuration saved and reloaded', 'success');
        } catch (error) {
            console.error('Failed to save config:', error);
            this.showToast(error.message, 'error');
        }
    }

    async validateConfig() {
        const config = document.getElementById('config-editor').value;

        try {
            // Try to reload - if it succeeds, config is valid
            const response = await fetch(`/api/caddy/instances/${this.instanceId}/reload`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ config }),
                signal: AbortSignal.timeout(5000) // 5 second timeout
            });

            if (response.ok) {
                this.showToast('Configuration is valid', 'success');
            } else {
                const error = await response.json();
                this.showToast(`Validation failed: ${error.error}`, 'error');
            }
        } catch (error) {
            if (error.name === 'TimeoutError') {
                this.showToast('Validation timed out', 'error');
            } else {
                this.showToast('Validation failed', 'error');
            }
        }
    }

    formatConfig() {
        const editor = document.getElementById('config-editor');
        const config = editor.value;

        try {
            // Try to parse as JSON and reformat
            const parsed = JSON.parse(config);
            editor.value = JSON.stringify(parsed, null, 4);
            this.updateLineNumbers();
            this.markUnsaved();
        } catch (e) {
            // Not JSON, try basic formatting
            const lines = config.split('\n');
            const formatted = lines.map(line => line.trim()).join('\n');
            editor.value = formatted;
            this.updateLineNumbers();
            this.markUnsaved();
        }
    }

    copyConfig() {
        const config = document.getElementById('config-editor').value;
        navigator.clipboard.writeText(config).then(() => {
            this.showToast('Configuration copied to clipboard', 'success');
        }).catch(() => {
            this.showToast('Failed to copy', 'error');
        });
    }

    resetConfig() {
        if (this.unsavedChanges) {
            if (!confirm('Discard unsaved changes?')) return;
        }

        document.getElementById('config-editor').value = this.originalConfig;
        this.updateLineNumbers();
        this.unsavedChanges = false;
        this.showToast('Configuration reset', 'info');
    }

    async reloadConfig() {
        this.showToast('Reloading configuration...', 'info');

        try {
            const response = await fetch(`/api/caddy/instances/${this.instanceId}/reload`, {
                method: 'POST'
            });

            if (response.ok) {
                this.showToast('Configuration reloaded', 'success');
                await this.loadConfig(this.currentFormat);
            } else {
                const error = await response.json();
                this.showToast(`Reload failed: ${error.error}`, 'error');
            }
        } catch (error) {
            console.error('Reload failed:', error);
            this.showToast('Reload failed', 'error');
        }
    }

    async restartServer() {
        if (!confirm('Restart the Caddy server? This may cause a brief downtime.')) return;

        this.showToast('Restarting server...', 'info');

        try {
            const response = await fetch(`/api/caddy/instances/${this.instanceId}/restart`, {
                method: 'POST'
            });

            const result = await response.json();
            this.showToast(result.message || 'Server restart initiated', 'success');

            // Wait a moment then check status
            setTimeout(() => this.loadInstance(), 3000);
        } catch (error) {
            console.error('Restart failed:', error);
            this.showToast('Restart failed', 'error');
        }
    }

    viewLogs() {
        window.open(`/caddy/instances/${this.instanceId}/logs`, '_blank');
    }

    exportConfig() {
        const config = document.getElementById('config-editor').value;
        const blob = new Blob([config], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `caddy-config-${this.instanceId}-${new Date().toISOString().split('T')[0]}.json`;
        a.click();
        URL.revokeObjectURL(url);
        this.showToast('Configuration exported', 'success');
    }

    async switchFormat() {
        const format = document.getElementById('config-format').value;
        if (this.unsavedChanges) {
            if (!confirm('Switch format? Unsaved changes will be lost.')) {
                document.getElementById('config-format').value = this.currentFormat;
                return;
            }
        }
        await this.loadConfig(format);
    }

    markUnsaved() {
        const title = document.querySelector('.page-title');
        if (!title.textContent.includes('*')) {
            title.textContent += ' *';
        }
        this.unsavedChanges = true;
    }

    setupAutoRefresh() {
        // Refresh status every 30 seconds
        setInterval(() => {
            this.loadInstance();
        }, 30000);
    }

    showToast(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        document.body.appendChild(toast);

        setTimeout(() => {
            toast.remove();
        }, 3000);
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Helper function to navigate to site
function navigateToSite(siteName) {
    // Could open site config in editor
    console.log('Navigate to site:', siteName);
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.configEditor = new ConfigEditor();
});
