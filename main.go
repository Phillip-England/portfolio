package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	_ "modernc.org/sqlite"
)

type Project struct {
	Name        string
	Description string
	Slug        string
}

type BlogPost struct {
	Title   string
	Slug    string
	Date    string
	Excerpt string
	Content template.HTML
}

type PageData struct {
	Title          string
	GitHubUser     string
	Projects       []Project
	BlogPosts      []BlogPost
	CurrentPost    BlogPost
	CurrentProject Project
	ProfileImage   string
	ReadmeHTML     template.HTML
	ActiveNav      string
	AdminError     string
	AdminAuthenticated bool
	Messages       []ContactMessage
}

var templates map[string]*template.Template
var contactDB *sql.DB
var recentRequestsMu sync.Mutex
var recentRequests = make(map[string][]time.Time)

const (
	dailyLimitContact     = 5
	dailyLimitAdminFailed = 5
	blacklistDuration     = 24 * time.Hour
	recentPoolSize        = 20
	recentAbuseWindow     = 10 * time.Second
)
var markdown = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		highlighting.NewHighlighting(
			highlighting.WithStyle("dracula"),
			highlighting.WithGuessLanguage(true),
		),
	),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

func main() {
	loadDotEnv()
	templates = parseTemplates()
	var err error
	contactDB, err = initContactDB()
	if err != nil {
		log.Fatalf("failed to initialize contact database: %v", err)
	}
	defer contactDB.Close()

	http.HandleFunc("/project/", projectHandler)
	http.HandleFunc("/projects", projectsHandler)
	http.HandleFunc("/blog", blogHandler)
	http.HandleFunc("/blog/", blogPostHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/contact", contactHandler)
	http.HandleFunc("/admin", adminLoginHandler)
	http.HandleFunc("/admin/messages", adminMessagesHandler)
	http.HandleFunc("/admin/messages/delete", adminDeleteMessageHandler)
	http.HandleFunc("/admin/logout", adminLogoutHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	addr := ":" + port

	log.Printf("Server running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

func loadDotEnv() {
	file, err := os.Open(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Printf("failed to read .env: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			log.Printf("failed to set .env var %s: %v", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("failed to parse .env: %v", err)
	}
}

func parseTemplates() map[string]*template.Template {
	pages := []string{
		"index.html",
		"about.html",
		"projects.html",
		"project.html",
		"blog.html",
		"post.html",
		"contact.html",
		"admin_login.html",
		"admin_messages.html",
	}

	funcs := template.FuncMap{
		"toLower": strings.ToLower,
	}

	result := make(map[string]*template.Template)
	partialsPath := filepath.Join("templates", "partials.html")
	partialsContent, err := os.ReadFile(partialsPath)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", partialsPath, err)
	}
	for _, page := range pages {
		path := filepath.Join("templates", page)
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", page, err)
			continue
		}
		tmpl := template.Must(template.New(page).Funcs(funcs).Parse(string(partialsContent) + string(content)))
		result[page] = tmpl
	}
	return result
}

type ContactMessage struct {
	ID        int
	Email     string
	Message   string
	CreatedAt string
}

func getProjects() []Project {
	return []Project{
		{Name: "gtml", Description: "An HTML compiler.", Slug: "gtml"},
		{Name: "godocument", Description: "HTMX-based documentation generator.", Slug: "godocument"},
		{Name: "inp", Description: "Control the mouse and keyboard with shell commands.", Slug: "inp"},
		{Name: "xerus", Description: "A Bun-first TypeScript web framework.", Slug: "xerus"},
		{Name: "sniper", Description: "A speech to action system built on gorobot.", Slug: "sniper"},
		{Name: "bible-bot", Description: "Scrape the whole Bible from bible.com and store it in a SQLite database.", Slug: "bible-bot"},
		{Name: "pride", Description: "Generate static sites from Markdown files.", Slug: "pride"},
		{Name: "flint", Description: "Generate static sites from a MPA server using simple config.", Slug: "flint"},
	}
}

func getBlogPosts() []BlogPost {
	return []BlogPost{
		{
			Title:   "Concurrency in Go: Practical Patterns",
			Slug:    "concurrency-in-go",
			Date:    "2026-01-19",
			Excerpt: "A quick tour of goroutines, channels, and production-ready patterns.",
		},
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := PageData{
			Title:        "Phillip England",
			GitHubUser:   "phillip-england",
			Projects:     getProjects(),
			BlogPosts:    getBlogPosts(),
			ProfileImage: "/static/profile.jpeg",
			ActiveNav:    "home",
			AdminAuthenticated: isAdminAuthenticated(r),
		}
		templates["index.html"].Execute(w, data)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/contact" && r.URL.Path != "/contact/" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		data := PageData{
			Title:      "Contact - Phillip England",
			GitHubUser: "phillip-england",
			ActiveNav:  "contact",
			AdminAuthenticated: isAdminAuthenticated(r),
		}
		templates["contact.html"].Execute(w, data)
	case http.MethodPost:
		ip := clientIP(r)
		if ip == "" {
			ip = "unknown"
		}
		if isBlocked, err := isBlacklisted(r.Context(), ip, "contact_message"); err != nil {
			log.Printf("failed to check blacklist for %s: %v", ip, err)
		} else if isBlocked {
			rateLimitJSON("Too many messages from this IP today.")(w, r)
			return
		}
		if abusive, window := trackRecentRequest(ip, "contact_message"); abusive {
			if err := blacklistIP(r.Context(), ip, "contact_message", blacklistDuration); err != nil {
				log.Printf("failed to blacklist %s: %v", ip, err)
			}
			log.Printf("blacklisted %s for contact_message abuse over %s", ip, window)
			rateLimitJSON("Too many messages from this IP today.")(w, r)
			return
		}
		contactPostHandler(w, r, ip)
	default:
		http.NotFound(w, r)
	}
}

func projectHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/project/")
	projects := getProjects()
	var project Project
	found := false
	for _, p := range projects {
		if p.Slug == slug {
			project = p
			found = true
			break
		}
	}
	if !found {
		http.NotFound(w, r)
		return
	}
	data := PageData{
		Title:          fmt.Sprintf("%s - Project Details", project.Name),
		GitHubUser:     "phillip-england",
		CurrentProject: project,
		ReadmeHTML:     loadProjectReadme("phillip-england", project.Slug),
		ActiveNav:      "projects",
		AdminAuthenticated: isAdminAuthenticated(r),
	}
	templates["project.html"].Execute(w, data)
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/projects" && r.URL.Path != "/projects/" {
		http.NotFound(w, r)
		return
	}
	data := PageData{
		Title:      "Projects - Phillip England",
		GitHubUser: "phillip-england",
		Projects:   getProjects(),
		ActiveNav:  "projects",
		AdminAuthenticated: isAdminAuthenticated(r),
	}
	templates["projects.html"].Execute(w, data)
}

func blogHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/blog" {
		http.NotFound(w, r)
		return
	}
	data := PageData{
		Title:      "Blog - Phillip England",
		GitHubUser: "phillip-england",
		BlogPosts:  getBlogPosts(),
		ActiveNav:  "blog",
		AdminAuthenticated: isAdminAuthenticated(r),
	}
	templates["blog.html"].Execute(w, data)
}

func blogPostHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/blog/")
	posts := getBlogPosts()
	var post BlogPost
	for _, p := range posts {
		if p.Slug == slug {
			post = p
			break
		}
	}
	if post.Title == "" {
		http.NotFound(w, r)
		return
	}
	post.Content = loadBlogContent(post.Slug)
	data := PageData{
		Title:       fmt.Sprintf("%s - Blog", post.Title),
		GitHubUser:  "phillip-england",
		CurrentPost: post,
		ActiveNav:   "blog",
		AdminAuthenticated: isAdminAuthenticated(r),
	}
	templates["post.html"].Execute(w, data)
}

func loadBlogContent(slug string) template.HTML {
	paths := []string{
		filepath.Join("content", "blog", slug+".md"),
		filepath.Join("content", "blog", slug+".html"),
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var buf bytes.Buffer
		if err := markdown.Convert(data, &buf); err != nil {
			return template.HTML("<p>Post coming soon.</p>")
		}
		return template.HTML(buf.String())
	}
	return template.HTML("<p>Post coming soon.</p>")
}

func loadProjectReadme(user, repo string) template.HTML {
	candidates := []string{
		fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/README.md", user, repo),
		fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/README.md", user, repo),
	}

	for _, url := range candidates {
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		var buf bytes.Buffer
		if err := markdown.Convert(body, &buf); err != nil {
			continue
		}
		return template.HTML(buf.String())
	}

	return template.HTML("<p class=\"text-[#555]\">Could not load README.md from GitHub.</p>")
}

func rateLimitJSON(message string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, message, http.StatusTooManyRequests)
	}
}

func rateLimitAdminLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
	renderAdminLogin(w, "Too many login attempts. Try again tomorrow.")
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func contactPostHandler(w http.ResponseWriter, r *http.Request, ip string) {
	var req struct {
		Email   string `json:"email"`
		Message string `json:"message"`
		Date    string `json:"createdAt"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Message = strings.TrimSpace(req.Message)
	req.Date = strings.TrimSpace(req.Date)

	if req.Email == "" || req.Message == "" || len(req.Message) > 255 {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if req.Date == "" {
		req.Date = time.Now().UTC().Format(time.RFC3339)
	}

	allowed, err := incrementDailyCount(r.Context(), ip, "contact_message", dailyLimitContact)
	if err != nil {
		log.Printf("failed to update contact rate limit for %s: %v", ip, err)
	} else if !allowed {
		if err := blacklistIP(r.Context(), ip, "contact_message", blacklistDuration); err != nil {
			log.Printf("failed to blacklist %s: %v", ip, err)
		}
		rateLimitJSON("Too many messages from this IP today.")(w, r)
		return
	}
	if err := insertContactMessage(r.Context(), req.Email, req.Message, req.Date); err != nil {
		log.Printf("failed to store contact message: %v", err)
		http.Error(w, "Failed to store message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Message received"})
}

func incrementDailyCount(ctx context.Context, ip, action string, limit int) (bool, error) {
	if contactDB == nil {
		return true, fmt.Errorf("database not initialized")
	}
	day := time.Now().UTC().Format("2006-01-02")
	tx, err := contactDB.BeginTx(ctx, nil)
	if err != nil {
		return true, err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO rate_limits (ip, action, day, count)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(ip, action, day) DO UPDATE SET count = count + 1
	`, ip, action, day)
	if err != nil {
		tx.Rollback()
		return true, err
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count FROM rate_limits WHERE ip = ? AND action = ? AND day = ?`, ip, action, day).Scan(&count); err != nil {
		tx.Rollback()
		return true, err
	}
	if err := tx.Commit(); err != nil {
		return true, err
	}
	return count <= limit, nil
}

func isBlacklisted(ctx context.Context, ip, action string) (bool, error) {
	if contactDB == nil {
		return false, fmt.Errorf("database not initialized")
	}
	var untilStr string
	err := contactDB.QueryRowContext(ctx, `SELECT blocked_until FROM ip_blacklist WHERE ip = ? AND action = ?`, ip, action).Scan(&untilStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	until, err := time.Parse(time.RFC3339, untilStr)
	if err != nil {
		_, _ = contactDB.ExecContext(ctx, `DELETE FROM ip_blacklist WHERE ip = ? AND action = ?`, ip, action)
		return false, nil
	}
	now := time.Now().UTC()
	if now.Before(until) {
		return true, nil
	}
	_, _ = contactDB.ExecContext(ctx, `DELETE FROM ip_blacklist WHERE ip = ? AND action = ?`, ip, action)
	return false, nil
}

func blacklistIP(ctx context.Context, ip, action string, duration time.Duration) error {
	if contactDB == nil {
		return fmt.Errorf("database not initialized")
	}
	until := time.Now().UTC().Add(duration).Format(time.RFC3339)
	_, err := contactDB.ExecContext(ctx, `
		INSERT INTO ip_blacklist (ip, action, blocked_until)
		VALUES (?, ?, ?)
		ON CONFLICT(ip, action) DO UPDATE SET blocked_until = excluded.blocked_until
	`, ip, action, until)
	return err
}

func trackRecentRequest(ip, action string) (bool, time.Duration) {
	now := time.Now().UTC()
	key := ip + "|" + action
	recentRequestsMu.Lock()
	defer recentRequestsMu.Unlock()
	window := time.Duration(0)
	list := recentRequests[key]
	list = append(list, now)
	if len(list) > recentPoolSize {
		list = list[len(list)-recentPoolSize:]
	}
	recentRequests[key] = list
	if len(list) < recentPoolSize {
		return false, window
	}
	window = now.Sub(list[0])
	return window <= recentAbuseWindow, window
}

func initContactDB() (*sql.DB, error) {
	dbPath := os.Getenv("CONTACT_DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join("data", "contact_messages.db")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	const schema = `
	CREATE TABLE IF NOT EXISTS contact_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rate_limits (
		ip TEXT NOT NULL,
		action TEXT NOT NULL,
		day TEXT NOT NULL,
		count INTEGER NOT NULL,
		PRIMARY KEY (ip, action, day)
	);
	CREATE TABLE IF NOT EXISTS ip_blacklist (
		ip TEXT NOT NULL,
		action TEXT NOT NULL,
		blocked_until TEXT NOT NULL,
		PRIMARY KEY (ip, action)
	);`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func insertContactMessage(ctx context.Context, email, message, createdAt string) error {
	if contactDB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := contactDB.ExecContext(ctx, `INSERT INTO contact_messages (email, message, created_at) VALUES (?, ?, ?)`, email, message, createdAt)
	return err
}

func adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if isAdminAuthenticated(r) {
		http.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		renderAdminLogin(w, "")
	case http.MethodPost:
		ip := clientIP(r)
		if ip == "" {
			ip = "unknown"
		}
		if isBlocked, err := isBlacklisted(r.Context(), ip, "admin_login"); err != nil {
			log.Printf("failed to check admin blacklist for %s: %v", ip, err)
		} else if isBlocked {
			rateLimitAdminLogin(w, r)
			return
		}
		if abusive, window := trackRecentRequest(ip, "admin_login"); abusive {
			if err := blacklistIP(r.Context(), ip, "admin_login", blacklistDuration); err != nil {
				log.Printf("failed to blacklist %s: %v", ip, err)
			}
			log.Printf("blacklisted %s for admin_login abuse over %s", ip, window)
			rateLimitAdminLogin(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			renderAdminLogin(w, "Invalid form submission.")
			return
		}
		username := strings.TrimSpace(r.FormValue("username"))
		password := strings.TrimSpace(r.FormValue("password"))
		if !validateAdminCredentials(username, password) {
			allowed, err := incrementDailyCount(r.Context(), ip, "admin_login_failed", dailyLimitAdminFailed)
			if err != nil {
				log.Printf("failed to update admin rate limit for %s: %v", ip, err)
			} else if !allowed {
				if err := blacklistIP(r.Context(), ip, "admin_login", blacklistDuration); err != nil {
					log.Printf("failed to blacklist %s: %v", ip, err)
				}
				rateLimitAdminLogin(w, r)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			renderAdminLogin(w, "Invalid credentials.")
			return
		}
		setAdminCookie(w)
		http.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
	default:
		http.NotFound(w, r)
	}
}

func adminMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	messages, err := fetchContactMessages(r.Context())
	if err != nil {
		log.Printf("failed to fetch messages: %v", err)
		http.Error(w, "Failed to load messages", http.StatusInternalServerError)
		return
	}
	data := PageData{
		Title:     "Admin Messages - Phillip England",
		Messages:  messages,
		ActiveNav: "admin",
		AdminAuthenticated: true,
	}
	templates["admin_messages.html"].Execute(w, data)
}

func adminDeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}
	idStr := strings.TrimSpace(r.FormValue("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid message id", http.StatusBadRequest)
		return
	}
	if err := deleteContactMessage(r.Context(), id); err != nil {
		log.Printf("failed to delete message %d: %v", id, err)
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
}

func adminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	clearAdminCookie(w)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func renderAdminLogin(w http.ResponseWriter, errMsg string) {
	data := PageData{
		Title:      "Admin Login - Phillip England",
		AdminError: errMsg,
		ActiveNav:  "admin",
	}
	templates["admin_login.html"].Execute(w, data)
}

func fetchContactMessages(ctx context.Context) ([]ContactMessage, error) {
	if contactDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	rows, err := contactDB.QueryContext(ctx, `SELECT id, email, message, created_at FROM contact_messages ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ContactMessage
	for rows.Next() {
		var msg ContactMessage
		if err := rows.Scan(&msg.ID, &msg.Email, &msg.Message, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func deleteContactMessage(ctx context.Context, id int) error {
	if contactDB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := contactDB.ExecContext(ctx, `DELETE FROM contact_messages WHERE id = ?`, id)
	return err
}

func adminCredentials() (string, string, bool) {
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")
	if username == "" || password == "" {
		return "", "", false
	}
	return username, password, true
}

func validateAdminCredentials(username, password string) bool {
	expectedUser, expectedPass, ok := adminCredentials()
	if !ok {
		return false
	}
	userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(expectedUser)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(expectedPass)) == 1
	return userMatch && passMatch
}

func adminToken() (string, bool) {
	username, password, ok := adminCredentials()
	if !ok {
		return "", false
	}
	sum := sha256.Sum256([]byte(username + ":" + password))
	return hex.EncodeToString(sum[:]), true
}

func setAdminCookie(w http.ResponseWriter) {
	token, ok := adminToken()
	if !ok {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_auth",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearAdminCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_auth",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func isAdminAuthenticated(r *http.Request) bool {
	token, ok := adminToken()
	if !ok {
		return false
	}
	cookie, err := r.Cookie("admin_auth")
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(token)) == 1
}
