# Portfolio

## Run

```sh
go run . serve 8080
```

Open `http://localhost:8080`.

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
