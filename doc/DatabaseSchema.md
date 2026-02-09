# Database Schema Reference

This page documents the database schema for the DSC Pull Server, supporting both SQLite and MSSQL backends. It describes the main tables, their fields, and relationships, to help administrators and developers understand the data model.

---

## Supported Databases

- **SQLite**: Lightweight, file-based database for development and small deployments.
- **MSSQL**: Microsoft SQL Server for production and enterprise environments.

Schema files:
- SQLite: `db/schema_sqlite.sql`
- MSSQL: `internal/db/schema.sql` (or `schema_mssql.sql` if present)

---

## Main Tables

### 1. `users`
| Column         | Type           | Description                        |
|---------------|----------------|------------------------------------|
| id            | INTEGER/INT    | Primary key, auto-increment        |
| username      | TEXT/VARCHAR   | Unique username                    |
| password_hash | TEXT/VARCHAR   | Hashed password                    |
| email         | TEXT/VARCHAR   | User email address                 |
| is_admin      | BOOL/INT       | Admin flag (1=true, 0=false)       |
| is_active     | BOOL/INT       | Active flag (1=true, 0=false)      |
| saml_id       | TEXT/VARCHAR   | SAML subject identifier (nullable) |

### 2. `api_token`
| Column      | Type         | Description                         |
|-------------|--------------|-------------------------------------|
| id          | INTEGER/INT  | Primary key, auto-increment         |
| user_id     | INTEGER/INT  | Foreign key to `users.id`           |
| token       | TEXT/VARCHAR | API token value                     |
| expires_at  | DATETIME     | Expiration timestamp                |
| created_at  | DATETIME     | Creation timestamp                  |

### 3. `agent`
| Column         | Type         | Description                        |
|---------------|--------------|------------------------------------|
| id            | INTEGER/INT  | Primary key, auto-increment        |
| agent_id      | TEXT/VARCHAR | Unique agent identifier (GUID)     |
| hostname      | TEXT/VARCHAR | Hostname of the agent              |
| tags          | TEXT/VARCHAR | Comma-separated tags               |
| registered_at | DATETIME     | Registration timestamp             |
| last_seen     | DATETIME     | Last check-in timestamp            |
| is_active     | BOOL/INT     | Active flag                        |

### 4. `configuration_model`
| Column         | Type         | Description                        |
|---------------|--------------|------------------------------------|
| id            | INTEGER/INT  | Primary key, auto-increment        |
| name          | TEXT/VARCHAR | Configuration name                 |
| content       | TEXT         | MOF/DSC configuration content      |
| created_at    | DATETIME     | Creation timestamp                 |
| updated_at    | DATETIME     | Last update timestamp              |

### 5. `dsc_report`
| Column         | Type         | Description                        |
|---------------|--------------|------------------------------------|
| id            | INTEGER/INT  | Primary key, auto-increment        |
| agent_id      | TEXT/VARCHAR | Foreign key to `agent.agent_id`    |
| job_id        | TEXT/VARCHAR | DSC Job identifier                 |
| status        | TEXT/VARCHAR | Report status                      |
| report        | TEXT         | Raw DSC report JSON                |
| created_at    | DATETIME     | Report creation timestamp          |

---

## Relationships

- `api_token.user_id` → `users.id`
- `dsc_report.agent_id` → `agent.agent_id`

---

## Schema Evolution

- **SQLite**: Update `db/schema_sqlite.sql` for schema changes.
- **MSSQL**: Update `internal/db/schema.sql` (or `schema_mssql.sql`).
- Use migration scripts or manual SQL for upgrades.

---

## Example: Table Creation (SQLite)

```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  email TEXT,
  is_admin INTEGER DEFAULT 0,
  is_active INTEGER DEFAULT 1,
  saml_id TEXT
);

CREATE TABLE agent (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  agent_id TEXT NOT NULL UNIQUE,
  hostname TEXT,
  tags TEXT,
  registered_at DATETIME,
  last_seen DATETIME,
  is_active INTEGER DEFAULT 1
);

-- ...other tables omitted for brevity
```

---

> For full schema details, see the SQL files in `db/` and `internal/db/`.
> Always back up your database before applying schema changes.
