# Go RESTful API — User-friendly Web UI + Railway deployment starter

This is an **overlay** for https://github.com/bannakonz/go-restful-api (PR #26 merged into `main`). It is not a copy of your private backend: add these files to that existing repo.

- Plain HTML/CSS/JS, no node/npm/frontend build.
- One Gin server serves `/ui/`, `/api/v1/*`, and `/uploads/*` under the same domain.
- Railway web service + **hosted PostgreSQL** (not localhost).
- Railway persistent volume for images (`/app/uploads`).
- JWT Bearer token stays in `sessionStorage` (per browser tab). This is a learning-project tradeoff, **not** hardened public production authentication.

## Files

```text
<your-go-repository>/
  main.go           # Add ONE r.Static line (see below)
  Dockerfile        # from this starter
  .dockerignore     # from this starter
  web/
    index.html
    style.css
    app.js
  ...existing config, controllers, middleware, migrations, go.mod, go.sum...
```

## Important: use Gin, not the IDE Live Preview

If your browser URL looks like `http://localhost:63342/go-restful-api/web/index.html`, it is the JetBrains preview server, **not your Go server**. The old starter used absolute `/ui/style.css` and `/ui/app.js` paths, which made it render without any CSS on the IDE server. This version uses relative CSS/JS paths, but **API requests still expect Gin to serve the UI and API on the same origin**.

After making the small `main.go` change, open **`http://localhost:5001/ui/`** (or your configured Gin port). Do **not** develop by opening the HTML through Live Preview; the API is on a different port. DevTools > Network can help you spot missing CSS/JS assets.

This version also adds a responsive sidebar, cleaner tables, friendly details, role-aware management buttons and forms that only open when needed. There are **no frontend dependencies or build commands**.

## 1. Install UI locally

Copy `web/`, `Dockerfile`, `.dockerignore` into the root of the existing Go repository.

Add ONE line to `main.go` after the existing upload static route:

```go
r.Static("/uploads", "./uploads")
r.Static("/ui", "./web") // add this line
routes.Serve(r)
```

**Do not** use `file:///.../web/index.html`, because the UI is designed to share its origin with your Go API.

Run locally as usual with your local development `.env` (not included in this starter):

```sh
go run .
```

Open `http://127.0.0.1:5001/ui/`. The UI requests `/api/v1/...` from the same Gin server. Your existing database is used for local development only.

## 2. Deploy to Railway

Official guides: https://docs.railway.com/guides/docker-compose and https://docs.railway.com/volumes/reference

1. Commit the UI, Dockerfile, and `main.go` update to GitHub. Never commit `.env`, tokens, database secrets, or actual uploaded files.
2. In Railway, create a project and add **PostgreSQL**. Railway provisions a hosted PostgreSQL service.
3. Add a web service linked to this GitHub repo (`main` branch). It should detect the `Dockerfile`. The web service builds and runs the Go server.
4. On the **Go web service**, set variables:

   | Name | Value |
   |---|---|
   | `DATABASE_CONNECTION` | Reference the **Postgres service's `DATABASE_URL`** (using Railway's variable reference UI) |
   | `SECRET_KEY` | A cryptographically random, long value; do not use `.env.example`'s example |
   | `GIN_MODE` | `release` |
   | `HOST` | The web service's **public HTTPS origin**, e.g. `https://my-api.up.railway.app`, with no trailing `/` |
   | `PORT` | `5001` **if** you set a fixed port for the service; otherwise use Railway's injected `PORT` |

   **IMPORTANT:** Your existing `config/db.go` reads `DATABASE_CONNECTION`, **not** `DATABASE_URL`. A database reference needs to be added with the exact name `DATABASE_CONNECTION` on the Go service. The value can be a standard Postgres URL accepted by GORM/pgx. Never point production at `localhost`.
5. Under the **Go web service**, attach a persistent Railway volume mounted at exactly **`/app/uploads`**. The Go program writes to `./uploads` and the Docker image's `WORKDIR` is `/app`. Without this volume, avatars and article images can disappear on redeploy.
6. On the web service, generate a public domain under Settings > Networking. Make sure its public network port matches the port on which Gin listens (`PORT`, default 5001). Set `HOST` to that HTTPS domain and redeploy if needed.
7. Check `https://your-domain/ui/` and `https://your-domain/api/v1/categories`.

The Go process already calls `migrations.Migrate()` on startup. **Do not enable `seed.Load()` in production**: your seed drops tables and inserts demo users, including a known admin password. Create a privileged initial admin account through a separate one-time secure bootstrap procedure, not by running destructive development seed logic on production.

## 3. Production safety notes

- JWTs are stored in `sessionStorage` for simplicity and included as `Authorization: Bearer ...`. This reduces persistence between tabs, but XSS can still steal tokens. For a real public production app, implement secure `HttpOnly`, `Secure`, `SameSite` cookies and appropriate CSRF protections on the Go backend (rather than browser storage), and add a real logout/revocation mechanism. Current UI Sign out only removes the browser copy of the token.
- Do not expose DB credentials in HTML or JavaScript. Only Go uses `DATABASE_CONNECTION`.
- Backend Casbin role checks remain authoritative: Member (view public content/profile), Editor (article management), Admin (user/category/dashboard management). Frontend buttons don't grant permissions and may show 403 for restricted actions.
- Validate file types, file size, filenames, and upload quotas **on the backend**. The frontend's `accept="image/*"` attribute is not a security check.
- Persistent volumes are suitable for a single-instance demo deployment; for scalable multi-instance production use cloud object storage (e.g. S3/R2) instead, which requires updating your upload code.
- Configure automatic database backups, monitor spending, use HTTPS, keep dependencies updated, rotate secrets, and test migrations before promotion.
- Database does **not** store the uploaded images. PostgreSQL stores URLs; `/app/uploads` stores the image bytes.
- The user profile form only updates `name`, `email`, and `avatar` because those are the backend's supported fields.
- In the backend, `CategoryID` in the article PATCH is now included, and the dashboard response uses `latestArticles` after PR #26.

## 4. Route coverage (22 endpoints)

| Method | Path | UI |
|---|---|---|
| POST | `/auth/signup` | Account > Sign up |
| POST | `/auth/sign-in` | Account > Sign in |
| GET | `/auth/profile` | Account > Load profile |
| PATCH | `/auth/profile` | Account > Update profile (multipart) |
| GET | `/articles` | Articles > Search/list/filter/pagination |
| GET | `/articles/:id` | Articles > View |
| POST | `/articles` | Articles > Save without ID (multipart; image required) |
| PATCH | `/articles/:id` | Articles > Save with ID (multipart) |
| DELETE | `/articles/:id` | Articles > Delete |
| GET | `/categories` | Categories > Refresh |
| GET | `/categories/:id` | Categories > View |
| POST | `/categories` | Categories > Save without ID |
| PATCH | `/categories/:id` | Categories > Save with ID |
| DELETE | `/categories/:id` | Categories > Delete |
| GET | `/users` | Users > Search/list/pagination |
| GET | `/users/:id` | Users > View |
| POST | `/users` | Users > Save without ID |
| PATCH | `/users/:id` | Users > Save with ID |
| DELETE | `/users/:id` | Users > Delete |
| PATCH | `/users/:id/promote` | Users > Promote |
| PATCH | `/users/:id/demote` | Users > Demote |
| GET | `/dashboard` | Dashboard > Refresh |

The API tester can call any of these JSON endpoints directly. Dedicated forms handle multipart file uploads.

## 5. Checklist

- [ ] `main.go` has `r.Static("/ui", "./web")`.
- [ ] `web/` and `Dockerfile` are committed.
- [ ] Hosted PostgreSQL is provisioned and the Go service has a working `DATABASE_CONNECTION`.
- [ ] Production secret is configured, and destructive `seed.Load()` is disabled.
- [ ] A persistent volume is mounted on the **web** service at `/app/uploads`.
- [ ] `HOST` matches the public HTTPS URL.
- [ ] New users can register; secure Admin bootstrap has been completed.
- [ ] Upload an article image, redeploy, and check the image still loads.
- [ ] Try all endpoints with the proper Member, Editor, and Admin accounts.

## Cost

Railway currently advertises a $5/month Hobby base cost which is applied towards resource usage. Exact charges depend on usage and service resources; check current Railway pricing before deploying: https://docs.railway.com/pricing/plans
