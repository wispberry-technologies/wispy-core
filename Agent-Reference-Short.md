**API Development Prompt**  

**Requirements:**  
- **Requests:** Use `application/x-www-form-urlencoded` (default), JSON only for nested data. Validate with `go-playground/validator`.  
- **Responses:**  
  - Errors → Plain text only (`RespondWithError`, `PlainTextError`).  
  - Debug info → Enabled via `__include_debug_info__` (query/header).  
  - Success → Format-appropriate (JSON/HTML).  
- **Errors:** Log internally, user-friendly messages, correct HTTP codes (400, 401, etc.).  
- **Utilities:**  
  - `common.ShouldIncludeDebugInfo()` – Checks debug flag.  
  - `common.RedirectWithMessage()` – Safe redirects with params.  
  - `common.PlainTextError()` – Standard error response in plain text.
  - `common.RespondWithError()` – Standard error response.
  - `common.RespondWithJSON()` – Standard JSON response.
  - `common.NormalizeHost()` – Standardizes hostnames.  
  - `common.RequestLogger()` – Middleware for request logging.  

**Rules:**  
- No stack traces in production.  
- Use provided utils for consistency.  
- Isolate tenant data in multi-tenant setups.  

*(Keep responses secure, structured, and logged.)*