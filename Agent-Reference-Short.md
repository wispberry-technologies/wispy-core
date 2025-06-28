**API Development Prompt**  

**Requirements:**  
- **Requests:** Use `application/x-www-form-urlencoded` (default), JSON only for nested data. Validate with `go-playground/validator`.  
- **Responses:**  
  - Errors → Plain text only (`respondWithError`, `PlainTextError`).  
  - Debug info → Enabled via `__include_debug_info__` (query/header).  
  - Success → Format-appropriate (JSON/HTML).  
- **Errors:** Log internally, user-friendly messages, correct HTTP codes (400, 401, etc.).  
- **Utilities:**  
  - `shouldIncludeDebugInfo()` – Checks debug flag.  
  - `redirectWithMessage()` – Safe redirects with params.  
  - `NormalizeHost()` – Standardizes hostnames.  
  - `RequestLogger()` – Middleware for request logging.  

**Rules:**  
- No stack traces in production.  
- Use provided utils for consistency.  
- Isolate tenant data in multi-tenant setups.  

*(Keep responses secure, structured, and logged.)*