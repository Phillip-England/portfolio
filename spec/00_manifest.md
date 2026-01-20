### 00_manifest.md

Project structure (workspace root)
- main.go: single Go entry point containing all handlers, data access, and utility logic
- go.mod, go.sum: module and dependency definitions
- content/: markdown/html content for blog posts (runtime-loaded from disk)
- data/: SQLite database file for contact messages and rate limits
- static/: static assets served under /static/ (images and JS)
- templates/: HTML templates including partials and page layouts
- Makefile, PROMPT.md, .env, .gitignore, .dockerignore: project tooling and configuration

Entry point
- main.go contains package main and the main() function; it registers HTTP routes and starts the server

External dependencies and usage
- github.com/yuin/goldmark: markdown parser used to render blog content and GitHub README content to HTML
- github.com/yuin/goldmark-highlighting/v2: syntax highlighting extension for markdown rendering
- github.com/yuin/goldmark/extension: provides GitHub Flavored Markdown extension used in renderer setup
- github.com/yuin/goldmark/renderer/html: renderer options to allow unsafe HTML output
- modernc.org/sqlite: SQLite driver used by database/sql to store contact messages and rate limiting data

Architecture summary
- Monolithic net/http server with handler functions in main package
- HTML template rendering for pages; JSON response for contact POST
- SQLite-backed persistence for contact messages, rate limiting, and IP blacklist
- Simple admin auth via cookie with SHA-256 token derived from env credentials
- Runtime markdown rendering for blog posts and GitHub README content
