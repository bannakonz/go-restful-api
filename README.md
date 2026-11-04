
# Go RESTful API

A RESTful API built with **Go, Gin, GORM, and PostgreSQL**, featuring JWT authentication, role-based access control (RBAC), CRUD operations, file uploads, and database migrations.

The project also includes a lightweight web management interface built with HTML, CSS, and JavaScript.

## Features

- JWT Authentication (Sign Up, Sign In, Profile)
- Role-Based Access Control (Admin, Editor, Member)
- User Management
- Article CRUD Operations
- Category CRUD Operations
- File Uploads (Article Images and User Avatars)
- Dashboard Statistics
- Database Migrations
- Web Management Interface
- Docker Support
- Railway Deployment Support

## Tech Stack

| Technology | Purpose |
|---|---|
| Go | Backend programming language |
| Gin | HTTP web framework |
| GORM | Object-Relational Mapping |
| PostgreSQL | Relational database |
| JWT | Authentication |
| Casbin | Role-based authorization |
| Gormigrate | Database migrations |
| HTML / CSS / JavaScript | Web management interface |
| Docker | Containerization |
| Railway | Deployment platform |

## Project Structure

```text
go-restful-api/
├── config/          # Database and RBAC configuration
├── controllers/     # HTTP request handlers
├── middleware/      # Authentication and authorization
├── migrations/      # Database schema migrations
├── models/          # GORM data models
├── routes/          # API route definitions
├── seed/            # Development seed data
├── web/             # Web UI (HTML, CSS, JavaScript)
├── main.go          # Application entry point
├── Dockerfile       # Docker configuration
├── go.mod           # Go module dependencies
├── go.sum           # Dependency checksums
└── README.md
```

## Prerequisites

Before running the application, make sure you have:

- Go 1.27.1 (as specified in `go.mod`)
- PostgreSQL
- Git
- Docker (optional)

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/bannakonz/go-restful-api.git
cd go-restful-api
```

### 2. Configure Environment Variables

Create a `.env` file from the example:

```bash
cp .env.example .env
```

Update the database configuration in `.env` to match your local PostgreSQL setup.

Example:

```dotenv
HOST=http://localhost:5001
PORT=5001
DATABASE_CONNECTION="host=localhost user=postgres password=your_password dbname=go_restful_api port=5432 sslmode=disable"
SECRET_KEY=replace_with_a_secure_random_secret
GIN_MODE=debug
```

Generate a random JWT secret:

```bash
openssl rand -base64 32
```

Replace `SECRET_KEY` with the generated value.

Never commit your real `.env` file or production credentials.

### 3. Create the Database

Make sure PostgreSQL is running.

Create a database named:

```sql
CREATE DATABASE go_restful_api;
```

You can execute this command using `psql`, pgAdmin, or another PostgreSQL client.

The application runs database migrations automatically on startup.

### 4. Install Dependencies

```bash
go mod download
```

### 5. Run the Application

```bash
go run .
```

By default, the application runs on port `5001`.

- **Server:** http://localhost:5001
- **Web UI:** http://localhost:5001/ui/
- **API Base URL:** http://localhost:5001/api/v1

> The Web UI should be accessed through the Gin server, not by opening `web/index.html` directly or using an IDE Live Preview server.

## Environment Variables

| Variable | Description |
|---|---|
| `HOST` | Public base URL used when generating uploaded file URLs |
| `PORT` | HTTP server port (default: `5001`) |
| `DATABASE_CONNECTION` | PostgreSQL connection string |
| `SECRET_KEY` | Secret key used for JWT signing |
| `GIN_MODE` | Gin mode (`debug` or `release`) |

The application loads environment variables from `.env` using `godotenv`.

If no `.env` file exists, it uses environment variables provided by the operating system or deployment platform.

## API Documentation

Base URL:

```text
http://localhost:5001/api/v1
```

### Authentication

| Method | Endpoint | Description | Access |
|---|---|---|---|
| POST | `/auth/signup` | Register a new user | Public |
| POST | `/auth/sign-in` | Authenticate and receive JWT | Public |
| GET | `/auth/profile` | Get current user profile | Authenticated |
| PATCH | `/auth/profile` | Update profile and avatar | Authenticated |

### Articles

| Method | Endpoint | Description | Access |
|---|---|---|---|
| GET | `/articles` | List articles with search and pagination | Public |
| GET | `/articles/:id` | Get article details | Public |
| POST | `/articles` | Create an article | Editor, Admin |
| PATCH | `/articles/:id` | Update an article | Editor, Admin |
| DELETE | `/articles/:id` | Delete an article | Editor, Admin |

### Categories

| Method | Endpoint | Description | Access |
|---|---|---|---|
| GET | `/categories` | List categories | Public |
| GET | `/categories/:id` | Get category details | Public |
| POST | `/categories` | Create a category | Admin |
| PATCH | `/categories/:id` | Update a category | Admin |
| DELETE | `/categories/:id` | Delete a category | Admin |

### Users

| Method | Endpoint | Description | Access |
|---|---|---|---|
| GET | `/users` | List users | Admin |
| GET | `/users/:id` | Get user details | Admin |
| POST | `/users` | Create a user | Admin |
| PATCH | `/users/:id` | Update a user | Admin |
| DELETE | `/users/:id` | Delete a user | Admin |
| PATCH | `/users/:id/promote` | Promote user to Editor | Admin |
| PATCH | `/users/:id/demote` | Demote user to Member | Admin |

### Dashboard

| Method | Endpoint | Description | Access |
|---|---|---|---|
| GET | `/dashboard` | Get dashboard statistics | Admin |

## API Usage Examples

### 1. Sign Up

```bash
curl -X POST http://localhost:5001/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

Newly registered users receive the `Member` role by default.

### 2. Sign In

```bash
curl -X POST http://localhost:5001/api/v1/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

A successful login returns a JWT access token.

### 3. Get Articles

```bash
curl http://localhost:5001/api/v1/articles
```

### 4. Get Categories

```bash
curl http://localhost:5001/api/v1/categories
```

### 5. Get User Profile

Protected endpoints require a JWT access token.

```bash
curl http://localhost:5001/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

Replace `YOUR_ACCESS_TOKEN` with the token received from the sign-in endpoint.

### 6. Create an Article

Article creation requires the `Editor` or `Admin` role.

The request uses `multipart/form-data` because an image is required.

```bash
curl -X POST http://localhost:5001/api/v1/articles \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "title=My First Article" \
  -F "body=Article content" \
  -F "excerpt=Short description" \
  -F "category_id=1" \
  -F "image=@./example.jpg"
```

Make sure the category exists and `example.jpg` is available in your current directory.

## Role-Based Access Control

The application uses **Casbin** to enforce role-based authorization.

Three roles are supported:

- **Member:** Access public resources and manage their own profile.
- **Editor:** Member capabilities plus article management.
- **Admin:** Full access to user, article, category, and dashboard management.

### Permission Matrix

| Feature | Member | Editor | Admin |
|---|---|---|---|
| View Articles | Yes | Yes | Yes |
| View Categories | Yes | Yes | Yes |
| Manage Own Profile | Yes | Yes | Yes |
| Create Articles | No | Yes | Yes |
| Update Articles | No | Yes | Yes |
| Delete Articles | No | Yes | Yes |
| Manage Categories | No | No | Yes |
| Manage Users | No | No | Yes |
| View Dashboard | No | No | Yes |

Public article and category endpoints are accessible without authentication.

Authorization rules are defined in:

- `config/acl_model.conf`
- `config/policy.csv`

Backend authorization is authoritative. Hiding buttons in the frontend does not grant or restrict API permissions.

## Running with Docker

The project includes a multi-stage Dockerfile that builds the Go application and packages it into a lightweight Alpine Linux image.

### 1. Build the Docker Image

```bash
docker build -t go-restful-api:latest .
```

### 2. Prepare Environment Variables

Make sure `.env` exists in the project directory.

When PostgreSQL runs on your macOS host, use `host.docker.internal` as the database hostname inside the container.

For example:

```dotenv
DATABASE_CONNECTION="host=host.docker.internal user=postgres password=your_password dbname=go_restful_api port=5432 sslmode=disable"
```

> `localhost` inside a Docker container refers to the container itself, not your Mac.

### 3. Run the Container

```bash
docker run -d \
  --name go-restful-api-test \
  -p 5002:5001 \
  --env-file .env \
  -e HOST=http://localhost:5002 \
  -e PORT=5001 \
  -e DATABASE_CONNECTION="host=host.docker.internal user=postgres password=your_password dbname=go_restful_api port=5432 sslmode=disable" \
  -v go-restful-api-uploads:/app/uploads \
  go-restful-api:latest
```

Replace the PostgreSQL username and password with your own values.

Port mapping:

| Location | Port |
|---|---|
| Host machine | `5002` |
| Docker container | `5001` |

Open:

- Web UI: http://localhost:5002/ui/
- API: http://localhost:5002/api/v1

The Docker image is named `go-restful-api:latest`, while the running container is named `go-restful-api-test`.

### 4. Check Container Status

```bash
docker ps
```

View logs:

```bash
docker logs go-restful-api-test
```

### 5. Stop and Start the Container

Stop:

```bash
docker stop go-restful-api-test
```

Start an existing container:

```bash
docker start go-restful-api-test
```

### 6. Remove the Container

```bash
docker stop go-restful-api-test
docker rm go-restful-api-test
```

Removing the container does not automatically remove the named upload volume.

## Deploying to Railway

The application can be deployed to Railway using the included Dockerfile.

### 1. Create a Railway Project

Visit https://railway.com and create a new project.

Add a PostgreSQL service.

### 2. Deploy the Go Application

Create a new service connected to this GitHub repository.

Railway can build the application using the existing Dockerfile.

### 3. Configure Environment Variables

Set the following variables on the Go application service:

| Variable | Value |
|---|---|
| `DATABASE_CONNECTION` | Railway PostgreSQL connection URL |
| `SECRET_KEY` | Secure random secret |
| `GIN_MODE` | `release` |
| `HOST` | Public HTTPS URL of the deployed application |
| `PORT` | Railway-provided port, or `5001` if configured explicitly |

**Important:** The backend reads `DATABASE_CONNECTION`, not `DATABASE_URL`.

If Railway provides a `DATABASE_URL`, reference its value using the variable name `DATABASE_CONNECTION` on the Go service.

Do not configure production PostgreSQL with `localhost`.

### 4. Configure Persistent Storage

The application stores uploaded files in:

```text
./uploads/articles/
./uploads/users/
```

The Docker image uses `/app` as its working directory.

For persistent file storage, attach a Railway volume mounted at:

```text
/app/uploads
```

Without persistent storage, uploaded images may be lost when the container is replaced.

### 5. Configure Public Networking

Generate a public domain for the Go application.

Set `HOST` to the public HTTPS origin without a trailing slash.

Example:

```dotenv
HOST=https://your-app.up.railway.app
```

Ensure the public networking configuration routes traffic to the port on which Gin listens.

### 6. Verify Deployment

After deployment, check:

```text
https://your-app.up.railway.app/ui/
```

And:

```text
https://your-app.up.railway.app/api/v1/categories
```

Database migrations run automatically when the application starts.

### Important: Development Seed Data

Do not enable `seed.Load()` in production.

The development seed logic can drop database tables and insert demonstration accounts with known credentials.

Use a separate secure bootstrap process to create the initial production administrator.

## Security Considerations

This repository is primarily a learning and demonstration project.

Before using it as a public production application, review and strengthen the following areas.

### Authentication

- JWT authentication uses Bearer tokens.
- The Web UI stores the access token in `sessionStorage`.
- Tokens stored in browser-accessible storage may be exposed through XSS vulnerabilities.
- Signing out of the Web UI removes the local token but does not revoke an already issued token.
- Consider secure HttpOnly cookies, CSRF protection, token revocation, and session management for production use.

### Authorization

- Casbin enforces permissions on protected backend routes.
- Frontend role-based controls are for user experience only.
- Never rely on frontend checks as the sole authorization mechanism.

### File Uploads

- Validate file types, sizes, and filenames on the backend.
- Apply upload size limits.
- Do not trust browser-side file restrictions alone.
- Use persistent storage for deployed applications.
- Consider object storage such as Amazon S3 or Cloudflare R2 for scalable production environments.

### Database

- Never commit database credentials.
- Use secure database connections where appropriate.
- Configure backups.
- Test migrations before deploying schema changes.
- Never run destructive development seed operations in production.

### Environment Variables

- Keep `.env` out of version control.
- Generate a strong random `SECRET_KEY`.
- Use separate credentials for development and production.
- Rotate secrets when necessary.

## Troubleshooting

### Database Connection Failed

Check that:

1. PostgreSQL is running.
2. The database exists.
3. `DATABASE_CONNECTION` is correct.
4. The configured database user has the required permissions.

When running the API inside Docker on macOS, use `host.docker.internal` to connect to PostgreSQL running on the host.

### Docker: No Such Container

If you see:

```text
Error response from daemon: No such container
```

List all containers:

```bash
docker ps -a
```

If the container does not exist, create it using `docker run`.

Use `docker start` only for containers that have already been created.

### Docker: .env File Not Found

If Docker reports:

```text
docker: open .env: no such file or directory
```

Make sure you are running the command from the directory containing `.env`.

For example:

```bash
cd /path/to/go-restful-api
ls -la .env
```

Alternatively, provide an absolute path with `--env-file`.

### Web UI Does Not Load

Make sure the Go server is running and access the UI through Gin:

```text
http://localhost:5001/ui/
```

Do not open `web/index.html` directly using a `file://` URL or an IDE Live Preview server.

### Port Already in Use

If port `5001` is occupied, configure a different port using the `PORT` environment variable.

When using Docker, you can also change the host-side port mapping.

For example:

```bash
-p 5002:5001
```

This exposes container port `5001` through host port `5002`.

## Development Notes

- The application uses Gin for HTTP routing.
- GORM manages database access.
- Gormigrate handles schema migrations.
- JWT middleware authenticates protected requests.
- Casbin authorizes requests according to user roles.
- Uploaded images are stored on the filesystem.
- PostgreSQL stores image URLs rather than image file contents.
- The Web UI is served by the same Gin application as the API.

## License

No license is currently specified in this repository.

Add a `LICENSE` file if you intend to distribute the project under an open-source license.
