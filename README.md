# Portfolio

## Run

```sh
go run . serve 8080
```

Open `http://localhost:8080`.

## Projects

Projects shown on the homepage live in `projects.json`. Add another object to the
array and restart the site:

```json
{
  "name": "Project name",
  "description": "What it does and why it matters.",
  "tags": ["Go", "JavaScript"],
  "image": "/static/project-images/project-name.png",
  "imageAlt": "Project name website preview",
  "sourceUrl": "https://github.com/your-name/project",
  "sourceLabel": "View source",
  "docsUrl": "https://docs.example.com/",
  "docsLabel": "View docs",
  "featured": false
}
```

Use an empty link field to omit that link. The older `url` and `linkLabel` fields
still work for projects with a single link. Set `featured` to `true` to give a
project extra visual emphasis.

## Blog

The blog is powered by markdown files in the `blog/` directory. To add a post, create a new `.md` file:

```text
blog/my-new-post.md
```

Use this format:

````md
---
title: My New Post
date: 2026-06-17
description: Short summary shown on the blog index.
---

Write the post in markdown.

```go
fmt.Println("code blocks use the Dracula theme")
```
````

The filename becomes the URL slug. For example, `blog/my-new-post.md` is available at `/blog/my-new-post`.

The blog index is at `/blog`, and every blog page includes a link back to the main site.
