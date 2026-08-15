package main

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"

	_ "modernc.org/sqlite"
)

//go:embed templates/*
var templatesFS embed.FS

const (
	sessionCookieName  = "portfolio_session"
	sessionLifetime    = 12 * time.Hour
	loginWindow        = 24 * time.Hour
	maxLoginFailures   = 5
	contactWindow      = 24 * time.Hour
	maxContactMessages = 5
)

// --- Config ---

type appConfig struct {
	EnvPath       string
	AdminUsername string
	AdminPassword string
	SessionSecret string
	DBPath        string
	BlogDir       string
}

func readEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		values[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func loadConfig(envPath string) (appConfig, error) {
	values, err := readEnvFile(envPath)
	if err != nil {
		return appConfig{}, fmt.Errorf("open env file: %w", err)
	}

	cfg := appConfig{
		EnvPath:       envPath,
		AdminUsername: values["ADMIN_USERNAME"],
		AdminPassword: values["ADMIN_PASSWORD"],
		SessionSecret: values["SESSION_SECRET"],
		DBPath:        filepath.Join("data", "main.sqlite"),
		BlogDir:       "posts",
	}
	if strings.TrimSpace(values["BLOG_DIR"]) != "" {
		cfg.BlogDir = strings.TrimSpace(values["BLOG_DIR"])
	}

	missing := make([]string, 0)
	for key, value := range map[string]string{
		"ADMIN_USERNAME": cfg.AdminUsername,
		"ADMIN_PASSWORD": cfg.AdminPassword,
		"SESSION_SECRET": cfg.SessionSecret,
	} {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return appConfig{}, fmt.Errorf("env file missing required value(s): %s", strings.Join(missing, ", "))
	}

	cfg.DBPath = filepath.Clean(cfg.DBPath)
	cfg.BlogDir = filepath.Clean(cfg.BlogDir)

	return cfg, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func cmdInit(envPath string) {
	if envPath == "" {
		envPath = filepath.Join("config", ".env")
	}
	if err := os.MkdirAll(filepath.Dir(envPath), 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll("data", 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating data directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll("posts", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating posts directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Join("static", "blog-images"), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating blog images directory: %v\n", err)
		os.Exit(1)
	}
	if _, err := os.Stat(envPath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: %s already exists\n", envPath)
		os.Exit(1)
	} else if !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "Error checking env file: %v\n", err)
		os.Exit(1)
	}

	secret, err := randomHex(32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating session secret: %v\n", err)
		os.Exit(1)
	}

	content := fmt.Sprintf("ADMIN_USERNAME=admin\nADMIN_PASSWORD=change-me-now\nSESSION_SECRET=%s\nBLOG_DIR=posts\n", secret)
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing env file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s\n", envPath)
	fmt.Println("Change ADMIN_PASSWORD before deployment.")
}

// --- Database ---

func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode = WAL;
CREATE TABLE IF NOT EXISTS login_failures (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ip TEXT NOT NULL,
  attempted_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_login_failures_ip_time
ON login_failures (ip, attempted_at);
CREATE TABLE IF NOT EXISTS contact_messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL,
  message TEXT NOT NULL,
  ip TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_contact_messages_created
ON contact_messages (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_contact_messages_ip_created
ON contact_messages (ip, created_at);`); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func purgeOldLoginFailures(db *sql.DB, now int64) error {
	_, err := db.Exec(`DELETE FROM login_failures WHERE attempted_at < ?`, now-int64(loginWindow.Seconds()))
	return err
}

func loginFailureCount(db *sql.DB, ip string, now int64) (int, error) {
	if err := purgeOldLoginFailures(db, now); err != nil {
		return 0, err
	}
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM login_failures WHERE ip = ? AND attempted_at >= ?`, ip, now-int64(loginWindow.Seconds())).Scan(&count)
	return count, err
}

func insertLoginFailure(db *sql.DB, ip string, now int64) (int, error) {
	if err := purgeOldLoginFailures(db, now); err != nil {
		return 0, err
	}
	if _, err := db.Exec(`INSERT INTO login_failures (ip, attempted_at) VALUES (?, ?)`, ip, now); err != nil {
		return 0, err
	}
	return loginFailureCount(db, ip, now)
}

type contactMessage struct {
	ID        int64
	Email     string
	Message   string
	IP        string
	CreatedAt time.Time
}

func insertContactMessage(db *sql.DB, email, message, ip string) error {
	_, err := db.Exec(
		`INSERT INTO contact_messages (email, message, ip, created_at) VALUES (?, ?, ?, ?)`,
		email,
		message,
		ip,
		time.Now().Unix(),
	)
	return err
}

func contactMessageCount(db *sql.DB, ip string, now int64) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM contact_messages WHERE ip = ? AND created_at >= ?`, ip, now-int64(contactWindow.Seconds())).Scan(&count)
	return count, err
}

func loadContactMessages(db *sql.DB) ([]contactMessage, error) {
	rows, err := db.Query(`SELECT id, email, message, ip, created_at FROM contact_messages ORDER BY created_at DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []contactMessage
	for rows.Next() {
		var msg contactMessage
		var created int64
		if err := rows.Scan(&msg.ID, &msg.Email, &msg.Message, &msg.IP, &created); err != nil {
			return nil, err
		}
		msg.CreatedAt = time.Unix(created, 0)
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// --- Sessions ---

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]time.Time
	secret   []byte
}

func newSessionStore(secret string) *sessionStore {
	return &sessionStore{
		sessions: make(map[string]time.Time),
		secret:   []byte(secret),
	}
}

func (s *sessionStore) create() (string, error) {
	id, err := randomHex(32)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	s.sessions[id] = time.Now().Add(sessionLifetime)
	s.mu.Unlock()
	return id, nil
}

func (s *sessionStore) delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func (s *sessionStore) valid(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	expires, ok := s.sessions[id]
	if !ok {
		return false
	}
	if time.Now().After(expires) {
		delete(s.sessions, id)
		return false
	}
	return true
}

func (s *sessionStore) sign(id string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(id))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *sessionStore) cookieValue(id string) string {
	return id + "." + s.sign(id)
}

func (s *sessionStore) sessionIDFromCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", false
	}
	id, sig, ok := strings.Cut(cookie.Value, ".")
	if !ok || id == "" || sig == "" {
		return "", false
	}
	if !hmac.Equal([]byte(sig), []byte(s.sign(id))) {
		return "", false
	}
	if !s.valid(id) {
		return "", false
	}
	return id, true
}

func (s *sessionStore) setCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    s.cookieValue(id),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionLifetime),
		MaxAge:   int(sessionLifetime.Seconds()),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

// --- Helpers ---

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type contactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type contactResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type blogPost struct {
	Title       string
	Subtitle    string
	Slug        string
	Date        time.Time
	DateDisplay string
	Description string
	Tags        []string
	Image       string
	Content     template.HTML
	Markdown    string
}

type postFrontMatter struct {
	Title       string   `yaml:"title"`
	Subtitle    string   `yaml:"subtitle"`
	Date        string   `yaml:"date"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Draft       bool     `yaml:"draft"`
}

func formatTime(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04 PM")
}

func formatPostDate(t time.Time) string {
	return t.Format("January 2, 2006")
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '.' || r == '/':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func parsePostDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{time.DateOnly, time.RFC3339, "2006-01-02 15:04"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q", value)
}

func splitFrontMatter(src []byte) (postFrontMatter, []byte, error) {
	var fm postFrontMatter
	trimmed := bytes.TrimPrefix(src, []byte{0xef, 0xbb, 0xbf})
	if !bytes.HasPrefix(trimmed, []byte("---\n")) {
		return fm, trimmed, errors.New("missing YAML front matter")
	}
	rest := trimmed[len("---\n"):]
	end := bytes.Index(rest, []byte("\n---\n"))
	if end < 0 {
		return fm, nil, errors.New("unterminated YAML front matter")
	}
	if err := yaml.Unmarshal(rest[:end], &fm); err != nil {
		return fm, nil, err
	}
	return fm, rest[end+len("\n---\n"):], nil
}

func markdownRenderer() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			highlighting.NewHighlighting(
				highlighting.WithStyle("native"),
				highlighting.WithGuessLanguage(true),
				highlighting.WithFormatOptions(chromahtml.WithClasses(false)),
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
	)
}

func renderMarkdown(src []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := markdownRenderer().Convert(src, &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

func featuredImageForSlug(slug string) string {
	for _, ext := range []string{".webp", ".jpg", ".jpeg", ".png", ".gif", ".avif"} {
		path := filepath.Join("static", "blog-images", slug+ext)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return "/static/blog-images/" + slug + ext
		}
	}
	return ""
}

func loadBlogPostFromFile(path string) (blogPost, bool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return blogPost{}, false, err
	}
	fm, body, err := splitFrontMatter(src)
	if err != nil {
		return blogPost{}, false, fmt.Errorf("%s: %w", path, err)
	}
	if fm.Draft {
		return blogPost{}, false, nil
	}
	title := strings.TrimSpace(fm.Title)
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	date, err := parsePostDate(fm.Date)
	if err != nil {
		return blogPost{}, false, fmt.Errorf("%s: %w", path, err)
	}
	content, err := renderMarkdown(body)
	if err != nil {
		return blogPost{}, false, fmt.Errorf("%s: %w", path, err)
	}
	slug := slugify(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	if slug == "" {
		return blogPost{}, false, fmt.Errorf("%s: invalid slug", path)
	}
	return blogPost{
		Title:       title,
		Subtitle:    strings.TrimSpace(fm.Subtitle),
		Slug:        slug,
		Date:        date,
		DateDisplay: formatPostDate(date),
		Description: strings.TrimSpace(fm.Description),
		Tags:        fm.Tags,
		Image:       featuredImageForSlug(slug),
		Content:     content,
		Markdown:    string(src),
	}, true, nil
}

func loadBlogPosts(postsDir string) ([]blogPost, error) {
	entries, err := os.ReadDir(postsDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	posts := make([]blogPost, 0)
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".md" {
			continue
		}
		post, ok, err := loadBlogPostFromFile(filepath.Join(postsDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if ok {
			posts = append(posts, post)
		}
	}
	sort.Slice(posts, func(i, j int) bool {
		if posts[i].Date.Equal(posts[j].Date) {
			return posts[i].Title < posts[j].Title
		}
		return posts[i].Date.After(posts[j].Date)
	})
	return posts, nil
}

func loadBlogPost(postsDir, slug string) (blogPost, bool, error) {
	posts, err := loadBlogPosts(postsDir)
	if err != nil {
		return blogPost{}, false, err
	}
	for _, post := range posts {
		if post.Slug == slug {
			return post, true, nil
		}
	}
	return blogPost{}, false, nil
}

func cmdNewPost(title, postsDir string) {
	title = strings.TrimSpace(title)
	if title == "" {
		fmt.Fprintln(os.Stderr, "Error: post title cannot be empty")
		os.Exit(1)
	}
	slug := slugify(title)
	if slug == "" {
		fmt.Fprintln(os.Stderr, "Error: post title must contain at least one letter or number")
		os.Exit(1)
	}
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating posts directory: %v\n", err)
		os.Exit(1)
	}
	path := filepath.Join(postsDir, slug+".md")
	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(os.Stderr, "Error: %s already exists\n", path)
		os.Exit(1)
	} else if !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "Error checking post path: %v\n", err)
		os.Exit(1)
	}
	content := fmt.Sprintf(`---
title: %q
subtitle: ""
date: %s
description: ""
tags: []
draft: false
---

Write the post here.
`, title, time.Now().Format(time.DateOnly))
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing post: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s\n", path)
	fmt.Printf("Featured image path: static/blog-images/%s.webp\n", slug)
}

func printUsage() {
	fmt.Println(`Usage: portfolio [command]

Commands:

  portfolio                         Start the server on port 8112
  init [env-file]                   Create config/.env and data/ by default
  new-post "Post Title" [posts-dir] Create a Markdown blog post draft
  serve <port> [env-file]           Start the server with config/.env by default`)
}

func cmdServe(port, envPath string) {
	cfg, err := loadConfig(envPath)
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	db, err := openDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	sessions := newSessionStore(cfg.SessionSecret)

	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"formatTime": formatTime,
	}).ParseFS(templatesFS, "templates/*.html"))

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	requireAuth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if _, ok := sessions.sessionIDFromCookie(r); !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			next(w, r)
		}
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/blog" {
			http.NotFound(w, r)
			return
		}
		posts, err := loadBlogPosts(cfg.BlogDir)
		if err != nil {
			log.Printf("Failed to load blog posts: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		data := struct {
			Posts []blogPost
		}{Posts: posts}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "blog.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/blog/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.Trim(strings.TrimPrefix(r.URL.Path, "/blog/"), "/")
		if slug == "" || strings.Contains(slug, "/") {
			http.NotFound(w, r)
			return
		}
		post, ok, err := loadBlogPost(cfg.BlogDir, slug)
		if err != nil {
			log.Printf("Failed to load blog post %q: %v", slug, err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.NotFound(w, r)
			return
		}
		data := struct {
			Post blogPost
		}{Post: post}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "post.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/contact", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(contactResponse{Message: "Method not allowed"})
			return
		}

		ip := clientIP(r)
		now := time.Now().Unix()
		count, err := contactMessageCount(db, ip, now)
		if err != nil {
			log.Printf("Failed to check contact message limit: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(contactResponse{Message: "Server error. Please try again later."})
			return
		}
		if count >= maxContactMessages {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(contactWindow.Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(contactResponse{Message: "Too many messages sent today. Please try again tomorrow."})
			return
		}

		var req contactRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(contactResponse{Message: "Invalid request body"})
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		req.Email = strings.TrimSpace(req.Email)
		req.Message = strings.TrimSpace(req.Message)

		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(contactResponse{Message: "Please provide your name"})
			return
		}
		if len(req.Name) > 120 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(contactResponse{Message: "Name must be 120 characters or fewer"})
			return
		}
		if req.Email == "" || !emailRegex.MatchString(req.Email) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(contactResponse{Message: "Please provide a valid email address"})
			return
		}
		if req.Message == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(contactResponse{Message: "Please provide a message"})
			return
		}
		if len(req.Message) > 1200 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(contactResponse{Message: "Message must be 1200 characters or fewer"})
			return
		}

		storedMessage := fmt.Sprintf("Name: %s\n\n%s", req.Name, req.Message)
		if err := insertContactMessage(db, req.Email, storedMessage, ip); err != nil {
			log.Printf("Failed to save contact message: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(contactResponse{Message: "Failed to save message. Please try again later."})
			return
		}

		json.NewEncoder(w).Encode(contactResponse{OK: true, Message: "Message received. I'll get back to you soon."})
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sessions.sessionIDFromCookie(r); ok && r.Method == http.MethodGet {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		data := struct {
			Error string
		}{}

		switch r.Method {
		case http.MethodGet:
		case http.MethodPost:
			ip := clientIP(r)
			now := time.Now().Unix()
			count, err := loginFailureCount(db, ip, now)
			if err != nil {
				log.Printf("Failed to check login failures: %v", err)
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}
			if count >= maxLoginFailures {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			username := r.FormValue("username")
			password := r.FormValue("password")

			if username == cfg.AdminUsername && password == cfg.AdminPassword {
				id, err := sessions.create()
				if err != nil {
					log.Printf("Failed to create session: %v", err)
					http.Error(w, "Server error", http.StatusInternalServerError)
					return
				}
				sessions.setCookie(w, id)
				http.Redirect(w, r, "/admin", http.StatusSeeOther)
				return
			}

			count, err = insertLoginFailure(db, ip, now)
			if err != nil {
				log.Printf("Failed to record login failure: %v", err)
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}
			if count >= maxLoginFailures {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			data.Error = "Invalid username or password."
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "login.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if id, ok := sessions.sessionIDFromCookie(r); ok {
			sessions.delete(id)
		}
		clearSessionCookie(w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	mux.HandleFunc("/admin", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin" {
			http.NotFound(w, r)
			return
		}
		messages, err := loadContactMessages(db)
		if err != nil {
			log.Printf("Failed to load contact messages: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		data := struct {
			Messages []contactMessage
		}{Messages: messages}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "admin.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))

	addr := ":" + port
	log.Printf("Listening on %s with env %s, db %s, and blog dir %s", addr, cfg.EnvPath, cfg.DBPath, cfg.BlogDir)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		cmdServe("8112", filepath.Join("config", ".env"))
		return
	}

	switch args[0] {
	case "init":
		envPath := filepath.Join("config", ".env")
		if len(args) > 1 {
			envPath = args[1]
		}
		cmdInit(envPath)

	case "new-post":
		if len(args) < 2 || len(args) > 3 {
			printUsage()
			os.Exit(1)
		}
		postsDir := "posts"
		if len(args) == 3 {
			postsDir = args[2]
		}
		cmdNewPost(args[1], postsDir)

	case "serve":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: serve requires a port")
			fmt.Fprintln(os.Stderr, "Usage: portfolio serve <port> [env-file]")
			os.Exit(1)
		}
		envPath := filepath.Join("config", ".env")
		if len(args) > 2 {
			envPath = args[2]
		}
		cmdServe(args[1], envPath)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		printUsage()
		os.Exit(1)
	}
}
