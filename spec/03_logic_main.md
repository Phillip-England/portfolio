### 03_logic_main.md

Globals and configuration
- templates: map[string]*template.Template shared across handlers
- contactDB: *sql.DB SQLite connection used for contact messages and rate limits
- recentRequestsMu: mutex protecting recentRequests
- recentRequests: map from "ip|action" to a slice of request timestamps for burst detection
- Constants: dailyLimitContact=5, dailyLimitAdminFailed=5, blacklistDuration=24h, recentPoolSize=20, recentAbuseWindow=10s
- markdown renderer: goldmark configured with GFM, code highlighting, and unsafe HTML rendering

Function: main() returns none
- Call loadDotEnv.
- Assign templates = parseTemplates().
- Initialize contactDB via initContactDB; if error, log.Fatal and exit.
- Defer contactDB.Close.
- Register HTTP handlers for project, projects, blog, blog post, index, contact, admin login/messages/delete/logout routes.
- Register /static/ file server rooted at ./static.
- Read PORT from environment; if empty, use "8000".
- Build address by prefixing ":" to the port.
- Log server start URL.
- Start http.ListenAndServe with the address and nil handler; on error, log.Fatal and exit.

Function: loadDotEnv() returns none
- Open .env for reading.
- If file does not exist, return.
- On other open errors, log and return.
- Scan the file line by line.
- Trim whitespace from the line; skip if empty or starts with #.
- If line starts with "export ", strip that prefix and trim again.
- Split the line into key and value at the first "="; skip if split does not yield two parts.
- Trim whitespace around key and value; skip if key is empty.
- If value is wrapped in matching single or double quotes, strip the surrounding quotes.
- If the key is already set in the environment, skip it.
- Set the environment variable; log on error.
- After scanning, if scanner error is non-nil, log it.

Function: parseTemplates() returns map[string]*template.Template
- Define the list of page template filenames.
- Define a FuncMap with toLower mapped to strings.ToLower.
- Read templates/partials.html into memory; log on error.
- For each page file:
  - Read the template file; on error, log and continue to next.
  - Create a new template with the page name, register FuncMap, and parse partials + page content together.
  - Store the compiled template in the result map using the page name as key.
- Return the map.

Function: getProjects() returns []Project
- Return a static slice of Project values with Name, Description, and Slug for each project.

Function: getBlogPosts() returns []BlogPost
- Return a static slice of BlogPost values with Title, Slug, Date, and Excerpt fields populated.

Function: baseURL(r *http.Request) returns string
- Set scheme to "http" by default.
- If r.TLS is non-nil, set scheme to "https".
- If X-Forwarded-Proto header is present, split by comma, trim the first value, and use it as scheme.
- Return scheme + "://" + r.Host.

Function: absoluteURL(r *http.Request, path string) returns string
- If path starts with "http://" or "https://", return path unchanged.
- If path does not start with "/", prepend "/".
- Return baseURL(r) + path.

Function: indexHandler(w http.ResponseWriter, r *http.Request) returns none
- Use the HTTP Method Switch Pattern.
- On GET:
  - Set profileImage to "/static/profile.jpeg".
  - Populate PageData with title, meta description, canonical URL, social image URL, GitHub user, projects, blog posts, profile image, ActiveNav="home", and AdminAuthenticated from isAdminAuthenticated.
  - Apply the PageData Render Pattern for index.html.
- On other methods: respond with MethodNotAllowed.

Function: contactHandler(w http.ResponseWriter, r *http.Request) returns none
- Apply the Route Path Guard Pattern for /contact and /contact/.
- Use the HTTP Method Switch Pattern.
- On GET:
  - Build PageData with title, GitHub user, ActiveNav="contact", and AdminAuthenticated.
  - Apply the PageData Render Pattern for contact.html.
- On POST:
  - Determine client IP via clientIP; if empty, set to "unknown".
  - Apply the Contact Abuse Defense Pattern.
  - If allowed, call contactPostHandler with the resolved IP.
- On other methods: respond with NotFound.

Function: projectHandler(w http.ResponseWriter, r *http.Request) returns none
- Extract slug by trimming the "/project/" prefix from r.URL.Path.
- Iterate getProjects to find a matching slug.
- If not found, respond with NotFound and return.
- Build PageData with title derived from project name, GitHub user, CurrentProject, ReadmeHTML from loadProjectReadme, ActiveNav="projects", and AdminAuthenticated.
- Apply the PageData Render Pattern for project.html.

Function: projectsHandler(w http.ResponseWriter, r *http.Request) returns none
- Apply the Route Path Guard Pattern for /projects and /projects/.
- Build PageData with title, GitHub user, projects list, ActiveNav="projects", and AdminAuthenticated.
- Apply the PageData Render Pattern for projects.html.

Function: blogHandler(w http.ResponseWriter, r *http.Request) returns none
- Apply the Route Path Guard Pattern for /blog only.
- Build PageData with title, GitHub user, blog posts list, ActiveNav="blog", and AdminAuthenticated.
- Apply the PageData Render Pattern for blog.html.

Function: blogPostHandler(w http.ResponseWriter, r *http.Request) returns none
- Extract slug by trimming the "/blog/" prefix.
- Search getBlogPosts for a matching slug.
- If no match, respond with NotFound and return.
- Load markdown content via loadBlogContent and assign to post.Content.
- Build PageData with title derived from post title, GitHub user, CurrentPost, ActiveNav="blog", and AdminAuthenticated.
- Apply the PageData Render Pattern for post.html.

Function: loadBlogContent(slug string) returns template.HTML
- Build candidate paths for content/blog/slug.md and content/blog/slug.html.
- For each candidate path:
  - Try to read the file; if read fails, continue.
  - Convert file content to HTML via markdown renderer into a buffer.
  - If conversion fails, return the fallback HTML "Post coming soon".
  - Return the buffer contents as template.HTML.
- If no candidate succeeds, return the fallback HTML.

Function: loadProjectReadme(user, repo string) returns template.HTML
- Build candidate URLs for GitHub raw README.md in main and master branches.
- For each URL:
  - HTTP GET the URL; on error, continue.
  - If status code is not 200, close the body and continue.
  - Read the response body; close the body; on error, continue.
  - Convert markdown to HTML via the markdown renderer into a buffer; on error, continue.
  - Return the buffer contents as template.HTML.
- If no candidate succeeds, return the fallback HTML indicating failure.

Function: rateLimitJSON(message string) returns func(http.ResponseWriter, *http.Request)
- Return a closure that writes an HTTP 429 response with the provided message using http.Error.

Function: rateLimitAdminLogin(w http.ResponseWriter, r *http.Request) returns none
- Write HTTP status 429.
- Render admin login page with a throttling error message.

Function: clientIP(r *http.Request) returns string
- If X-Forwarded-For header exists, split by comma and return the first trimmed value.
- Else if X-Real-IP header exists, return the trimmed value.
- Else parse r.RemoteAddr using net.SplitHostPort; if a host is found, return it.
- Fallback: return trimmed r.RemoteAddr.

Function: contactPostHandler(w http.ResponseWriter, r *http.Request, ip string) returns none
- Decode JSON body into a struct with Email, Message, and Date fields.
- If JSON decode fails, respond with HTTP 400 and return.
- Trim whitespace on Email, Message, and Date.
- If Email is empty, Message is empty, or Message length exceeds 255, respond with HTTP 400 and return.
- If Date is empty, set it to current UTC time in RFC3339.
- Call incrementDailyCount with action "contact_message" and dailyLimitContact.
- If incrementDailyCount returns an error, log it and continue.
- If incrementDailyCount indicates not allowed:
  - Blacklist the IP for contact_message.
  - Respond with 429 using rateLimitJSON and return.
- Insert the contact message into the database; on error, log and respond with HTTP 500.
- On success, set Content-Type to application/json and encode a JSON success payload.

Function: incrementDailyCount(ctx context.Context, ip, action string, limit int) returns (bool, error)
- If contactDB is nil, return true with an error.
- Compute day as current UTC date formatted as YYYY-MM-DD.
- Begin a database transaction.
- Insert or update the rate_limits row by incrementing count for the ip+action+day tuple; on error, rollback and return true with error.
- Query the current count for the ip+action+day tuple; on error, rollback and return true with error.
- Commit the transaction; on error, return true with error.
- Return whether count is less than or equal to limit.

Function: isBlacklisted(ctx context.Context, ip, action string) returns (bool, error)
- If contactDB is nil, return false with an error.
- Query blocked_until for the ip+action.
- If no rows, return false nil.
- If another error, return false with error.
- Parse blocked_until as RFC3339; on parse failure, delete the row and return false nil.
- If current UTC time is before blocked_until, return true nil.
- Otherwise delete the row and return false nil.

Function: blacklistIP(ctx context.Context, ip, action string, duration time.Duration) returns error
- If contactDB is nil, return error.
- Compute blocked_until as current UTC time plus duration formatted as RFC3339.
- Insert or update ip_blacklist with the new blocked_until.

Function: trackRecentRequest(ip, action string) returns (bool, time.Duration)
- Capture current UTC time.
- Build key as ip + "|" + action.
- Lock recentRequestsMu and defer unlock.
- Append current time to the list for the key.
- If list exceeds recentPoolSize, keep only the most recent recentPoolSize entries.
- Store the updated list back into recentRequests.
- If the list length is less than recentPoolSize, return false and a zero duration.
- Compute window as now minus the oldest timestamp.
- Return whether window is less than or equal to recentAbuseWindow, plus the window duration.

Function: initContactDB() returns (*sql.DB, error)
- Read CONTACT_DB_PATH; if empty, set to data/contact_messages.db.
- Ensure the parent directory exists with mode 0755.
- Open SQLite database using driver name "sqlite".
- Ping the database; on failure, close and return error.
- Execute schema creation SQL for contact_messages, rate_limits, and ip_blacklist tables.
- On schema error, close and return error.
- Return the database handle.

Function: insertContactMessage(ctx context.Context, email, message, createdAt string) returns error
- If contactDB is nil, return error.
- Execute an insert into contact_messages with the provided fields.

Function: adminLoginHandler(w http.ResponseWriter, r *http.Request) returns none
- If isAdminAuthenticated is true, redirect to /admin/messages with StatusSeeOther and return.
- Use the HTTP Method Switch Pattern.
- On GET: renderAdminLogin with empty error message.
- On POST:
  - Determine client IP; if empty set to "unknown".
  - Apply the Admin Abuse Defense Pattern for the admin_login action.
  - Parse the form; if error, renderAdminLogin with form error and return.
  - Read and trim username and password from form fields.
  - If validateAdminCredentials is false:
    - Increment the daily count for admin_login_failed; log any error.
    - If not allowed, blacklist admin_login and rate limit response.
    - Set HTTP status to 401 and renderAdminLogin with invalid credentials.
    - Return.
  - Set the admin cookie and redirect to /admin/messages with StatusSeeOther.
- On other methods: respond with NotFound.

Function: adminMessagesHandler(w http.ResponseWriter, r *http.Request) returns none
- If not authenticated, redirect to /admin with StatusSeeOther and return.
- If method is not GET, respond with NotFound and return.
- Fetch messages via fetchContactMessages; on error, log and respond with HTTP 500.
- Build PageData with title, messages, ActiveNav="admin", and AdminAuthenticated=true.
- Apply the PageData Render Pattern for admin_messages.html.

Function: adminDeleteMessageHandler(w http.ResponseWriter, r *http.Request) returns none
- If not authenticated, redirect to /admin with StatusSeeOther and return.
- If method is not POST, respond with NotFound and return.
- Parse the form; on error, respond with HTTP 400 and return.
- Read id from form, trim, and parse to int.
- If parse fails or id <= 0, respond with HTTP 400 and return.
- Delete the message via deleteContactMessage; on error, log and respond with HTTP 500.
- Redirect to /admin/messages with StatusSeeOther.

Function: adminLogoutHandler(w http.ResponseWriter, r *http.Request) returns none
- If not authenticated, redirect to /admin with StatusSeeOther and return.
- If method is not POST, respond with NotFound and return.
- Clear the admin cookie and redirect to /admin with StatusSeeOther.

Function: renderAdminLogin(w http.ResponseWriter, errMsg string) returns none
- Build PageData with title, AdminError set to errMsg, and ActiveNav="admin".
- Apply the PageData Render Pattern for admin_login.html.

Function: fetchContactMessages(ctx context.Context) returns ([]ContactMessage, error)
- If contactDB is nil, return error.
- Query contact_messages ordered by id descending.
- Iterate over rows, scanning each into a ContactMessage and appending to a slice.
- If rows iteration returns an error, return it.
- Return the slice of messages.

Function: deleteContactMessage(ctx context.Context, id int) returns error
- If contactDB is nil, return error.
- Execute a delete on contact_messages by id.

Function: adminCredentials() returns (string, string, bool)
- Read ADMIN_USERNAME and ADMIN_PASSWORD from environment.
- If either is empty, return empty values and false.
- Return username, password, and true.

Function: validateAdminCredentials(username, password string) returns bool
- Load expected username and password via adminCredentials; if not ok, return false.
- Compare the provided username to expected username using constant-time compare.
- Compare the provided password to expected password using constant-time compare.
- Return true only if both comparisons match.

Function: adminToken() returns (string, bool)
- Load credentials via adminCredentials; if not ok, return empty string and false.
- Compute SHA-256 of "username:password".
- Return hex-encoded token string and true.

Function: setAdminCookie(w http.ResponseWriter) returns none
- Compute admin token; if unavailable, return.
- Set admin_auth cookie with token value, path "/", HttpOnly true, and SameSite=Strict.

Function: clearAdminCookie(w http.ResponseWriter) returns none
- Set admin_auth cookie with empty value, path "/", HttpOnly true, SameSite=Strict, MaxAge=-1, and Expires set to Unix epoch.

Function: isAdminAuthenticated(r *http.Request) returns bool
- Compute admin token; if unavailable, return false.
- Read admin_auth cookie; on error, return false.
- Compare cookie value with token using constant-time compare and return the result.
