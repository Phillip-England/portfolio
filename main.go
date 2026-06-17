package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed blog/*.md
var blogFS embed.FS

// --- .env loader ---

func loadEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
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
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func readEnvValue(path, key string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
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
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		v = strings.Trim(v, `"'`)
		if k == key {
			return v
		}
	}
	return ""
}

// --- Rate limiter ---

type rateLimiter struct {
	mu      sync.Mutex
	records map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{records: make(map[string][]time.Time)}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-24 * time.Hour)
	timestamps := rl.records[ip]
	// Prune old entries
	filtered := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	rl.records[ip] = filtered
	return len(filtered) < 3
}

func (rl *rateLimiter) record(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.records[ip] = append(rl.records[ip], time.Now())
}

func (rl *rateLimiter) reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.records = make(map[string][]time.Time)
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(1 * time.Hour)
		rl.mu.Lock()
		cutoff := time.Now().Add(-24 * time.Hour)
		for ip, timestamps := range rl.records {
			filtered := timestamps[:0]
			for _, t := range timestamps {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(rl.records, ip)
			} else {
				rl.records[ip] = filtered
			}
		}
		rl.mu.Unlock()
	}
}

// --- Helpers ---

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
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

type blogPost struct {
	Slug        string
	Title       string
	Date        string
	Description string
	HTML        template.HTML
}

func sendPostmarkEmail(apiKey, from, to, subject, body string) error {
	payload := map[string]string{
		"From":     from,
		"To":       to,
		"Subject":  subject,
		"TextBody": body,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "https://api.postmarkapp.com/email", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("postmark returned status %d", resp.StatusCode)
	}
	return nil
}

func markdownRenderer() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
}

func parseFrontMatter(src []byte) (map[string]string, []byte) {
	meta := make(map[string]string)
	text := strings.TrimPrefix(string(src), "\ufeff")
	if !strings.HasPrefix(text, "---\n") {
		return meta, []byte(text)
	}
	rest := text[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return meta, []byte(text)
	}
	for _, line := range strings.Split(rest[:end], "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		meta[strings.ToLower(strings.TrimSpace(key))] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	body := strings.TrimLeft(rest[end+len("\n---"):], "\r\n")
	return meta, []byte(body)
}

func titleFromMarkdown(src []byte, fallback string) string {
	scanner := bufio.NewScanner(bytes.NewReader(src))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return fallback
}

func loadBlogPosts() ([]blogPost, map[string]blogPost, error) {
	entries, err := fs.ReadDir(blogFS, "blog")
	if err != nil {
		return nil, nil, err
	}

	md := markdownRenderer()
	posts := make([]blogPost, 0, len(entries))
	bySlug := make(map[string]blogPost)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		raw, err := blogFS.ReadFile("blog/" + entry.Name())
		if err != nil {
			return nil, nil, err
		}

		slug := strings.TrimSuffix(entry.Name(), ".md")
		meta, body := parseFrontMatter(raw)
		var rendered bytes.Buffer
		if err := md.Convert(body, &rendered); err != nil {
			return nil, nil, err
		}

		title := meta["title"]
		if title == "" {
			title = titleFromMarkdown(body, slug)
		}

		post := blogPost{
			Slug:        slug,
			Title:       title,
			Date:        meta["date"],
			Description: meta["description"],
			HTML:        template.HTML(rendered.String()),
		}
		posts = append(posts, post)
		bySlug[slug] = post
	}

	sort.Slice(posts, func(i, j int) bool {
		if posts[i].Date == posts[j].Date {
			return posts[i].Title < posts[j].Title
		}
		return posts[i].Date > posts[j].Date
	})

	return posts, bySlug, nil
}

// --- CLI commands ---

func printUsage() {
	fmt.Println(`Usage: portfolio <command>

Commands:
  serve <port>           Start the server on the given port
  set-key <key>          Set the Postmark API key in .env
  reset-ips [--port N]   Reset all rate-limited IPs (default port: 8080)`)
}

func cmdServe(port string) {
	loadEnv(".env")

	postmarkKey := os.Getenv("POSTMARK_API_KEY")
	fromEmail := os.Getenv("POSTMARK_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "admin@phillip-england.com"
	}
	toEmail := "admin@phillip-england.com"
	adminSecret := os.Getenv("ADMIN_SECRET")

	limiter := newRateLimiter()

	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
	blogPosts, blogBySlug, err := loadBlogPosts()
	if err != nil {
		log.Fatalf("Failed to load blog posts: %v", err)
	}

	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub-filesystem: %v", err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/blog" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := struct {
			Posts []blogPost
		}{Posts: blogPosts}
		if err := tmpl.ExecuteTemplate(w, "blog.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/blog/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.Trim(strings.TrimPrefix(r.URL.Path, "/blog/"), "/")
		if slug == "" {
			http.Redirect(w, r, "/blog", http.StatusMovedPermanently)
			return
		}
		if strings.Contains(slug, "/") {
			http.NotFound(w, r)
			return
		}
		post, ok := blogBySlug[slug]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "post.html", post); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/api/contact", func(w http.ResponseWriter, r *http.Request) {
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

		ip := clientIP(r)
		if !limiter.allow(ip) {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(contactResponse{Message: "You've sent too many messages. Please try again later."})
			return
		}

		if postmarkKey == "" {
			log.Println("POSTMARK_API_KEY not set, skipping email send")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(contactResponse{Message: "Email service is not configured"})
			return
		}

		subject := "Portfolio Contact: New Message"
		body := fmt.Sprintf("From: %s\n\nMessage:\n%s", req.Email, req.Message)

		if err := sendPostmarkEmail(postmarkKey, fromEmail, toEmail, subject, body); err != nil {
			log.Printf("Failed to send email: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(contactResponse{Message: "Failed to send message. Please try again later."})
			return
		}

		limiter.record(ip)
		json.NewEncoder(w).Encode(contactResponse{OK: true, Message: "Message sent! I'll get back to you soon."})
	})

	http.HandleFunc("/api/admin/reset-ips", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(contactResponse{Message: "Method not allowed"})
			return
		}

		if adminSecret == "" {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(contactResponse{Message: "ADMIN_SECRET not configured"})
			return
		}

		auth := r.Header.Get("Authorization")
		if auth != adminSecret {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(contactResponse{Message: "Unauthorized"})
			return
		}

		limiter.reset()
		json.NewEncoder(w).Encode(contactResponse{OK: true, Message: "All rate-limited IPs have been reset"})
	})

	addr := ":" + port
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func cmdSetKey(key string) {
	envPath := ".env"
	targetLine := "POSTMARK_API_KEY=" + key

	content, err := os.ReadFile(envPath)
	if err != nil {
		// File doesn't exist, create it
		if err := os.WriteFile(envPath, []byte(targetLine+"\n"), 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing .env: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Postmark API key set in .env")
		return
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "POSTMARK_API_KEY=") {
			lines[i] = targetLine
			found = true
			break
		}
	}
	if !found {
		// Append, ensuring there's a newline before our new line
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines[len(lines)-1] = targetLine
			lines = append(lines, "")
		} else {
			lines = append(lines, targetLine)
		}
	}

	if err := os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing .env: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Postmark API key set in .env")
}

func cmdResetIPs(port string) {
	secret := readEnvValue(".env", "ADMIN_SECRET")
	if secret == "" {
		fmt.Fprintln(os.Stderr, "Error: ADMIN_SECRET not found in .env")
		os.Exit(1)
	}

	url := fmt.Sprintf("http://localhost:%s/api/admin/reset-ips", port)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", secret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result contactResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.OK {
		fmt.Println(result.Message)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Message)
		os.Exit(1)
	}
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "serve":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: serve requires a port argument")
			fmt.Fprintln(os.Stderr, "Usage: portfolio serve <port>")
			os.Exit(1)
		}
		cmdServe(args[1])

	case "set-key":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: set-key requires a key argument")
			fmt.Fprintln(os.Stderr, "Usage: portfolio set-key <key>")
			os.Exit(1)
		}
		cmdSetKey(args[1])

	case "reset-ips":
		port := "8080"
		for i := 1; i < len(args); i++ {
			if args[i] == "--port" && i+1 < len(args) {
				port = args[i+1]
				i++
			}
		}
		cmdResetIPs(port)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		printUsage()
		os.Exit(1)
	}
}
