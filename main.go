package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
}

var templates map[string]*template.Template

func main() {
	templates = parseTemplates()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/projects", projectsHandler)
	http.HandleFunc("/project/", projectHandler)
	http.HandleFunc("/blog", blogHandler)
	http.HandleFunc("/blog/", blogPostHandler)
	http.HandleFunc("/contact", contactHandler)

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
	for _, page := range pages {
		path := filepath.Join("templates", page)
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", page, err)
			continue
		}
		tmpl := template.Must(template.New(page).Funcs(funcs).Parse(string(content)))
		result[page] = tmpl
	}
	return result
}

func getProjects() []Project {
	return []Project{
		{Name: "gtml", Description: "A compact Go-first HTML templating tool.", Slug: "gtml"},
		{Name: "godocument", Description: "Documentation tooling built for Go projects.", Slug: "godocument"},
		{Name: "inp", Description: "Minimal input helpers for command-line tools.", Slug: "inp"},
		{Name: "xerus", Description: "A focused experiment in language tooling.", Slug: "xerus"},
		{Name: "sniper", Description: "Automation and scraping utilities.", Slug: "sniper"},
		{Name: "bible-bot", Description: "A bot for serving Bible verses in chat.", Slug: "bible-bot"},
		{Name: "pride", Description: "Terminal styling helpers and experiments.", Slug: "pride"},
		{Name: "flint", Description: "Small tooling for configuration and logging.", Slug: "flint"},
	}
}

func getBlogPosts() []BlogPost {
	return []BlogPost{
		{
			Title:   "Designing With Constraints",
			Slug:    "designing-with-constraints",
			Date:    "2026-01-15",
			Excerpt: "Why small constraints lead to sharper, calmer interfaces.",
		},
		{
			Title:   "Go Templates in the Small",
			Slug:    "go-templates-in-the-small",
			Date:    "2026-01-07",
			Excerpt: "A practical, no-framework way to ship fast.",
		},
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := PageData{
		Title:      "Phillip England",
		GitHubUser: "phillip-england",
		Projects:   getProjects(),
		BlogPosts:  getBlogPosts(),
	}
	templates["index.html"].Execute(w, data)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/about" {
		http.NotFound(w, r)
		return
	}
	data := PageData{
		Title:      "About - Phillip England",
		GitHubUser: "phillip-england",
	}
	templates["about.html"].Execute(w, data)
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/projects" {
		http.NotFound(w, r)
		return
	}
	data := PageData{
		Title:      "Projects - Phillip England",
		GitHubUser: "phillip-england",
		Projects:   getProjects(),
	}
	templates["projects.html"].Execute(w, data)
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
	}
	templates["project.html"].Execute(w, data)
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
	}
	templates["post.html"].Execute(w, data)
}

func loadBlogContent(slug string) template.HTML {
	path := filepath.Join("content", "blog", slug+".html")
	data, err := os.ReadFile(path)
	if err != nil {
		return template.HTML("<p>Post coming soon.</p>")
	}
	return template.HTML(data)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := PageData{
			Title:      "Contact - Phillip England",
			GitHubUser: "phillip-england",
		}
		templates["contact.html"].Execute(w, data)
	case http.MethodPost:
		contactPostHandler(w, r)
	default:
		http.NotFound(w, r)
	}
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
