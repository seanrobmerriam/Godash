// MetricsChart - Chart rendering for Caddy analytics
class MetricsChart {
    constructor(containerId, options = {}) {
        this.container = document.getElementById(containerId);
        if (!this.container) {
            console.error(`Container not found: ${containerId}`);
            return;
        }

        this.options = {
            type: options.type || 'line',
            width: options.width || this.container.offsetWidth || 400,
            height: options.height || 200,
            color: options.color || '#3b82f6',
            fillColor: options.fillColor || 'rgba(59, 130, 246, 0.1)',
            animate: options.animate !== false,
            showPoints: options.showPoints !== false,
            ...options
        };

        this.canvas = document.createElement('canvas');
        this.canvas.width = this.options.width;
        this.canvas.height = this.options.height;
        this.container.appendChild(this.canvas);

        this.ctx = this.canvas.getContext('2d');
    }

    // Draw a line chart
    drawLineChart(data, options = {}) {
        if (!data || data.length === 0) {
            this.drawEmptyState('No data available');
            return;
        }

        const ctx = this.ctx;
        const width = this.canvas.width;
        const height = this.canvas.height;
        const padding = { top: 20, right: 20, bottom: 40, left: 50 };

        const chartWidth = width - padding.left - padding.right;
        const chartHeight = height - padding.top - padding.bottom;

        // Clear canvas
        ctx.clearRect(0, 0, width, height);

        // Find min/max values
        const values = data.map(d => d.value);
        const min = Math.min(...values) * 0.9;
        const max = Math.max(...values) * 1.1;
        const range = max - min || 1;

        // Draw grid lines
        ctx.strokeStyle = '#e2e8f0';
        ctx.lineWidth = 1;

        const gridLines = 5;
        for (let i = 0; i <= gridLines; i++) {
            const y = padding.top + (chartHeight / gridLines) * i;
            ctx.beginPath();
            ctx.moveTo(padding.left, y);
            ctx.lineTo(width - padding.right, y);
            ctx.stroke();

            // Y-axis labels
            const value = max - (range / gridLines) * i;
            ctx.fillStyle = '#64748b';
            ctx.font = '11px system-ui';
            ctx.textAlign = 'right';
            ctx.fillText(this.formatNumber(value), padding.left - 10, y + 4);
        }

        // Draw data line
        const color = options.color || this.options.color;
        ctx.strokeStyle = color;
        ctx.lineWidth = 2;
        ctx.beginPath();

        data.forEach((point, index) => {
            const x = padding.left + (chartWidth / (data.length - 1)) * index;
            const y = padding.top + chartHeight - ((point.value - min) / range) * chartHeight;

            if (index === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });
        ctx.stroke();

        // Draw fill area
        if (this.options.fillColor || options.fillColor) {
            ctx.fillStyle = options.fillColor || this.options.fillColor;
            ctx.beginPath();

            data.forEach((point, index) => {
                const x = padding.left + (chartWidth / (data.length - 1)) * index;
                const y = padding.top + chartHeight - ((point.value - min) / range) * chartHeight;

                if (index === 0) {
                    ctx.moveTo(x, y);
                } else {
                    ctx.lineTo(x, y);
                }
            });

            const lastX = padding.left + chartWidth;
            const lastY = padding.top + chartHeight;
            ctx.lineTo(lastX, lastY);
            ctx.lineTo(padding.left, lastY);
            ctx.closePath();
            ctx.fill();
        }

        // Draw data points
        if (this.options.showPoints) {
            ctx.fillStyle = color;
            data.forEach((point, index) => {
                const x = padding.left + (chartWidth / (data.length - 1)) * index;
                const y = padding.top + chartHeight - ((point.value - min) / range) * chartHeight;

                ctx.beginPath();
                ctx.arc(x, y, 4, 0, 2 * Math.PI);
                ctx.fill();

                // White center
                ctx.fillStyle = '#ffffff';
                ctx.beginPath();
                ctx.arc(x, y, 2, 0, 2 * Math.PI);
                ctx.fill();
                ctx.fillStyle = color;
            });
        }

        // Draw X-axis labels
        ctx.fillStyle = '#64748b';
        ctx.font = '11px system-ui';
        ctx.textAlign = 'center';

        const labelInterval = Math.ceil(data.length / 6);
        data.forEach((point, index) => {
            if (index % labelInterval === 0 || index === data.length - 1) {
                const x = padding.left + (chartWidth / (data.length - 1)) * index;
                ctx.fillText(point.label || point.time, x, height - padding.bottom + 20);
            }
        });
    }

    // Draw a bar chart
    drawBarChart(data, options = {}) {
        if (!data || data.length === 0) {
            this.drawEmptyState('No data available');
            return;
        }

        const ctx = this.ctx;
        const width = this.canvas.width;
        const height = this.canvas.height;
        const padding = { top: 20, right: 20, bottom: 40, left: 50 };

        const chartWidth = width - padding.left - padding.right;
        const chartHeight = height - padding.top - padding.bottom;

        // Clear canvas
        ctx.clearRect(0, 0, width, height);

        // Find max value
        const max = Math.max(...data.map(d => d.value)) * 1.1;
        const barWidth = (chartWidth / data.length) * 0.7;
        const barGap = (chartWidth / data.length) * 0.3;

        // Draw grid lines
        ctx.strokeStyle = '#e2e8f0';
        ctx.lineWidth = 1;

        for (let i = 0; i <= 5; i++) {
            const y = padding.top + (chartHeight / 5) * i;
            ctx.beginPath();
            ctx.moveTo(padding.left, y);
            ctx.lineTo(width - padding.right, y);
            ctx.stroke();

            // Y-axis labels
            const value = max - (max / 5) * i;
            ctx.fillStyle = '#64748b';
            ctx.font = '11px system-ui';
            ctx.textAlign = 'right';
            ctx.fillText(this.formatNumber(value), padding.left - 10, y + 4);
        }

        // Draw bars
        const color = options.color || this.options.color;
        data.forEach((item, index) => {
            const x = padding.left + (barWidth + barGap) * index + barGap / 2;
            const barHeight = (item.value / max) * chartHeight;
            const y = padding.top + chartHeight - barHeight;

            // Bar gradient
            const gradient = ctx.createLinearGradient(x, y, x, y + barHeight);
            gradient.addColorStop(0, color);
            gradient.addColorStop(1, this.adjustColor(color, -20));

            ctx.fillStyle = gradient;
            ctx.fillRect(x, y, barWidth, barHeight);

            // X-axis labels
            ctx.fillStyle = '#64748b';
            ctx.font = '11px system-ui';
            ctx.textAlign = 'center';
            ctx.fillText(item.label || item.name, x + barWidth / 2, height - padding.bottom + 20);
        });
    }

    // Draw a pie/donut chart
    drawPieChart(data, options = {}) {
        if (!data || data.length === 0) {
            this.drawEmptyState('No data available');
            return;
        }

        const ctx = this.ctx;
        const width = this.canvas.width;
        const height = this.canvas.height;

        // Clear canvas
        ctx.clearRect(0, 0, width, height);

        const centerX = width / 2;
        const centerY = height / 2;
        const radius = Math.min(width, height) / 2 - 20;
        const innerRadius = options.innerRadius || radius * 0.6;

        const total = data.reduce((sum, item) => sum + item.value, 0);
        let startAngle = -Math.PI / 2;

        const colors = [
            '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
            '#ec4899', '#06b6d4', '#84cc16', '#f97316', '#6366f1'
        ];

        // Draw segments
        data.forEach((item, index) => {
            const sliceAngle = (item.value / total) * 2 * Math.PI;
            const endAngle = startAngle + sliceAngle;

            ctx.beginPath();
            ctx.moveTo(centerX, centerY);
            ctx.arc(centerX, centerY, radius, startAngle, endAngle);
            ctx.closePath();

            ctx.fillStyle = colors[index % colors.length];
            ctx.fill();

            startAngle = endAngle;
        });

        // Draw inner circle (donut hole)
        ctx.beginPath();
        ctx.arc(centerX, centerY, innerRadius, 0, 2 * Math.PI);
        ctx.fillStyle = '#ffffff';
        ctx.fill();

        // Draw legend
        this.drawLegend(data, colors, width);
    }

    drawLegend(data, colors, width) {
        const ctx = this.ctx;
        const legendX = 10;
        const legendY = 10;
        const itemHeight = 20;

        data.forEach((item, index) => {
            const y = legendY + index * itemHeight;

            // Color box
            ctx.fillStyle = colors[index % colors.length];
            ctx.fillRect(legendX, y, 12, 12);

            // Label
            ctx.fillStyle = '#374151';
            ctx.font = '12px system-ui';
            ctx.textAlign = 'left';
            ctx.fillText(`${item.label || item.name}: ${this.formatNumber(item.value)}`, legendX + 20, y + 10);
        });
    }

    drawEmptyState(message) {
        const ctx = this.ctx;
        ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        ctx.fillStyle = '#94a3b8';
        ctx.font = '14px system-ui';
        ctx.textAlign = 'center';
        ctx.fillText(message, this.canvas.width / 2, this.canvas.height / 2);
    }

    formatNumber(num) {
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        }
        if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K';
        }
        return num.toFixed(0);
    }

    adjustColor(color, amount) {
        // Simple hex color adjustment
        const hex = color.replace('#', '');
        const r = Math.max(0, Math.min(255, parseInt(hex.substr(0, 2), 16) + amount));
        const g = Math.max(0, Math.min(255, parseInt(hex.substr(2, 2), 16) + amount));
        const b = Math.max(0, Math.min(255, parseInt(hex.substr(4, 2), 16) + amount));
        return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`;
    }

    destroy() {
        if (this.canvas && this.canvas.parentNode) {
            this.canvas.parentNode.removeChild(this.canvas);
        }
    }
}

// Utility function to create chart from data
function createChart(containerId, type, data, options = {}) {
    const chart = new MetricsChart(containerId, { type, ...options });

    switch (type) {
        case 'bar':
            chart.drawBarChart(data, options);
            break;
        case 'pie':
            chart.drawPieChart(data, options);
            break;
        case 'line':
        default:
            chart.drawLineChart(data, options);
    }

    return chart;
}
