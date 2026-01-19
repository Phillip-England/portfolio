package main

import (
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
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	Messages       []ContactMessage
}

var templates map[string]*template.Template
var contactDB *sql.DB
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
	templates = parseTemplates()
	var err error
	contactDB, err = initContactDB()
	if err != nil {
		log.Fatalf("failed to initialize contact database: %v", err)
	}
	defer contactDB.Close()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/project/", projectHandler)
	http.HandleFunc("/projects", projectsHandler)
	http.HandleFunc("/blog", blogHandler)
	http.HandleFunc("/blog/", blogPostHandler)
	http.HandleFunc("/contact", contactHandler)
	http.HandleFunc("/admin", adminLoginHandler)
	http.HandleFunc("/admin/messages", adminMessagesHandler)

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
		}
		templates["index.html"].Execute(w, data)
	case http.MethodPost:
		contactPostHandler(w, r)
	default:
		http.NotFound(w, r)
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
		}
		templates["contact.html"].Execute(w, data)
	case http.MethodPost:
		contactPostHandler(w, r)
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

func contactPostHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := insertContactMessage(r.Context(), req.Email, req.Message, req.Date); err != nil {
		log.Printf("failed to store contact message: %v", err)
		http.Error(w, "Failed to store message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Message received"})
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
		if err := r.ParseForm(); err != nil {
			renderAdminLogin(w, "Invalid form submission.")
			return
		}
		username := strings.TrimSpace(r.FormValue("username"))
		password := strings.TrimSpace(r.FormValue("password"))
		if !validateAdminCredentials(username, password) {
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
	}
	templates["admin_messages.html"].Execute(w, data)
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
