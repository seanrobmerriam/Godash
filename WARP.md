# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Development Commands

### Build and Run
```bash
# Build the application
go build -o godash ./cmd/server

# Run directly with Go
go run ./cmd/server

# Initialize/update dependencies
go mod tidy

# Build for production
go build -ldflags "-s -w" -o godash ./cmd/server
```

### Development Server
```bash
# Start the development server (runs on localhost:8080)
go run ./cmd/server

# Run with custom port
PORT=3000 go run ./cmd/server

# Run with environment variables
SESSION_SECRET="your-secret-key" go run ./cmd/server

# Access the dashboard at http://localhost:8080/dashboard
# Default credentials: admin / password
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/services/...

# Run specific test
go test -v -run TestFunctionName ./internal/services

# Run tests with coverage
go test -cover ./...
```

### Dependencies
```bash
# Add a new dependency
go get github.com/example/package

# Remove unused dependencies
go mod tidy

# View dependency graph
go mod graph

# Update all dependencies
go get -u ./...
```

## Architecture Overview

Godash is a modern Go web application providing an admin dashboard with a clean separation of concerns and modular architecture:

### Project Structure

The application follows Go best practices with a well-organized directory structure:

```
/
├── cmd/server/          # Application entry point
│   └── main.go         # Server initialization and routing
├── internal/           # Private application packages
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP request handlers
│   ├── middleware/     # HTTP middleware (auth, logging)
│   ├── models/         # Data models and structures
│   └── services/       # Business logic layer
├── web/               # Frontend assets
│   ├── templates/     # HTML templates
│   └── static/        # CSS, JavaScript, images
├── pkg/               # Public packages (if any)
└── go.mod             # Go module definition
```

### Core Components

**Configuration Layer** (`internal/config`): Environment-based configuration with sensible defaults. Supports database, server, and session configuration.

**Service Layer** (`internal/services`): Contains business logic for user management, dashboard data, and system statistics. Services are thread-safe and use in-memory storage by default.

**Middleware Layer** (`internal/middleware`): Authentication middleware with session management, role-based access control, and API request detection.

**Handler Layer** (`internal/handlers`): HTTP request handlers for both web pages and API endpoints. Uses Go templates for server-side rendering.

**Data Layer** (`internal/models`): Structured data models for users, dashboard widgets, system stats, and various widget types (charts, tables, metrics, etc.).

### Authentication & Authorization

- **Session-based Authentication**: Uses Gorilla Sessions with secure cookie storage
- **Role-based Access**: Supports admin and user roles with middleware enforcement
- **API Support**: Detects API requests and returns appropriate JSON responses
- **Default Credentials**: admin/password (configurable via environment)

### Dashboard Architecture

**Widget System**: Modular widget architecture supporting multiple widget types:
- **Charts**: Line charts with custom drawing (canvas-based)
- **Metrics**: Key performance indicators with trend indicators
- **Tables**: Tabular data display with sorting capabilities
- **Activity**: Timeline-based activity feeds
- **Progress**: Progress bars with labels and descriptions
- **Text**: Key-value information displays

**Real-time Updates**: JavaScript-based auto-refresh every 30 seconds with manual refresh options per widget.

**Responsive Design**: Mobile-first CSS with grid-based layout that adapts to different screen sizes.

### API Architecture

**RESTful Endpoints**:
- `GET /api/dashboard` - Complete dashboard data
- `GET /api/stats` - System statistics only
- `GET /api/users` - User management (admin only)

**JSON Responses**: All API endpoints return structured JSON with consistent error handling.

### Frontend Architecture

**Template Engine**: Go's `html/template` with data binding and conditional rendering.

**CSS Framework**: Custom CSS with utility classes, responsive grid system, and component-based styling.

**JavaScript**: Vanilla JavaScript with ES6+ features, modular dashboard class, and async/await for API calls.

### Request Flow

1. **Router**: Gorilla Mux handles URL routing and method matching
2. **Middleware**: Authentication and authorization checks
3. **Handlers**: Business logic processing and response generation
4. **Services**: Data retrieval and manipulation
5. **Templates/JSON**: Response rendering based on request type

### Data Flow

1. **Services** generate dashboard data with system statistics
2. **Handlers** process requests and call appropriate services
3. **Templates** render server-side HTML with initial data
4. **JavaScript** handles client-side updates and interactions
5. **API endpoints** provide JSON data for dynamic updates

## Development Notes

### Key Features
- **Modular Architecture**: Clean separation of concerns with dedicated packages
- **Configuration Management**: Environment-based configuration with defaults
- **Authentication System**: Session-based auth with role-based access control
- **Widget-based Dashboard**: Extensible widget system for various data types
- **Real-time Updates**: Auto-refreshing dashboard with manual refresh options
- **Responsive Design**: Mobile-friendly interface with grid-based layout
- **API-first Design**: RESTful APIs for all data interactions

### Current Implementation
- **In-memory Storage**: All data stored in memory (suitable for development/demo)
- **Simulated Data**: Dashboard shows simulated system statistics and activity
- **Thread-safe Services**: Concurrent access handled with proper locking
- **Template-based Rendering**: Server-side rendering with Go templates
- **Vanilla JavaScript**: No external JS frameworks, pure ES6+ implementation

### Extension Points
- **Widget Types**: Add new widget types in `models/dashboard.go` and corresponding renderers
- **Data Sources**: Replace simulated data with real system metrics or external APIs
- **Authentication**: Add database-backed user storage and OAuth integration
- **Database**: Implement persistent storage with your preferred database
- **Monitoring**: Add logging, metrics, and health check endpoints
- **Deployment**: Add Docker configuration and deployment scripts
