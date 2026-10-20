# POS Backend - Tracking Money

## Prerequisites

- **Go 1.25.0** (defined in `go.mod`)
- **MySQL 8.0.30** (Laragon or compatible)
- **Air** for hot-reload (or `go build`/`go run`)
- **Migrate** CLI for DB migrations

## 1. Environment Setup


```dotenv
PORT=8080
DB_DSN=root@tcp(127.0.0.1:3306)/pos?parseTime=true&loc=Asia%2FJakarta
JWT_SECRET=FadhilGanteng123
```

> **Note:** `.env` is gitignored (`.gitignore:21`). Do not commit this file.

## 2. Database Migrations

The project uses Goose-style SQL migrations under `migrations/`. Run with the `migrate` CLI:

```bash
# Install migrate if needed (go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
migrate -database "root@tcp(127.0.0.1:3306)/pos?parseTime=true&loc=Asia%2FJakarta" -path migrations up
# or down
migrate -database "root@tcp(127.0.0.1:3306)/pos?parseTime=true&loc=Asia%2FJakarta" -path migrations down
```

**Migration files** (up/down pairs) in `migrations/`:
- `000001_create_users_table.up/down.sql`
- `000002_create_branch_table.up/down.sql`
- `000003_create_mm_user_branch_table.up/down.sql`
- `000004_services.up/down.sql`
- `000005_create_mm_branch_service_table.up/down.sql`
- `000006_add_slug_to_branch_table.up/down.sql`
- `000007_create_menus.up/down.sql`
- `000008_create_mm_service_menu.up/down.sql`
- `000009_add_is_active_to_service.up/down.sql`
- `000010_create_payment_type.up/down.sql`
- `000011_create_mm_service_payment.up/down.sql`
- `000012_create_ticket.up/down.sql`
- `000013_create_ticket_item.up/down.sql`
- `000014_create_ticket_payment.up/down.sql`

## 3. Running Tests

```bash
go test ./...
```

All packages pass (`ok pos-backend/internal/handler/httptest`).

## 4. Running the Project

### Option A: Air (hot-reload, recommended)

```bash
air
```

- Starts `http://localhost:8080` (default `PORT` from `.env`)
- Auto-restarts on source changes
- Logs appear in `tmp/server.log` / `tmp/server.err.log`

### Option B: Manual build & run

```bash
go build -o tmp/main.exe ./cmd/api
./tmp/main.exe
```

## 5. API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST /api/v1/auth/login` | login | public | Authenticate (uses bcrypt + JWT HS256 with secret `FadhilGanteng123`) |
| `POST /api/v1/tickets` | create-ticket | public | Create a new ticket |
| `GET /api/v1/tickets?branch_id=1` | list-tickets | Bearer JWT | List recent tickets (3‑day window) |
| `GET /api/v1/preload/{id}` | preload | private | Pre‑load branch/menu/service data |

**Key filter behavior**: `GET /api/v1/tickets?branch_id=1` applies a 3‑day window from today (D‑2 00:00 → D+1 00:00 Asia/Jakarta). After the timezone fix (`loc=Asia/Jakarta` in DSN), dates align correctly with MySQL session timezone.

## 6. Important Notes from Development

- **Timezone fix**: The root cause of "3‑day filter missing today's data" was a mismatch between Go's `time.Time` conversion (default `UTC` when DSN lacks `loc`) and MySQL's `session time_zone = SYSTEM = +07:00` (SE Asia Standard Time). Added `loc=Asia%2FJakarta` to DSN so driver sends wall-clock Jakarta times that MySQL interprets consistently.
- **Created_at serialization**: With the fix, `created_at` now returns `2026-09-28T21:28:10+07:00` instead of the erroneous `2026-09-28T21:28:10Z`.
- **Session timezone log**: At startup the app logs `DB session time_zone=SYSTEM, NOW()=..., UTC_TIMESTAMP()=..., offset=7h0m0s` — no warning means everything is aligned.
- **Test data**: Tickets inserted via API appear immediately in `GET /api/v1/tickets` with correct timestamps.

## 7. Folder Structure Overview

```
pos-backend/
├─ cmd/api/           # Entry point (main.go)
├─ internal/
│  ├─ domain/        # Data structures (Ticket, User, etc.)
│  ├─ service/       # Business logic (TicketService, AuthService)
│  ├─ repository/    # MySQL implementations
│  ├─ handler/       # HTTP handlers
│  └─ pkg/           # JWT utility, hasher, utils
├─ migrations/        # SQL up/down files
├─ tmp/               # Built binary + logs
├─ .env               # (gitignored) DB/DN + JWT secret
├─ go.mod             # Go 1.25.0 + dependencies
└─ .air.toml          # Air hot‑reload config
```

---
**End of README content.**
