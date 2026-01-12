// CaddyDashboard - Main controller for Caddy instance management
class CaddyDashboard {
    constructor() {
        this.instances = [];
        this.selectedInstance = null;
        this.refreshInterval = null;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadInstances();
        this.startAutoRefresh();
    }

    setupEventListeners() {
        // Add instance button
        const addBtn = document.getElementById('add-instance-btn');
        if (addBtn) {
            addBtn.addEventListener('click', () => this.showAddInstanceModal());
        }

        // Refresh button
        const refreshBtn = document.getElementById('refresh-instances-btn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.loadInstances());
        }

        // Modal close buttons
        document.querySelectorAll('.modal-close, .modal-cancel').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const modal = e.target.closest('.modal');
                if (modal) modal.style.display = 'none';
            });
        });

        // Click outside modal to close
        window.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal')) {
                e.target.style.display = 'none';
            }
        });
    }

    async loadInstances() {
        try {
            const response = await fetch('/api/caddy/instances');
            const data = await response.json();

            if (data.instances) {
                this.instances = data.instances;
                this.renderInstances();
            }
        } catch (error) {
            console.error('Failed to load instances:', error);
            this.showError('Failed to load Caddy instances');
        }
    }

    renderInstances() {
        const container = document.getElementById('instances-container');
        if (!container) return;

        if (this.instances.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <p>No Caddy instances configured</p>
                    <button class="btn btn-primary" onclick="caddyDashboard.showAddInstanceModal()">
                        Add First Instance
                    </button>
                </div>
            `;
            return;
        }

        container.innerHTML = this.instances.map(inst => this.renderInstanceCard(inst)).join('');

        // Add click handlers for each card
        container.querySelectorAll('.instance-card').forEach(card => {
            card.addEventListener('click', (e) => {
                if (!e.target.closest('.instance-actions')) {
                    this.selectInstance(card.dataset.id);
                }
            });
        });

        // Add action button handlers
        container.querySelectorAll('[data-action]').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const action = btn.dataset.action;
                const instanceId = btn.dataset.instanceId;
                this.handleAction(action, instanceId);
            });
        });
    }

    renderInstanceCard(instance) {
        const statusClass = instance.status === 'online' ? 'status-online' :
                           instance.status === 'offline' ? 'status-offline' : 'status-unknown';
        const statusText = instance.status || 'unknown';

        return `
            <div class="instance-card ${statusClass}" data-id="${instance.id}">
                <div class="instance-header">
                    <h3 class="instance-name">${this.escapeHtml(instance.name)}</h3>
                    <span class="instance-status ${statusClass}">${statusText}</span>
                </div>
                <div class="instance-details">
                    <div class="instance-url">${this.escapeHtml(instance.url)}</div>
                    ${instance.tags && instance.tags.length > 0 ? `
                        <div class="instance-tags">
                            ${instance.tags.map(tag => `<span class="tag">${this.escapeHtml(tag)}</span>`).join('')}
                        </div>
                    ` : ''}
                </div>
                <div class="instance-actions">
                    <button class="btn btn-sm" data-action="refresh" data-instance-id="${instance.id}" title="Refresh">
                        ↻
                    </button>
                    <button class="btn btn-sm" data-action="metrics" data-instance-id="${instance.id}" title="Metrics">
                        📊
                    </button>
                    <button class="btn btn-sm" data-action="config" data-instance-id="${instance.id}" title="Config">
                        ⚙️
                    </button>
                    <button class="btn btn-sm btn-danger" data-action="delete" data-instance-id="${instance.id}" title="Delete">
                        🗑️
                    </button>
                </div>
            </div>
        `;
    }

    selectInstance(instanceId) {
        this.selectedInstance = this.instances.find(i => i.id === instanceId);
        if (this.selectedInstance) {
            this.showInstanceDetails(this.selectedInstance);
        }
    }

    showInstanceDetails(instance) {
        // Update details panel or navigate to details page
        window.location.href = `/caddy/instances/${instance.id}`;
    }

    async handleAction(action, instanceId) {
        switch (action) {
            case 'refresh':
                await this.refreshInstance(instanceId);
                break;
            case 'metrics':
                window.location.href = `/caddy/instances/${instanceId}/analytics`;
                break;
            case 'config':
                window.location.href = `/caddy/instances/${instanceId}/config`;
                break;
            case 'delete':
                if (confirm('Are you sure you want to delete this instance?')) {
                    await this.deleteInstance(instanceId);
                }
                break;
        }
    }

    async refreshInstance(instanceId) {
        try {
            const response = await fetch(`/api/caddy/instances/${instanceId}/refresh`, {
                method: 'POST'
            });
            if (response.ok) {
                this.loadInstances();
            }
        } catch (error) {
            console.error('Failed to refresh instance:', error);
            this.showError('Failed to refresh instance');
        }
    }

    async deleteInstance(instanceId) {
        try {
            const response = await fetch(`/api/caddy/instances/${instanceId}`, {
                method: 'DELETE'
            });
            if (response.ok) {
                this.loadInstances();
            } else {
                const error = await response.json();
                this.showError(error.error || 'Failed to delete instance');
            }
        } catch (error) {
            console.error('Failed to delete instance:', error);
            this.showError('Failed to delete instance');
        }
    }

    showAddInstanceModal() {
        const modal = document.getElementById('add-instance-modal');
        if (modal) {
            modal.style.display = 'block';
            modal.querySelector('form').reset();
        }
    }

    async addInstance(formData) {
        try {
            const response = await fetch('/api/caddy/instances', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formData)
            });

            if (response.ok) {
                this.loadInstances();
                this.hideModals();
            } else {
                const error = await response.json();
                this.showError(error.error || 'Failed to add instance');
            }
        } catch (error) {
            console.error('Failed to add instance:', error);
            this.showError('Failed to add instance');
        }
    }

    hideModals() {
        document.querySelectorAll('.modal').forEach(modal => {
            modal.style.display = 'none';
        });
    }

    startAutoRefresh() {
        // Refresh every 60 seconds
        this.refreshInterval = setInterval(() => {
            this.loadInstances();
        }, 60000);
    }

    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    showError(message) {
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

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Instance management functions for form handling
document.addEventListener('DOMContentLoaded', () => {
    window.caddyDashboard = new CaddyDashboard();

    // Handle add instance form submission
    const addInstanceForm = document.getElementById('add-instance-form');
    if (addInstanceForm) {
        addInstanceForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = {
                name: document.getElementById('instance-name').value,
                url: document.getElementById('instance-url').value,
                api_key_file: document.getElementById('instance-api-key-file').value,
                tags: document.getElementById('instance-tags').value.split(',').map(t => t.trim()).filter(t => t)
            };
            await window.caddyDashboard.addInstance(formData);
        });
    }
});
