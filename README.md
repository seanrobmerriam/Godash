![Alt text](/web/static/images/godash1.png)
# Godash

A modern caddy admin dashboard built with Go, HTML, CSS, and JavaScript. Features a modular architecture with widget-based dashboard components, real-time updates, and responsive design.

## Features

- **Modular Architecture**: Clean separation with dedicated packages for handlers, middleware, models, and services
- **Widget-based Dashboard**: Extensible system supporting charts, metrics, tables, activity feeds, and progress bars
- **Real-time Updates**: Auto-refreshing dashboard with manual refresh options
- **Session-based Authentication**: Secure login system with role-based access control
- **Responsive Design**: Mobile-friendly interface with grid-based layout
- **RESTful API**: JSON endpoints for all data interactions


## Quick Start

1. **Clone the repository**:
   ```bash
   git clone https://github.com/seanrobmerriam/Godash.git
   cd Godash
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run the server**:
   ```bash
   go run ./cmd/server
   ```

4. **Access the dashboard**:
   - Open http://localhost:8080/dashboard
   - Default credentials: `admin` / `password`

## Caddy Integration

Godash can manage Caddy webserver instances through the admin API.

### Adding a Caddy Instance

1. Navigate to **Caddy Instances** at `/caddy/instances`
2. Click **Add Instance**
3. Enter the instance details:
   - **Name**: A descriptive name for the instance
   - **Admin API URL**: The Caddy admin API URL (e.g., `http://localhost:2019`)
   - **API Key File**: Path to a file containing the API key (optional)
   - **Tags**: Comma-separated tags for grouping (e.g., `production, web`)

### Managing Instances

- **View Instances**: See all configured instances with their status
- **Filter by Tag**: Click on tags to filter instances
- **Refresh Status**: Update the status of an instance
- **Analytics**: View metrics for a specific instance
- **Config Editor**: Edit the configuration directly
- **Delete**: Remove an instance

### Configuration Editor

Access the config editor at `/caddy/instances/{id}/config`:
- View and edit configuration as JSON or Caddyfile
- Save and reload configuration
- Validate configuration before applying
- Export configuration to file

### Analytics Dashboard

Access analytics at `/caddy/analytics`:
- View requests over time
- Monitor bandwidth usage
- See response code distribution
- Track top sites

## Configuration

The application can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | 8080 |
| `HOST` | Server host | localhost |
| `SESSION_SECRET` | Session secret key | auto-generated |
| `SESSION_MAX_AGE` | Session duration in seconds | 86400 |

## Project Structure

```
/
├── cmd/server/          # Application entry point
├── internal/            # Private application packages
│   ├── caddy/          # Caddy integration
│   │   ├── audit.go    # Audit logging
│   │   ├── client.go   # Caddy API client
│   │   ├── config.go   # Configuration operations
│   │   ├── instances.go # Instance management
│   │   ├── models.go   # Data models
│   │   └── analytics.go # Analytics storage
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP request handlers
│   ├── middleware/     # Authentication middleware
│   ├── models/         # Data models
│   └── services/       # Business logic
├── web/                # Frontend assets
│   ├── templates/      # HTML templates
│   └── static/         # CSS, JavaScript, images
└── data/               # File-based storage (created at runtime)
    ├── instances.json  # Instance configurations
    ├── analytics/      # Metrics history
    └── logs/           # Audit logs
```

## API Endpoints

### Dashboard API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/dashboard` | GET | Complete dashboard data |
| `/api/stats` | GET | System statistics |
| `/api/users` | GET | User list (admin only) |

### Caddy Instance Management

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/caddy/instances` | GET | List all instances |
| `/api/caddy/instances` | POST | Add new instance |
| `/api/caddy/instances/{id}` | GET | Get instance details |
| `/api/caddy/instances/{id}` | PUT | Update instance |
| `/api/caddy/instances/{id}` | DELETE | Delete instance |
| `/api/caddy/instances/{id}/test` | POST | Test connection |
| `/api/caddy/instances/{id}/refresh` | POST | Refresh status |
| `/api/caddy/instances/{id}/health` | GET | Health check |

### Caddy Control Operations

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/caddy/instances/{id}/metrics` | GET | Get metrics |
| `/api/caddy/instances/{id}/config` | GET | Get config (JSON) |
| `/api/caddy/instances/{id}/config/caddyfile` | GET | Get config (Caddyfile) |
| `/api/caddy/instances/{id}/reload` | POST | Reload config |
| `/api/caddy/instances/{id}/stop` | POST | Stop server |
| `/api/caddy/instances/{id}/restart` | POST | Restart server |
| `/api/caddy/instances/{id}/logs` | GET | Get logs |

### Caddy Site Management

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/caddy/instances/{id}/sites` | GET | List sites |
| `/api/caddy/instances/{id}/sites` | POST | Create site |
| `/api/caddy/instances/{id}/sites/{site}` | DELETE | Delete site |

## Security

- **API Keys**: Stored in separate files, referenced by path
- **Authentication**: Session-based with role-based access control
- **HTTPS**: Use HTTPS for all connections to Caddy instances
- **Audit Logging**: All control operations are logged

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o godash ./cmd/server
```

## License

MIT License
