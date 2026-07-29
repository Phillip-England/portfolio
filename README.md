# Portfolio

## Run

```sh
go run . init
go run . serve 8080 config/.env
```

Open `http://localhost:8080`.

The `init` command creates `config/.env` and prepares the `data/` directory. The
environment file is required at startup and contains:

```env
ADMIN_USERNAME=admin
ADMIN_PASSWORD=change-me-now
SESSION_SECRET=<random-secret>
DB_PATH=../data/main.sqlite
```

Change `ADMIN_PASSWORD` before deploying. The SQLite database stores contact
messages and the IP-based login-failure ledger.

## Admin

The contact form saves messages internally instead of sending email. Log in at
`/login` and read messages at `/admin`.

Failed admin login attempts are tracked by remote IP in SQLite. An IP is blocked
after 5 failed attempts within 24 hours. Old login-failure rows are purged during
normal login checks and failure inserts.

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
