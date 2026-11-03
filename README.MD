# Whimpy Intern API (Go + MySQL)

A small REST API written in Go for managing intern accounts (registration, profile lookup/update, login/logout sessions, and deletion), backed by a MySQL database. Routing is handled by [`gorilla/mux`](https://github.com/gorilla/mux), database access by [`go-sql-driver/mysql`](https://github.com/go-sql-driver/mysql), and password hashing by [`golang.org/x/crypto/bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt).

---

## Table of Contents

- [API Introduction](#api-introduction)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Database Setup](#database-setup)
- [Configuration](#configuration)
- [Running Locally](#running-locally)
- [API Endpoints](#api-endpoints)
- [Testing with Postman](#testing-with-postman)
- [Troubleshooting & Known Limitations](#troubleshooting--known-limitations)

---

## API Introduction

The API exposes a set of HTTP endpoints that operate on the `wp_intern_list` and `sessions` tables of a MySQL database named `whimpy_intern_db`.

**Base URL:** `http://localhost:2105`

**Design notes:**

- All request/response bodies are JSON (`Content-Type: application/json`).
- Intern IDs are generated randomly by the server on registration (32-bit value, with a duplicate-key retry loop).
- Login issues an HttpOnly `sessionID` cookie valid for 24 hours; the session row is stored in the `sessions` table and removed on logout.
- Passwords are expected to be stored as bcrypt hashes (`password_hash` column) and are verified with `bcrypt.CompareHashAndPassword`.

### Available features

| Feature | Description |
| --- | --- |
| Register | Create a new intern account with a server-generated ID |
| List | Fetch all intern accounts |
| Read | Fetch a single intern account by ID |
| Update | Update username, full name and e-mail of an intern |
| Login / Logout | Create and destroy a cookie-based session |
| Delete | Remove an intern account |

> Mentor and Human Resource endpoints exist in the source but are currently commented out / not registered on the router.

---

## Project Structure

```
whimpy-mysql-api/
├── app.go        # Entire application: routes, handlers, DB access
├── go.mod        # Module definition (module golang-mysql-api)
├── go.sum        # Dependency checksums
└── README.MD     # This file
```

---

## Prerequisites

| Tool | Version | Notes |
| --- | --- | --- |
| Go | **1.26.1+ (declared in `go.mod`)** | https://go.dev/dl/ |
| MySQL Server | 8.x recommended | Running on `localhost` |
| MySQL client | any (`mysql` CLI, MySQL Workbench, phpMyAdmin…) | For creating the schema |
| Postman (optional) | v10+ | For API testing — see [Testing with Postman](#testing-with-postman) |

Verify your toolchain:

```bash
go version
mysql --version
```

---

## Installation

```bash
# 1. Clone the repository
git clone https://github.com/adam-ghafara/whimpy-mysql-api.git
cd whimpy-mysql-api

# 2. Download/verify the Go dependencies
go mod download

# 3. Build the binary
go build -o whimsy-api.exe .

# (optional) Run it directly without building
go run app.go
```

Dependencies resolved from `go.mod`:

- `github.com/gorilla/mux v1.8.1` — HTTP router
- `github.com/go-sql-driver/mysql v1.10.1` — MySQL driver
- `golang.org/x/crypto v0.58.0` — bcrypt password hashing

---

## Database Setup

The server expects a MySQL instance listening on **port `8111`** (not the default `3306`) with a database called `whimpy_intern_db`.

### 1. Create the database and user

```sql
CREATE DATABASE IF NOT EXISTS whimpy_intern_db
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- If your MySQL user is not 'root', create one that matches the DSN in app.go
CREATE USER 'root'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON whimpy_intern_db.* TO 'root'@'localhost';
FLUSH PRIVILEGES;
```

### 2. Create the tables

```sql
USE whimpy_intern_db;

CREATE TABLE IF NOT EXISTS wp_intern_list (
  intern_id       INT          NOT NULL PRIMARY KEY,
  intern_username VARCHAR(100) NOT NULL UNIQUE,
  intern_fullname VARCHAR(150) NOT NULL,
  intern_email    VARCHAR(150) NOT NULL UNIQUE,
  intern_password VARCHAR(255) NOT NULL,          -- plain/legacy password column
  password_hash   VARCHAR(255) NOT NULL,          -- bcrypt hash, used by /intern/login
  intern_phone    BIGINT       NULL,
  intern_edu      VARCHAR(150) NULL,
  intern_major    VARCHAR(150) NULL,
  department_id   INT          NULL
);

CREATE TABLE IF NOT EXISTS sessions (
  id         VARCHAR(64) NOT NULL PRIMARY KEY,   -- 32-char hex session id
  user_id    INT         NOT NULL,
  expires_at DATETIME    NOT NULL,
  INDEX idx_sessions_user (user_id)
);
```

> **Note:** Column names in the queries are written in mixed case (`internID`, `internId`, `intern_id`, `password_hash`…). MySQL column names are case-insensitive on most platforms, so a single set of snake_case columns works for all queries.

### 3. Make sure MySQL listens on port 8111

In `my.ini` / `my.cnf`:

```ini
[mysqld]
port = 8111
```

Then restart the MySQL service.

---

## Configuration

The connection string is hard-coded near the top of `main()` in `app.go`:

```go
dsn := "root:your_password@tcp(localhost:8111)/whimpy_intern_db?parseTime=true"
```

Edit it to match your local environment:

| DSN part | Meaning |
| --- | --- |
| `root` | MySQL username |
| `your_password` | MySQL password (replace it) |
| `localhost:8111` | MySQL host and port |
| `whimpy_intern_db` | Database name |
| `parseTime=true` | Parse `DATETIME`/`TIMESTAMP` into `time.Time` |

The HTTP server listens on **port `2105`**:

```go
if err := http.ListenAndServe(":2105", r); err != nil { ... }
```

---

## Running Locally

### 1. Start MySQL

Make sure the server is running on `localhost:8111` and the `whimpy_intern_db` schema exists.

### 2. Start the API

```bash
go run app.go
```

### 3. Smoke-test the server

Open a browser or use `curl`:

```d
http://localhost:2105/
```

```bash
curl http://localhost:2105/
```

You should receive the ASCII banner back as plain JSON-ish text — that confirms the router and server are up.

---

## API Endpoints

All endpoints are relative to `http://localhost:2105`.

| Method | Path | Handler | Description |
| --- | --- | --- | --- |
| `GET` | `/` | `BaseHandler` | Health check / banner (basic API test) |
| `POST` | `/intern/register` | `AddInternUserHandler` | Create a new intern (server generates `intern_id`) |
| `GET` | `/intern` | `GetInternUsersHandler` | List all interns |
| `GET` | `/getuser/{id}` | `GetInternUserByIDHandler` | Get one intern by ID |
| `UPDATE`* | `/intern/update/{id}` | `UpdateInternHandler` | Update username / fullname / email |
| `PUT` | `/intern/update-info/{id}` | `AdvancedInternUserUpdateHandler` | Advanced profile update (currently a stub) |
| `POST` | `/intern/login/{id}` | `SigninInternUserHandler` | Log in, sets `sessionID` cookie |
| `POST` | `/intern/logout/{id}` | `SignOutInternUserHandler` | Log out, clears the cookie |
| `DELETE` | `/intern/delete/{id}` | `DeleteInternUserHandler` | Delete an intern |

\* `UPDATE` is not a standard HTTP verb; see [Troubleshooting](#troubleshooting--known-limitations).

### Request / response examples

**POST `/intern/register`** — body:

```json
{
  "intern_username": "johndoe",
  "intern_fullname": "John Doe",
  "intern_email": "john.doe@example.com",
  "intern_password": "Secret123!",
  "intern_phone": 81234567890
}
```

Response `201 Created`:

```json
{
  "intern_id": 1847293641,
  "intern_username": "johndoe",
  "intern_fullname": "John Doe",
  "intern_email": "john.doe@example.com",
  "intern_password": "Secret123!",
  "intern_phone": 81234567890
}
```

**GET `/getuser/{id}`** → `200 OK` with the intern object, or `404 User Not Found`.

**POST `/intern/login/{id}`** — body:

```json
{ "username": "johndoe", "password": "Secret123!" }
```

Response `200 OK` + `Set-Cookie: sessionID=…; Path=/; HttpOnly`:

```json
{ "message": "logged in", "user_id": 1847293641 }
```

**POST `/intern/logout/{id}`** → `204 No Content` and the cookie is expired.

**DELETE `/intern/delete/{id}`** → `204 No Content`, or `404 User Not Found`.

### Status codes used

| Code | Meaning |
| --- | --- |
| `200 OK` | Successful read / login |
| `201 Created` | Intern registered |
| `204 No Content` | Update or delete succeeded (no body) |
| `400 Bad Request` | Malformed JSON body |
| `401 Unauthorized` | Invalid credentials |
| `404 Not Found` | Intern ID does not exist / nothing changed |
| `409 Conflict` | Duplicate unique value |
| `500 Internal Server Error` | Database or server error |

---

## Testing with Postman

### Option A — Run the requests manually

1. Start the server (`go run app.go`) and make sure MySQL is up.
2. Open Postman → **Create a request**.
3. Set the method and URL, e.g. `GET` `http://localhost:2105/intern`.
4. Press **Send**.

#### 1. Health check

| Field | Value |
| --- | --- |
| Method | `GET` |
| URL | `http://localhost:2105/` |

Expected: the ASCII banner, HTTP `200`.

#### 2. Register an intern

| Field | Value |
| --- | --- |
| Method | `POST` |
| URL | `http://localhost:2105/intern/register` |
| Body → raw → JSON | see the JSON example above |

Copy the returned `intern_id` — you'll need it for the other requests.

#### 3. List interns / read one intern

```
GET http://localhost:2105/intern
GET http://localhost:2105/getuser/<intern_id>
```

#### 4. Update an intern

```
POST http://localhost:2105/intern/update/<intern_id>
```

Body:

```json
{
  "intern_username": "johndoe",
  "intern_fullname": "John Doe (updated)",
  "intern_email": "john.new@example.com"
}
```

#### 5. Login and keep the session

1. `POST http://localhost:2105/intern/login/<intern_id>` with body:

   ```json
   { "username": "johndoe", "password": "Secret123!" }
   ```

2. In Postman, open **Settings → General** and make sure **Automatically follow redirects** and cookie handling are enabled (Postman stores cookies per domain by default via the *Cookies* icon in the status bar).
3. Subsequent requests to the same host will automatically send the `sessionID` cookie.

#### 6. Logout

```
POST http://localhost:2105/intern/logout/<intern_id>
```

Expected: `204 No Content`, and the `sessionID` cookie is cleared.

#### 7. Delete an intern

```
DELETE http://localhost:2105/intern/delete/<intern_id>
```

### Option B — Import a ready-made collection

Create a Postman collection with these pre-defined requests (Postman → **New → Collection → Import** a JSON file you save as `whimpy-intern-api.postman_collection.json`):

```json
{
  "info": {
    "name": "Whimpy Intern API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    { "key": "baseUrl", "value": "http://localhost:2105" },
    { "key": "intern_id", "value": "0" }
  ],
  "item": [
    {
      "name": "Health check",
      "request": {
        "method": "GET",
        "url": { "raw": "{{baseUrl}}/" }
      }
    },
    {
      "name": "Register intern",
      "event": [
        {
          "listen": "test",
          "script": {
            "exec": [
              "const body = pm.response.json();",
              "if (body.intern_id) pm.collectionVariables.set('intern_id', body.intern_id);",
              "pm.test('Status is 201', () => pm.response.to.have.status(201));"
            ]
          }
        }
      ],
      "request": {
        "method": "POST",
        "header": [{ "key": "Content-Type", "value": "application/json" }],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"intern_username\": \"johndoe\",\n  \"intern_fullname\": \"John Doe\",\n  \"intern_email\": \"john.doe@example.com\",\n  \"intern_password\": \"Secret123!\"\n}"
        },
        "url": { "raw": "{{baseUrl}}/intern/register" }
      }
    },
    {
      "name": "List interns",
      "request": { "method": "GET", "url": { "raw": "{{baseUrl}}/intern" } }
    },
    {
      "name": "Get intern by ID",
      "request": {
        "method": "GET",
        "url": { "raw": "{{baseUrl}}/getuser/{{intern_id}}" }
      }
    },
    {
      "name": "Update intern",
      "request": {
        "method": "POST",
        "header": [{ "key": "Content-Type", "value": "application/json" }],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"intern_username\": \"johndoe\",\n  \"intern_fullname\": \"John Doe\",\n  \"intern_email\": \"john.doe@example.com\"\n}"
        },
        "url": { "raw": "{{baseUrl}}/intern/update/{{intern_id}}" }
      }
    },
    {
      "name": "Login",
      "event": [
        {
          "listen": "test",
          "script": {
            "exec": [
              "pm.test('Login successful', () => pm.response.to.have.status(200));",
              "pm.test('sessionID cookie set', () => pm.cookies.has('sessionID'));"
            ]
          }
        }
      ],
      "request": {
        "method": "POST",
        "header": [{ "key": "Content-Type", "value": "application/json" }],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"username\": \"johndoe\",\n  \"password\": \"Secret123!\"\n}"
        },
        "url": { "raw": "{{baseUrl}}/intern/login/{{intern_id}}" }
      }
    },
    {
      "name": "Logout",
      "request": {
        "method": "POST",
        "url": { "raw": "{{baseUrl}}/intern/logout/{{intern_id}}" }
      }
    },
    {
      "name": "Delete intern",
      "request": {
        "method": "DELETE",
        "url": { "raw": "{{baseUrl}}/intern/delete/{{intern_id}}" }
      }
    }
  ]
}
```

Save it as `whimpy-intern-api.postman_collection.json`, then in Postman choose **Import → Files** and select it. The `intern_id` captured by the *Register intern* test script is reused automatically by the following requests.

### Suggested test flow (Collection Runner)

1. **Health check**
2. **Register intern** (captures `intern_id`)
3. **List interns**
4. **Get intern by ID**
5. **Update intern**
6. **Login** (stores the session cookie)
7. **Logout**
8. **Delete intern**

Run it with **Runner → Run** against `http://localhost:2105`.

---

## Troubleshooting & Known Limitations

| Symptom | Cause / fix |
| --- | --- |
| `Error: `dial tcp 127.0.0.1:8111: connect` / connection refused` | MySQL is not running on port **8111** — start it, or change the DSN port in `app.go`. |
| `Error 1045: Access denied for user 'root'` | Wrong username/password in the DSN — update `your_password` in `app.go`. |
| `Error 1146: Table '…wp_intern_list' doesn't exist` | Run the schema in [Database Setup](#database-setup). |
| `go run` fails with a Go version error | Install Go **1.26.1+**, or lower the `go` directive in `go.mod`. |
| `POST /intern/update/{id}` returns `405` | The route is registered for the custom `UPDATE` HTTP verb, which most clients don't send. Use `PUT`/`POST` after changing `Methods("UPDATE")` in `app.go`, or send `UPDATE` explicitly from Postman's raw request builder. |
| `GET /intern` returns `500` | `GetInternUsersHandler` scans rows into non-pointer values — a known bug in `app.go` (line ~176); pass `&u.Username`, `&u.FullName`, etc. |
| Login returns `401` even with a correct password | Login reads `password_hash` (bcrypt). Make sure the value stored in that column is a bcrypt hash of the password, not the plain-text value in `intern_password`. |
| Cookie not sent by Postman | Enable cookie handling in Postman (**Cookies** icon / *Automatically follow cookies*) and send requests to the same host. |

Other notes:

- `AdvancedInternUserUpdateHandler` (`PUT /intern/update-info/{id}`) is an empty stub and returns an empty `200`.
- Mentor and HR routes are commented out in `main()`.
- Sessions expire after 24 hours (`expires` is stored in `sessions.expires_at`); expired rows are not auto-pruned yet.
- `Secure` is set to `false` on the session cookie for local HTTP testing — set it to `true` when serving over HTTPS in production.

---

## License

See the repository owner for licensing information. Created by [github.com/adam-ghafara](https://github.com/adam-ghafara).
