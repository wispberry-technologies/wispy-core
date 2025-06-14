# Wispy Core CMS

A modern, fast, and developer-friendly content management system built with Go. Perfect for managing multiple websites from a single installation.

## Features

- **Lightning Fast**: Built with Go for exceptional performance and low resource usage
- **Multi-Site Architecture**: Manage multiple websites from a single installation
- **Modern Design**: Beautiful, responsive themes powered by daisyUI and Tailwind CSS
- **Developer Friendly**: Clean architecture, excellent documentation, and extensible design
- **Secure**: Built-in security features and best practices for modern web applications
- **Responsive**: Mobile-first design that works perfectly on all devices

## Quick Start

### Prerequisites

- Go 1.21 or later
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/your-org/wispy-core.git
cd wispy-core
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run the application:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

### Environment Variables

- `PORT`: Server port (default: 8080)
- `HOST`: Server host (default: localhost)
- `SITES_PATH`: Path to sites directory (required)
- `ENV`: Environment (development/production, default: development)
- `RATE_LIMIT_REQUESTS_PER_SECOND`: Rate limiting (default: 12)
- `RATE_LIMIT_REQUESTS_PER_MINUTE`: Rate limiting (default: 240)

## Project Structure

```
wispy-core/
├── main.go                 # Main application entry point
├── api/                    # API-related code
│   ├── handlers/          # Request handlers
│   └── routes/            # Route definitions
├── modules/               # Reusable modules
│   ├── cache/            # Caching functionality
│   └── common/           # Common utilities
├── sites/                # Individual site configurations
│   └── localhost/        # Example site
│       ├── config/       # Site configuration
│       ├── pages/        # HTML pages with metadata
│       └── public/       # Static assets
└── tests/                # Test functionality
```

## Creating Pages

Pages are HTML files with metadata in comment blocks:

```html
<!--
@name index.html
@slug /
@author Your Name
@layout default
@is_draft false
@require_auth false
@required_roles []
@custom_field Custom Value
-->
<!DOCTYPE html>
<html>
<head>
    <title>My Page</title>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@5" rel="stylesheet" type="text/css" />
    <script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
</head>
<body>
    <h1>Hello World</h1>
</body>
</html>
```

### Metadata Fields

- `@name`: Page filename
- `@slug`: URL path for the page
- `@author`: Page author
- `@layout`: Layout template to use
- `@is_draft`: Whether page is a draft (true/false)
- `@require_auth`: Whether authentication is required (true/false)
- `@required_roles`: Array of required roles for access
- Custom fields: Any other `@field_name` becomes custom metadata

## API Endpoints

### Health Check
- `GET /api/v1/health` - Check server health

### Sites Management
- `GET /api/v1/sites` - List all sites

## Development

### Running Tests
```bash
# Run with test mode enabled
TEST_MODE=true go run main.go env config
```

### Adding a New Site

1. Create a directory in `sites/` with your domain name:
```bash
mkdir sites/example.com
```

2. Create the basic structure:
```bash
mkdir -p sites/example.com/{config,pages,public,assets}
```

3. Add a configuration file:
```toml
# sites/example.com/config/config.toml
title = "My Website"
theme = "default"
description = "My awesome website"

[database]
default = "main.db"

[cache]
enabled = true
ttl = 3600
```

4. Create your first page:
```html
<!-- sites/example.com/pages/index.html -->
<!--
@name index.html
@slug /
@author Your Name
@layout default
@is_draft false
@require_auth false
@required_roles []
-->
<!DOCTYPE html>
<html>
<head>
    <title>Welcome</title>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@5" rel="stylesheet" type="text/css" />
    <script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
</head>
<body>
    <div class="hero min-h-screen bg-base-200">
        <div class="hero-content text-center">
            <h1 class="text-5xl font-bold">Welcome to My Site</h1>
        </div>
    </div>
</body>
</html>
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, please open an issue on GitHub or contact the development team.