package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
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
}

var templates map[string]*template.Template
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

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/project/", projectHandler)
	http.HandleFunc("/projects", projectsHandler)
	http.HandleFunc("/blog", blogHandler)
	http.HandleFunc("/blog/", blogPostHandler)

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

	if req.Email == "" || req.Message == "" || len(req.Message) > 255 {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	fmt.Printf("Contact form submission: Email=%s, Message=%s, Date=%s\n", req.Email, req.Message, req.Date)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Message received"})
}
