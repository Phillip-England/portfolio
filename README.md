# Portfolio

## Run

```sh
go run . init
go run . serve 8080
```

Open `http://localhost:8080`.

The `init` command creates `config/.env` and prepares the `data/` directory. The
server reads `config/.env` by default and stores SQLite data at
`data/main.sqlite`. The environment file is required at startup and contains:

```env
ADMIN_USERNAME=admin
ADMIN_PASSWORD=change-me-now
SESSION_SECRET=<random-secret>
```

Change `ADMIN_PASSWORD` before deploying. The SQLite database stores contact
messages and the IP-based login-failure ledger.

## Admin

The contact form saves messages internally instead of sending email. Log in at
`/login` and read messages at `/admin`.

Failed admin login attempts are tracked by remote IP in SQLite. An IP is blocked
after 5 failed attempts within 24 hours. Contact form submissions are also
limited to 5 accepted messages per IP within 24 hours. Old login-failure rows
are purged during normal login checks and failure inserts.

## Public site

The homepage is a personal engineering profile for Phillip England. It explains
his background, software learning path, technical interests, and professional
direction. It intentionally does not render a project gallery because
applications live at `https://apps.phillip-england.com`.
