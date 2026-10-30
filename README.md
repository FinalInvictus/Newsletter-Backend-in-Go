# Fieldnotes Newsletter

A small newsletter publication built with Go, PostgreSQL, and a responsive static frontend.

## Run locally

1. Create a PostgreSQL database and apply `db/migrations/001_init.sql`.
2. Set `DATABASE_URL` (and optionally `PORT`, which defaults to `8080`) in the environment or a `.env` file.
3. Run `go run ./cmd/server` from the repository root.
4. Open `http://localhost:8080`.

The same server serves the files in `web/` and the JSON API. The public page lists published stories, accepts newsletter signups, and lets readers comment on stories. The Editorial desk can create and edit stories, create categories, and view subscribers.

## API routes

- `GET /articles` and `GET /articles/{slug}` — published stories
- `GET /articles/all`, `POST /articles`, `PUT /articles/{id}`, `DELETE /articles/{id}` — editorial story management
- `GET /categories`, `POST /categories` — category management
- `POST /subscribe`, `GET /subscribers` — newsletter signup and subscriber listing
- `GET /articles/{id}/comments`, `POST /articles/{id}/comments` — story discussion
- `GET /health` — database-backed health check

The editor and its management endpoints currently have no authentication. Add access control before exposing this server to the public internet.
