![Alt text](/web/static/images/godash1.png)
# Godash

A modern admin dashboard web application built with Go, HTML, CSS, and JavaScript. Features a modular architecture with widget-based dashboard components, real-time updates, and responsive design.

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

## Configuration

The application can be configured using environment variables:

- `PORT` - Server port (default: 8080)
- `HOST` - Server host (default: localhost)
- `SESSION_SECRET` - Session secret key (default: auto-generated)
- `SESSION_MAX_AGE` - Session duration in seconds (default: 86400)

## Project Structure

```
/
├── cmd/server/          # Application entry point
├── internal/            # Private application packages
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP request handlers
│   ├── middleware/     # Authentication middleware
│   ├── models/         # Data models
│   └── services/       # Business logic
└── web/                # Frontend assets
    ├── templates/      # HTML templates
    └── static/         # CSS, JavaScript, images
```

## API Endpoints

- `GET /api/dashboard` - Complete dashboard data
- `GET /api/stats` - System statistics
- `GET /api/users` - User list (admin only)

## Development

See [WARP.md](WARP.md) for detailed development commands and architecture documentation.
