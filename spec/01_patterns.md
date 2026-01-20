### 01_patterns.md

Pattern: PageData Render Pattern
- Build a PageData struct with page-specific fields (Title, ActiveNav, etc.) and AdminAuthenticated set from the request when appropriate.
- Execute the relevant HTML template from the templates map with the PageData instance.

Pattern: Route Path Guard Pattern
- If the request path does not exactly match the expected path (with optional trailing slash where permitted), respond with NotFound and return.

Pattern: HTTP Method Switch Pattern
- Switch on r.Method and handle only the allowed methods.
- For unsupported methods, respond with NotFound or MethodNotAllowed depending on the handler.

Pattern: Contact Abuse Defense Pattern
- Resolve the client IP.
- Check the blacklist table for the given IP and action; if blocked, return a 429 response.
- Track recent requests in memory; if abusive within the burst window, blacklist the IP and return a 429 response.
- Increment the daily counter in the database; if the limit is exceeded, blacklist the IP and return a 429 response.

Pattern: Admin Abuse Defense Pattern
- Same steps as Contact Abuse Defense Pattern but for the admin_login and admin_login_failed actions.
- On rate limit breach, render the admin login page with a throttling error message.

Pattern: Template Parsing Pattern
- Read templates/partials.html once.
- For each page template name, read the file, parse partials + page, register functions, and store in the templates map.

Pattern: Markdown Load-and-Render Pattern
- Attempt to read content in priority order (markdown then HTML file, or remote URLs in order).
- On a successful read, run the markdown renderer into a buffer and return the buffer contents as template.HTML.
- If all candidates fail, return a fallback HTML message.

Pattern: Admin Cookie Auth Pattern
- Compute a token by hashing ADMIN_USERNAME:ADMIN_PASSWORD from env (SHA-256 hex).
- Set or clear an HttpOnly admin_auth cookie on the root path.
- Authenticate by comparing the cookie value with the computed token using constant-time compare.

Pattern: SQLite Initialization Pattern
- Determine the database path from CONTACT_DB_PATH or a default under data/.
- Ensure the database directory exists.
- Open, ping, and initialize tables if they do not already exist.

Pattern: SQLite Query Pattern
- For inserts and deletes, execute parameterized SQL via ExecContext.
- For reads, query rows, scan into structs, collect into slice, and check rows.Err().

Pattern: Env .env Loader Pattern
- Read .env if it exists.
- Ignore comments and blank lines.
- Strip leading "export" if present.
- Parse key=value pairs, strip optional surrounding quotes, and set only if not already set in the process environment.
