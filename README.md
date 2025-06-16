# Wispy Core CMS - Project Guide

## Overview
Wispy Core is a modern, high-performance, multi-site content management system built with Go.

## Code Style and Standards

### API Design Principles
- **Versioning**: APIs use URL path versioning (e.g., `/api/v1/`)
- **HTTP Methods**: RESTful design with meaningful HTTP methods (GET, POST, PUT, DELETE)
- **Error Handling**: Standardized error responses as text/plain with debugging information in `X-Debug` header
- **Naming Conventions**: 
  - Plural nouns for resource collections (e.g., `/users`, `/products`)
  - Singular nouns with identifiers for individual resources (e.g., `/users?uuid=123`, `/products?uuid=456`)
- **HTTP Status Codes**: Proper status codes to indicate success or failure
- **Query Parameters**: Consistent filtering, sorting, and pagination (e.g., `?page=1&limit=10&sort=name`)
- **Input Validation**: Abstracted validation and sanitization for all API requests

### Go Code Standards
- **Standard Library First**: Use Go's standard library for HTTP handling, templating, and file operations
- **Error Handling**: Follow Go's error handling idioms (errors as last return value)
- **Security**: bcrypt for password hashing, session-based authentication with secure cookies
- **No JWT**: Use session-based authentication instead of JWT tokens
- **Documentation**: Comments for complex logic and important decisions
- **Error Responses**: Plain text error messages, no JSON for error responses
- **Third-party Libraries**: Minimal external dependencies, only when absolutely necessary

### File Organization Standards
- **Models**: External data structures in `/models/<name>-structs.go`
- **SQL Queries**: Complex/reused queries in `/models/<name>-sql.go`
- **Module Structure**: Organized by functionality in dedicated packages

## Architecture & Directory Structure

### Core Application
```
/main.go                    # Main application entry point ✅
├── /auth/                  # Authentication and authorization module ✅
├── /cache/                 # Caching module for performance optimization ✅
├── /common/                # Common utilities and helpers ✅
├── /core/                  # Core CMS functionality ✅
├── /models/                # Data structures and models ✅
├── /tail/                  # Tailwind CSS v4 integration ✅
├── /template-sections/     # Global Templates ⏳
└── /sites/                 # Multi-site directory structure ✅
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
- ✅ Conditional rendering: `{% if condition %}...{% endif %}`
- ✅ Loops: `{% for item in collection %}...{% endfor %}`
- ✅ Template inheritance: `{% define "block" %}...{% enddefine %}`
- ✅ Component inclusion: `{% render "template" %}`
- ✅ Block rendering: `{% block "name" %}`
- ✅ Verbatim content: `{% verbatim %}...{% endverbatim %}`
- ✅ Filter chains: `{{ value | filter1(boop=123,array=["123","456","789"]) | filter2 }}`

**Example**:
```html
{% if user.authenticated %}
    <h1>Welcome, {{ user.name | upcase }}!</h1>
    {% for post in user.posts %}
        <article>{{ post.title }}</article>
    {% endfor %}
{% endif %}
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
**Status**: Working POC
**Description**: Built-in Support with dynamic class extraction and CSS generation.

**Features**:
- ✅ Runtime CSS generation
- ✅ Cascade layers support (theme, base, components, utilities)
- ✅ Responsive breakpoints (sm, md, lg, xl, 2xl, 3xl)
- ✅ Color system with OKLCH support
- ✅ Arbitrary value support: `h-[100px]`, `text-[#123456]`
- ✅ CSS custom properties: `border-(--pattern-fg)`
- ✅ Modern opacity handling with `color-mix()`

**Example Classes**:
```html
<div class="grid-cols-[1fr_2.5rem_auto] h-[1lh] text-sky-400/25">
    <span class="decoration-sky-400 hover:decoration-2">Text</span>
</div>
```

### 5. Authentication & Authorization 
**Status**: Implemented
**Description**: Secure user authentication with session management and role-based access control.

**Features**:
- ✅ User registration and login
- ✅ Session-based authentication (no JWT)
- ⏳ Role-based access control
- ⏳ Per-page authentication requirements
- ⏳ OAuth provider support structure
- ⏳Security features (rate limiting, failed attempt tracking)

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
**Status**: Structure Ready
- ⏳ `GET /api/v1/health` - Health check endpoint
- ⏳ `GET /api/v1/sites` - List all sites
- Additional CRUD endpoints for content management

### Authentication APIs
**Status**: Backend Ready, API Layer Pending
-  `POST /api/v1/auth/login` - User login
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
```

### Site Configuration (config.toml)
```toml
## TBD
```

## Testing & Quality Assurance

### Test Coverage 
- ⏳ **Template Engine**: 12+ test suites covering all template features
- ⏳ **Asset System**: Complete asset import/export testing
- 🔴 **Authentication**: User registration, login, session management
- ⏳ **Tailwind CSS**: v4 features and class generation
- 🔴 **Error Handling**: Graceful  failure scenarios
- 🔴 **Integration**: Real HTML file processing

### Test Commands
```bash
go test ./...                    # Run all tests
go test ./core -v               # Verbose core tests
go test ./auth -v               # Authentication tests
```

## Production Readiness

### Security Features
- HTTPS enforcement for remote assets
- Directory traversal prevention
- Input validation and sanitization
- Session-based authentication
- Rate limiting
- Secure password hashing

### Performance Features
- Efficient template caching
- Database connection pooling
- Static asset optimization
- Minimal memory allocation
- Fast route matching

### Monitoring & Logging 
- Structured logging with levels
- Request timing and statistics
- Error tracking and reporting
- Development vs production logging

## Roadmap & Future Features
- **API Layer**: RESTful API for content management
- **Admin Interface**: Web-based administration panel
- **Database Migrations**: Automated schema management
- **Content Types**: Custom content type definitions
- **Plugin System**: Extensible plugin architecture
- **Advanced Caching**: Redis integration, CDN support
- **Media Management**: File upload and management system
- **SEO Tools**: Advanced SEO optimization features
- **Multi-language Support**: Internationalization
- **E-commerce Integration**: Shopping cart and payment processing
- **Advanced Analytics**: Built-in analytics and reporting
- **Cloud Integration**: AWS, GCP, Azure deployment tools

---

## Legend
- ✅ **Fully Implemented**: Feature is complete and tested
- ⏳ **In Progress/Planned**: Feature is partially implemented or planned
- 🔴 **Not Started**: Feature is in roadmap but not yet started

---

*Last Updated: June 15, 2025*