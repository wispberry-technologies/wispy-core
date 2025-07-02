# Wispy Core CMS - Project Guide

## Overview
Wispy Core is a modern, high-performance, multi-site content management system built with Go.

## Code Style and Standards

### API Design Principles
- **Routers**: Uses `chi` for routing
- **RESTful**: Adheres to REST principles for resource-oriented APIs
- **Versioning**: APIs use URL path versioning (e.g., `/api/v1/`)
- **HTTP Methods**: RESTful d### API Endpoints
- **Error Handling**: Standardized error responses as text/plain with debugging information in `X-Debug` header
- **Naming Conventions**: 
  - Plural nouns for resource collections (e.g., `/users`, `/products`)
  - Singular nouns with identifiers for individual resources (e.g., `/users?uuid=123`, `/products?uuid=456`)
- **HTTP Status Codes**: Proper status codes to indicate success or failure
- **Query Parameters**: Consistent filtering, sorting, and pagination (e.g., `?page=1&limit=10&sort=name`)
- **Input Validation**: Abstracted validation and sanitization for all API requests

### Go Code Standards
- **Standard Library First**: Use Go's standard library for HTTP handling, templating, and file operations
- **Functional Programming**: Default to functional programming approaches unless OOP offers clear benefits
- **Minimal OOP**: Use OOP and dependency injection only when it improves readability and maintainability 
- **Error Handling**: Follow Go's error handling idioms (errors as last return value)
- **Security**: bcrypt for password hashing, session-based authentication with secure cookies
- **No JWT**: Use session-based authentication instead of JWT tokens
- **Documentation**: Comments for complex logic and important decisions
- **Third-party Libraries**: Minimal external dependencies, only when absolutely necessary

### File Organization Standards
- **Models**: External data structures in `pkg/models/<name>_structs.go`
- **SQL Queries**: Complex/reused queries in `internal/database/<name>.go`
- **Module Structure**: Organized by functionality in dedicated packages
- **Public vs Private**: Clear separation between internal (private) and pkg (public) code

## Architecture & Directory Structure

### Application Structure
```
wispy-core/
├── cmd/               # Command-line tools
│   ├── server/        # Main server application ⏳
│   ├── migrate/       # Database migration tool ⏳
│   └── dev/           # Development utilities ⏳
├── internal/          # Private application code
│   ├── api/           # API handlers and routes ⏳
│   │   └── v1/        # API version 1 ⏳
│   ├── auth/          # Authentication and authorization ⏳
│   ├── cache/         # Caching functionality ⏳
│   ├── core/          # Core CMS functionality ⏳
│   ├── database/      # Database connections and utilities ⏳
│   ├── server/        # HTTP server and middleware ✅
│   └── sites/         # Site management ⏳
├── pkg/               # Public packages for external consumption
│   ├── common/        # Shared utilities ✅
│   ├── models/        # Data structures and models ✅
│   └── wispytail/     # Tailwind CSS v4 integration ✅
├── data/              # Application data
│   └── sites/         # Multi-site content ✅
│   └── templates/     # Global templates and components ✅
├── scripts/           # Build and maintenance scripts ✅
├── tests/             # Testing utilities ⏳
├── go.mod             # Go module definition ✅
└── README.md          # Project documentation ✅
```

### Per-Site Structure
```
/sites/<domain>/
├── config/
│   ├── config.toml         # Site configuration ✅
│   └── themes/             # Theme CSS variables based on DaisyUI ✅
│       ├── pale-wisp.css
│       └── midnight-wisp.css
├── dbs/                    # Site databases ✅
│   ├── users.db
│   └── databases.toml      # Database registry ⏳
├── assets/                 # Private assets (compiled) ✅
├── public/                 # Public static files ✅
├── layouts/                # Page layout templates ✅
├── pages/                  # Content pages ✅
│   ├── 404.html
│   └── home.html
├── templates/              # Reusable template components ✅
│   ├── partials/           # Small reusable templates for use in pages or to be rendered as api response  ✅
│   ├── components/         # Small reusable components ✅
│   └── sections/           # Larger content sections ✅
└── migrations/             # Database migration files ⏳
```

## Core Features

### 1. Multi-Site Management ✅
**Status**: Fully Implemented
**Description**: Manage multiple independent websites from a single Wispy Core installation.

**Features**:
- ✅ Domain-based site routing
- ✅ Per-site configuration and assets
- ✅ Isolated databases per site
- ✅ Independent themes and layouts

**Example**:
```bash
/sites/example.com/     # Site 1
/sites/blog.example.com/ # Site 2
/sites/abc.com/ # Site 3
```

### 2. Advanced Template Engine ✅
**Status**: Fully Implemented
**Description**: Custom template engine with support for variables, loops, conditionals, and components.

**Features**:
- ✅ Variable interpolation: `{{ variable }}`
- ✅ Conditional rendering: `{% if condition %}...{% end %}`
- ✅ Loops: `{% for item in collection %}...{% endfor %}`
- ✅ Template inheritance: `{% define "block" %}...{% enddefine %}`
- ✅ Component inclusion: `{% render "template" %}`
- ✅ Block rendering: `{% block "name" %}`
- ✅ Verbatim content: `{% verbatim %}...{% endverbatim %}`
- ✅ Filter chains: `{{ value | filter1(boop=123,array=["123","456","789"]) | filter2 }}`
- ✅ Custom filters: `{{ value | upcase }}`
- ✅ Custom tags: `{% asset "css" "path/to/style.css" %}`
- ⏳ Custom components with props: `{% component "name" (prop1=value1, prop2="value2") %}`


**Example**:
```html
{% if user.authenticated %}
    <h1>Welcome, {{ user.name | upcase }}!</h1>
    {% for post in user.posts %}
        <article>{{ post.title }}</article>
    {% endfor %}
{% end %}
```

### 3. Asset Management System ✅
**Status**: Fully Implemented
**Description**: Secure asset management with support for CSS, JavaScript, and external resources.

**Features**:
- ✅ External CSS/JS: `{% asset "css" "public/css/style.css" %}`
- ✅ Inline CSS/JS: `{% asset "css-inline" "assets/css/critical.css" %}`
- ✅ Remote assets: `{% asset "css" "https://cdn.example.com/style.css" %}`
- ✅ JavaScript location control: `{% asset "js" "app.js" location="head" %}`
- ✅ Deduplication and conflict detection
- ✅ Security restrictions (assets/ and public/ only, HTTPS remote only)
- ✅ Graceful error handling

**Example**:
```html
<!-- Critical inline CSS -->
{% asset "css-inline" "assets/css/critical.css" %}

<!-- External CSS -->
{% asset "css" "public/css/main.css" %}

<!-- JavaScript in footer -->
{% asset "js" "public/js/app.js" location="pre-footer" %}
```

### 4. WispyTail (Tailwind CSS v4 Inspired) Real-time Atomic CSS Compiler 
**Status**: Fully Implemented
**Description**: Built-in Support with dynamic class extraction and CSS generation.

**Features**:
- ✅ Runtime CSS generation
- ✅ Cascade layers support (theme, base, components, utilities)
- ✅ Responsive breakpoints (sm, md, lg, xl, 2xl, 3xl)
- ✅ Color system with OKLCH support
- ✅ Arbitrary value support: `h-[100px]`, `text-[#123456]`
- ✅ CSS custom properties: `border-(--pattern-fg)`
- ✅ Modern opacity handling with `color-mix()`
- ✅ DaisyUI component class support

**Example Classes**:
```html
<div class="grid-cols-[1fr_2.5rem_auto] h-[1lh] text-sky-400/25">
    <span class="decoration-sky-400 hover:decoration-2">Text</span>
</div>
```

### 5. Authentication & Authorization 
**Status**: Fully Implemented
**Description**: Secure user authentication with session management and role-based access control.

**Features**:
- ✅ User registration and login
- ✅ Session-based authentication (no JWT)
- ✅ SQL-based session and user storage
- ⏳ Role-based access control
- ✅ Per-page authentication requirements
- ⏳ OAuth provider support structure
- ⏳ Security features (rate limiting, failed attempt tracking)

**Example Page Metadata**:
```html
<!--
@require_auth true
@required_roles ["admin", "editor"]
-->
```

### 6. Page System
**Status**: Fully Implemented
**Description**: Flexible page system with metadata, layouts, and content management.

**Features**:
- ✅ Metadata-driven pages with frontmatter
- ✅ Layout inheritance
- ⏳Draft mode support
- ✅ Custom data fields
- ⏳SEO metadata (title, description, keywords)
- ⏳ Author attribution
- ⏳ Publication dates

**Example Page**:
```html
<!--
@name home.html
@slug /
@author Wispy Core Team
@layout default
@is_draft false
@require_auth false
@required_roles []
@title Welcome to Our Site
@description A modern CMS built with Go
-->
<div class="hero">
    <h1>{{ Page.Title }}</h1>
</div>
```

### 7. Static File Serving 
**Status**: Fully Implemented
**Description**: Secure static file serving for public assets and resources.

**Features**:
- ✅ Public directory serving (`/public/*`)
- ✅ Assets directory serving (`/assets/*`)
- ✅ Security checks (directory traversal prevention)
- ✅ Per-site asset isolation

**URL Structure**:
```
/public/css/style.css   → sites/domain/public/css/style.css
/assets/js/app.js       → sites/domain/assets/js/app.js
```

### 8. Error Handling & 404 Pages 
**Status**: Fully Implemented
**Description**: Graceful error handling with custom 404 pages and comprehensive logging.

**Features**:
- ✅ Custom 404 page rendering
- ✅ Template error resilience
- ⏳ Structured error logging
- ✅ Non-blocking asset errors
- ⏳ Fallback error pages

**Example 404 Page**:
```html
<!--
@name 404.html
@slug /404
@layout default
-->
<div class="hero">
    <h1>404 - Page Not Found</h1>
    <a href="/" class="btn btn-primary">Go Home</a>
</div>
```

### 9. Performance & Caching
**Status**: Implemented
**Description**: Multiple caching layers for optimal performance.

**Features**:
- ⏳ Static Page caching
- 🔴 Dynamic Page partial caching
- 🔴 Template caching
- 🔴 Database connection pooling
- ⏳ In-memory caching with expiration
- 🔴 Route statistics and monitoring
- 🔴 Per-site database caching

### 10. Development Tools
**Status**: Implemented
**Description**: Developer-friendly tools and testing infrastructure.

**Features**:
- ⏳ Comprehensive test suite
- ⏳ Environment-based configuration
- ⏳ Detailed logging with colors
- ⏳ Development vs production modes/debugging
- ⏳ Database migrations (structure ready)

## API Endpoints

### Core APIs
- ⏳ `GET /api/v1/health` - Health check endpoint
- ⏳ `GET /api/v1/sites` - List all sites
### Pages APIs
- ⏳ `GET /api/v1/pages` - List all pages
- ⏳ `GET /api/v1/pages?page_id=<id>` - Get page details
- ⏳ `POST /api/v1/pages` - Create new page
- ⏳ `PUT /api/v1/pages/?page_id=<id>` - Update page
### Content APIs
<!-- General Content Api (Used to load into sites) -->
- ⏳ `GET /api/v1/content` - List all content
- ⏳ `GET /api/v1/content/?content_id=<id>` - Get content details
- ⏳ `POST /api/v1/content` - Create new content
- ⏳ `PUT /api/v1/content/?content_id=<id>` - Update content
- ⏳ `DELETE /api/v1/content/?content_id=<id>` - Delete content
### E-Commerce APIs
- ⏳ `GET /api/v1/shop/products` - List all products
- ⏳ `GET /api/v1/shop/products/:id` - Get product details
- ⏳ `POST /api/v1/shop/cart/items` - Add item to cart
- ⏳ `GET /api/v1/shop/cart` - Get cart details
- ⏳ `POST /api/v1/shop/checkout` - Checkout

### Authentication APIs
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/auth/me` - Current user info

## Configuration

### Environment Variables
```bash
# Server Configuration
PORT=8080                           # Server port
HOST=localhost                      # Server host
ENV=development                     # Environment mode
SITES_PATH=/path/to/sites          # Sites directory

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_SECOND=12  # Request rate limit
RATE_LIMIT_REQUESTS_PER_MINUTE=240 # Request rate limit

# 
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=
DISCORD_REDIRECT_URI=

```

### Site Configuration (config.toml)
```toml
## TBD
```

## Testing & Quality Assurance

### Test Coverage 
- ✅ **Template Engine**: 12+ test suites covering all template features
- ⏳ **Asset System**: Complete asset import/export testing
- ⏳ **Authentication**: User registration, login, session management
- ✅ **WispyTail CSS**: Feature testing and class generation
- ⏳ **Error Handling**: Graceful failure scenarios
- ⏳ **Integration**: Real HTML file processing

### Test Commands
```bash
go test ./...                    # Run all tests
go test ./internal/core -v       # Verbose core tests
go test .auth -v       # Authentication tests
./scripts/run-tests.sh           # Run all tests with coverage reports
```

## Running the Application

### Using the CLI Tool

Wispy Core includes a unified command-line tool for all operations:

```bash
# Build the CLI tool
./scripts/build-cli.sh

# Start the server
./bin/wispy-cli server

# Fetch external assets
./bin/wispy-cli fetch

# Run security tests
./bin/wispy-cli fetch -target=http://localhost:8080
```

See detailed CLI documentation in `cmd/wispy-cli/README.md`.

### Legacy Commands

```bash
# Run the server directly
go run ./cmd/server

# Run database migrations
go run ./cmd/migrate -op up

# Development mode
./scripts/run.sh
```

## Building for Production

```bash
# Build the server binary
go build -o wispy-cms ./cmd/server

# Build the migration tool
go build -o wispy-migrate ./cmd/migrate
```

## Production Readiness

### Security Features
- ✅ HTTPS enforcement for remote assets
- ✅ Directory traversal prevention
- ✅ Input validation and sanitization
- ✅ Session-based authentication
- ✅ Rate limiting
- ✅ Secure password hashing

### Performance Features
- ⏳ Efficient template caching
- ⏳ Database connection pooling
- ⏳ Static asset optimization
- ⏳ Minimal memory allocation
- ⏳ Fast route matching

### Monitoring & Logging 
- ⏳ Structured logging with levels
- ⏳ Request timing and statistics
- ⏳ Error tracking and reporting
- ⏳ Development vs production logging

## Roadmap & Future Features
- ⏳ **API Layer**: RESTful API for content management (In Progress)
- 🔴 **Admin Interface**: Web-based administration panel
- ⏳ **Database Migrations**: Automated schema management (Structure Implemented)
- 🔴 **Content Types**: Custom content type definitions
- 🔴 **Plugin System**: Extensible plugin architecture
- 🔴 **Advanced Caching**: Redis integration, CDN support
- 🔴 **Media Management**: File upload and management system
- 🔴 **SEO Tools**: Advanced SEO optimization features
- 🔴 **Multi-language Support**: Internationalization
- 🔴 **E-commerce Integration**: Shopping cart and payment processing
- 🔴 **Advanced Analytics**: Built-in analytics and reporting
- 🔴 **Cloud Integration**: AWS, GCP, Azure deployment tools
- ⏳ **Projects API**: Project management API for collaborative development (In Progress)

---

## Legend
- ✅ **Fully Implemented**: Feature is complete and tested
- ⏳ **In Progress/Planned**: Feature is partially implemented or planned
- 🔴 **Not Started**: Feature is in roadmap but not yet started

---

*Last Updated: June 16, 2025*