# Wispy Core Project Standards

This document outlines the coding standards, architectural decisions, and best practices that must be followed by all contributors to the Wispy Core project.

## API Standards

### Request Format

- **All API input request data MUST be encoded as `application/x-www-form-urlencoded`**
- JSON input or response should NOT BE USED! unless specifically allowed/told
- Use URL query parameters for filtering and pagination
- Validate all incoming data against a defined schema using `go-playground/validator` [link](https://github.com/go-playground/validator)
### Response Format

- **All API error responses MUST be in plain text**
- Error responses should include debug information when requested by the client via:
  - Query parameter: `__include_debug_info__=true`
  - HTTP header: `__include_debug_info__: true`
- Success responses can be in appropriate formats based on the endpoint (HTML, JSON, etc.)

### Error Handling

- All errors must be properly logged
- User-facing error messages should be informative but not expose sensitive system details
- HTTP status codes must be used correctly (e.g., 400 for bad requests, 401 for unauthorized, etc.)

## Code Organization

### Directory Structure

- Follow the established project directory structure
- Place related functionality in the same package
- Use lowercase package names
- Test files should be in the same directory as the code they test with a `_test.go` suffix

### File Naming

- Use lowercase with underscores for file names
- File names should be descriptive of their contents
- Implementation files should have concise names (e.g., `auth.go`)
- Interface files should be named with their purpose (e.g., `interfaces.go`)

## Go Coding Standards

### General

- Follow [Go's official style guide](https://golang.org/doc/effective_go)
- Use `gofmt` to format all code
- Maximum line length of 100 characters
- Config files should be in toml format, not JSON or YAML

### Documentation

- All exported functions, variables, constants, and types MUST have documentation comments
- Comments should explain WHY, not WHAT (the code should be self-explanatory)
- Use meaningful variable and function names

### Dependencies

- Minimize external dependencies
- Regularly update dependencies for security fixes

## Security Standards

- All user input MUST be validated and sanitized
- Use prepared statements for database queries
- Store passwords using secure hashing algorithms (bcrypt or better)
- Implement proper CSRF protection for web forms
- Set appropriate security headers in HTTP responses
- TLS must be configured securely for production deployments

## Testing

- All new code MUST have unit tests
- Aim for at least 70% test coverage
- Use table-driven tests where appropriate
- Include integration tests for critical paths
- Tests should be fast and not depend on external services

## Git Workflow

- Use feature branches for all changes
- Write meaningful commit messages
- Rebase feature branches on main before creating pull requests
- Pull requests require at least one code review before merging

## Template and Front-End Standards

- Use consistent naming conventions in HTML templates
- Follow a mobile-first responsive design approach
- Minimize JavaScript usage where possible
- CSS should be modular and follow BEM naming convention (https://getbem.com/)

## Database

- Document all schema changes
- Queries must be parameterized to prevent SQL injection
- Sql statements should be stored as constants or variables instead of hardcoded strings
- Use migrations for schema changes
- Follow naming conventions for tables and columns
- Indexes should be added for frequently queried fields
- Use transactions for operations that modify multiple tables

## Performance

- Optimize database queries
- Use connection pooling
- Implement caching for expensive operations
- Set appropriate timeouts for all external service calls

## Accessibility
- All user interfaces should meet WCAG 2.1 AA standards
- Use semantic HTML
- Ensure keyboard navigation works correctly
- Provide alternative text for images

This document is subject to change and will be updated as the project evolves. All team members are encouraged to suggest improvements.
