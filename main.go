package main

import (
	"bufio"
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
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed projects.json
var projectsJSON []byte

const (
	sessionCookieName = "portfolio_session"
	sessionLifetime   = 12 * time.Hour
	loginWindow       = 24 * time.Hour
	maxLoginFailures  = 5
)

// --- Config ---

type appConfig struct {
	EnvPath       string
	AdminUsername string
	AdminPassword string
	SessionSecret string
	DBPath        string
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

	content := fmt.Sprintf("ADMIN_USERNAME=admin\nADMIN_PASSWORD=change-me-now\nSESSION_SECRET=%s\n", secret)
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
  created_at INTEGER NOT NULL,
  read_at INTEGER
);
CREATE INDEX IF NOT EXISTS idx_contact_messages_created
ON contact_messages (created_at DESC);`); err != nil {
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
	ReadAt    sql.NullInt64
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

func loadContactMessages(db *sql.DB) ([]contactMessage, error) {
	rows, err := db.Query(`SELECT id, email, message, ip, created_at, read_at FROM contact_messages ORDER BY created_at DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []contactMessage
	for rows.Next() {
		var msg contactMessage
		var created int64
		if err := rows.Scan(&msg.ID, &msg.Email, &msg.Message, &msg.IP, &created, &msg.ReadAt); err != nil {
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
	Email   string `json:"email"`
	Message string `json:"message"`
}

type contactResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type project struct {
	Number      int      `json:"-"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	SiteURL     string   `json:"siteUrl"`
	SourceURL   string   `json:"sourceUrl"`
}

func loadProjects() ([]project, error) {
	var projects []project
	if err := json.Unmarshal(projectsJSON, &projects); err != nil {
		return nil, fmt.Errorf("parse projects.json: %w", err)
	}
	for i := range projects {
		projects[i].Number = i + 1
	}
	return projects, nil
}

func formatTime(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04 PM")
}

func printUsage() {
	fmt.Println(`Usage: portfolio [command]

Commands:

  portfolio                    Start the server on port 8112
  init [env-file]              Create config/.env and data/ by default
  serve <port> [env-file]      Start the server with config/.env by default`)
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

	projects, err := loadProjects()
	if err != nil {
		log.Fatalf("Failed to load projects: %v", err)
	}

	mux := http.NewServeMux()

	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub-filesystem: %v", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

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
		data := struct{ Projects []project }{Projects: projects}
		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
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

		var req contactRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(contactResponse{Message: "Invalid request body"})
			return
		}

		req.Email = strings.TrimSpace(req.Email)
		req.Message = strings.TrimSpace(req.Message)

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
		if len(req.Message) > 255 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(contactResponse{Message: "Message must be 255 characters or fewer"})
			return
		}

		if err := insertContactMessage(db, req.Email, req.Message, clientIP(r)); err != nil {
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
	log.Printf("Listening on %s with env %s and db %s", addr, cfg.EnvPath, cfg.DBPath)
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
