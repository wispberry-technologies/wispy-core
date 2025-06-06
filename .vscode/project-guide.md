# Initial Plans

## Code Style and Rules
### API Design
- Modern HTTP API design principles use of RESTful principles
- Versioning of APIs using URL path (e.g., /api/v1/) meaningful HTTP methods (GET, POST, PUT, DELETE)
- Default response format is JSON array for data
- Error handling with standardized error responses as text with debugging sent in the `X-Debug` header
- Use of consistent naming conventions for endpoints (e.g., plural nouns for resources)
- Use of plural nouns for resource names (e.g., /users, /products)
- Use of singular nouns for individual resource endpoints (e.g., /users/?uuid=53sds24fs234123, /products?uuid=53sds24fs234123)
- Use of HTTP status codes to indicate success or failure
- Use of query parameters for filtering, sorting, and pagination (e.g., ?page=1&limit=10&sort=name)
- Abstracted input validation and sanitization for all API requests
- Logging of all API requests and responses for debugging and monitoring purposes

### Code Style
- Use of Go's standard formatting tools (gofmt) for consistent code style
- Use of meaningful variable and function names like `getUserByID`, `createProduct` using CamelCase for function names
- Use of snake_case for variable names (e.g., `user_id`, `product_name`)
- Use of Kebab-case for file names (e.g., `get-user-by-id.go`, `create-product.go`)
- Use of comments to explain complex logic or important decisions
- Use of Go's error handling idioms (returning errors as the last return value)

## Directory Structure
### `/main.go` #main file for go application
### - `/modules/<module_name>` Directory for reusable modules (go packages)
- ├── auth # Authentication and authorization module 
- └── cache # Caching module for performance optimization
- - In-memory caching implementation with expiration
- - cache interface for different caching strategies
- - interfaces for initialization caches for pages, products, users
- - as well as caching open db connections per site
- ├── common # Common utilities and helpers
###  `/api` Directory for API-related code
- ├── routes # API route definitions
- ├── handlers # API request handlers
- ├── middleware # Middleware for request processing
- ├── services # Business logic and service layer
- └── models # Data models and database interaction

## `/sites/<site_domain>/*` Directory for individual site configurations and files
- ├── .env            # Environment variables for the site
- ├── dbs             # Database files for the site
- ├───├── databases.toml     # tracks site databases since users will be able to add multiple databases from the CRM
- ├───├── logs.db            # Log files for site operations (default)
- ├───├── users.db           # User data for the site (default)
- ├───├── products.db        # Product data for the site (default)
- ├───├── pages.db           # Page data for the site (default)
- ├───└── `<name>`.db          # any additional database files for the site
- ├── migrations      # Database migration files
- ├───└── 0001_initial.sql # Initial migration file
- ├── public          # Publicly accessible files \ public assets (CSS, JS, images, fonts, etc.)
- ├── assets          # Stores private static assets (CSS, JS, images, fonts, etc.)
- ├── blocks          # Reusable, nestable, customizable UI components
- ├── config          # Global theme settings and customization options
- ├───├── config.toml     # Main configuration file for the site
- ├───├── themes          # Theme css variable files for the site
- ├───├── pale-wisp.css
- ├───└── midnight-wisp.css
- ├── layout          # Top-level wrappers for pages (layout templates)
- ├── snippets        # Reusable Liquid code or HTML fragments
- ├── pages           # Template stryle pages that can be rendered but also have a head with metadata and function calls
- ├── templates       # Templates combining sections to define page structures
- └── sections        # Modular full-width page components

## How pages are built using sections and templates
- Pages are built using sections and templates, where sections are modular components that can be reused across different pages.
- Templates define using the go html/template package, which allows for dynamic content rendering.
- Pages have a head section that contains metadata and a body section that contains the main content.
- - this metadata includes what should be check at render time such as authorization check, or if the page is a draft
- - the body section contains the main content of the page, which can include sections, templates, and other HTML elements.
- Templates can include other templates, allowing for nested structures and complex layouts.
- Sections can be customized with parameters, allowing for flexible and reusable components.
- Example of a page with a head and body section
```html
<!--
@name home.html
@url /
@author Wispy Core Team
@layout default
@is_draft false
@require_auth false
@required_roles []
-->
<div class="hero min-h-screen bg-base-200">
    <div class="hero-content text-center">
        <div class="max-w-4xl">
            <h1 class="text-5xl font-bold mb-8">{{ .Page.Meta.CustomData.hero_title }}</h1>
            <p class="text-xl mb-8">{{ .Page.Meta.CustomData.hero_description }}</p>
            <a href="{{ .Page.Meta.CustomData.hero_button_link }}" class="btn btn-primary btn-lg">{{ .Page.Meta.CustomData.hero_button_text }}</a>
        </div>
    </div>
</div>
```

```html
<!--
@name blog-post.html
@url /blog/post/:slug:
@author Wispy Core Team
@layout default
@is_draft false
@require_auth false
@required_roles []
-->
<div class="container mx-auto px-4 py-8">
<div class="prose max-w-4xl mx-auto">
    <h2>About Wispy Core</h2>
    <p>Wispy Core is a modern, multisite content management system built with Go. It's designed to be fast, secure,
    and developer-friendly while providing a beautiful user experience.</p>

    <h3>Key Features</h3>
    <ul>
    <li><strong>Multisite Architecture:</strong> Manage multiple websites from a single installation</li>
    <li><strong>High Performance:</strong> Built with Go for exceptional speed and efficiency</li>
    <li><strong>Modern Templates:</strong> Powered by daisyUI and Tailwind CSS</li>
    <li><strong>Flexible Content:</strong> HTML-based page structure with template inheritance</li>
    <li><strong>Developer Experience:</strong> Clean architecture and extensible design</li>
    </ul>

    <h3>Getting Started</h3>
    <p>To get started with Wispy Core, check out our documentation or visit the admin panel to create your first
    pages.</p>
</div>
</div>
```
- Then for each page file we will call a function that will pull the metadata register the url in a mux for that site
- the page will be rendered to /.wispy-cache/<site_domain>/<page_slug>.html
