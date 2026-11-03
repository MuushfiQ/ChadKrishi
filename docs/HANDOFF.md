> Historical handoff snapshot. The current implementation and revised scope are in progress.md and IMPLEMENTATION_BN.md. Orders, manual payment acknowledgement and private chat now supersede the original exclusions.

# ChadKrishi technical handoff

Verified against the current workspace on 2026-10-22. This is a university
course-project gardening community for Bangladesh, not professional
agricultural advice.

## 1. Project overview

### Original requirements and MVP scope

The original first release covers:

- Account registration, login, logout, server-side sessions, and an editable
  city-level gardening profile.
- Explainable crop guidance using month, gardening space, and user-provided
  direct sunlight.
- Plants with crop, nickname, planting information, growth stage, photo,
  notes, and archive behavior.
- Idempotent daily care checklists.
- Community text/image posts, post types, comments, likes, pagination, and
  public profiles.
- Marketplace listings implemented as sale posts in the same feed.

Explicit boundaries: no exact home addresses, payments, carts, delivery, chat,
closed-app notifications, plant-health detection, fixed water/fertilizer
dosing, or enterprise infrastructure. Images are intended to be local files
with generated names and validation.

### Stack

- Frontend: React 19, Vite 6, React Router 7, Axios, Bootstrap 5, Vitest and
  Testing Library dependencies.
- Backend: Go 1.26, Gin 1.11, `go-sql-driver/mysql` 1.9, bcrypt from
  `golang.org/x/crypto`.
- Database: MySQL-compatible MariaDB (local XAMPP MariaDB 10.4.32 was used),
  InnoDB, `utf8mb4`.
- Provider: Open-Meteo Forecast API; source catalog references BAMIS and
  University of Maryland Extension.

### Relevant structure

- [src/App.jsx](../src/App.jsx): auth, app shell, profile, dashboard, crop
  card, and route guards.
- [src/api.js](../src/api.js): Axios client, `/api/v1` base path, credentials,
  API error extraction.
- [src/main.jsx](../src/main.jsx): React root, Bootstrap, router.
- [src/styles.css](../src/styles.css): shared visual system and responsive UI.
- [backend/main.go](../backend/main.go): server, handlers, auth, profile,
  recommendation rules, weather provider/cache.
- [backend/auth_test.go](../backend/auth_test.go): validation, suitability,
  WMO description, Bengali JSON tests.
- [backend/weather_test.go](../backend/weather_test.go): provider parsing and
  cache tests.
- [migrations/](../migrations/): ordered MySQL schema/seed/repair scripts.
- [docs/requirements.md](requirements.md), [docs/schema.md](schema.md), and
  [docs/progress.md](progress.md): requirements, data model, and progress.

## 2. Current status

| Feature | Status | Actual implementation |
|---|---|---|
| Registration | Complete | `POST /api/v1/auth/register`, bcrypt, session cookie, React form |
| Login | Complete | `POST /api/v1/auth/login`, bcrypt verification, session rotation |
| Logout | Complete | Session revocation, expired cookie, frontend state clear and redirect |
| Sessions/protected routes | Complete | Hashed token in `sessions`, middleware, `/auth/me`, private React routes |
| Profile | Complete | City, coordinates, rooftop/balcony, sunlight, rain exposure, bio; persisted in `users` |
| Weather | Complete/partial | Live Open-Meteo current + 7-day forecast, cache, stale/unavailable states; no care-task integration |
| Crop catalog | Partial | Four `source-backed` crops plus three preserved `source-pending` demo rows |
| Recommendations | Complete for current scope | Authenticated, month/space/sunlight/source-status matching; no container-size input in profile |
| My Plants/growth stage | Pending | Tables exist, but no API, React route, or working UI |
| Daily care tasks | Pending | Tables exist, but no generator, API, dashboard list, or completion UI |
| Community | Pending | Tables exist, but no feed/composer/comments/likes APIs or working UI |
| Marketplace | Pending | Sale columns/tables are prepared, but no listing/filter/API/UI |
| Plant-to-sale post | Pending/prepared | `posts.plant_id` and sale fields exist; no plant picker or sale-post flow |

Placeholder UI is limited to disabled “Soon”, “Coming next”, and “Add a
plant/Share an update” elements in the dashboard/sidebar. They do not call
working APIs.

Implemented file references:

- Auth/profile/dashboard UI: [src/App.jsx](../src/App.jsx)
- Auth/profile/recommendation/weather backend:
  [backend/main.go](../backend/main.go)
- Auth regression tests: [backend/auth_test.go](../backend/auth_test.go)
- Weather tests: [backend/weather_test.go](../backend/weather_test.go)
- Schema prepared for pending modules: [migrations/001_initial.sql](../migrations/001_initial.sql)

## 3. Backend/API

Base URL is `/api/v1`.

| Method | Endpoint | Auth | Purpose |
|---|---|---|---|
| GET | `/api/v1/health` | No | API and database health |
| POST | `/api/v1/auth/register` | No | Create account and start session |
| POST | `/api/v1/auth/login` | No | Verify password and start session |
| POST | `/api/v1/auth/logout` | Yes | Revoke current session |
| GET | `/api/v1/auth/me` | Yes | Return current user/profile |
| PUT | `/api/v1/profile` | Yes | Validate and update profile |
| GET | `/api/v1/recommendations` | Yes | Return location/month/weather/crop matches |
| GET | `/api/v1/weather` | Yes | Return weather for saved supported city |

### Request shapes

Registration:

```json
{"name":"Gardener","email":"user@example.com","password":"at-least-8-chars"}
```

Login:

```json
{"email":"user@example.com","password":"at-least-8-chars"}
```

Profile:

```json
{
  "name":"Gardener",
  "district":"Dhaka",
  "latitude":23.8103,
  "longitude":90.4125,
  "gardeningSpace":"rooftop",
  "sunlightHours":8,
  "rainExposure":true,
  "bio":""
}
```

Successful auth/profile responses contain a `user` object. Recommendations
contain `location`, `month`, `weather`, and `crops`. Crop results include
source metadata, sunlight/space requirements, recommendation reason, unknowns,
and source URLs. Weather contains separate `current` and `forecast` values,
provider attribution, valid time, fetched time, and `stale` when applicable.

### Errors and authorization

Errors use:

```json
{"error":{"message":"human-readable message"}}
```

Typical statuses: `400` invalid input/JSON, `401` missing/invalid/expired
session, `403` disallowed write origin, `409` duplicate email, `422`
incomplete/unsupported profile, `500` database/catalog error, and `503`
database not configured.

Sessions use a random 32-byte token in an HttpOnly, SameSite=Lax cookie named
`chadkrishi_session`. Only its SHA-256 hash is stored. Sessions expire after
seven days and logout sets `revoked_at` plus a deletion cookie. User identity
always comes from the session, not a client user ID. Writes are origin-guarded
against `FRONTEND_ORIGIN`.

Validation covers name/email/password, supported Bangladesh city coordinates,
rooftop/balcony, sunlight `0..24`, required rain exposure, and bio length.
Supported city mapping is in [backend/main.go](../backend/main.go).

### Environment variables

See [.env.example](../.env.example). Sanitized names/defaults:

- `APP_ENV`
- `PORT` (default `8080`)
- `FRONTEND_ORIGIN` (default `http://localhost:5173`)
- `DATABASE_DSN` (local example uses XAMPP root with empty password)
- `SESSION_COOKIE_SECURE` (`false` for local HTTP)
- `WEATHER_CACHE_MINUTES`
- Optional `OPEN_METEO_BASE_URL` for provider tests

Never copy `.env` or `backend/.env` into a handoff/archive.

## 4. Database

### Tables and relationships

- `users`: identity, bcrypt hash, city/general location, gardening space,
  sunlight, rain exposure, bio, timestamps.
- `sessions`: user FK, SHA-256 token hash, expiry, revocation, active index.
- `crops`: names/descriptions, JSON planting months/spaces, sunlight bounds,
  container/care guidance, source and verification metadata.
- `user_plants`: user/crop FKs, nickname, planting date, seed/seedling method,
  growth stage enum, photo/note, archive timestamp.
- `care_tasks`: plant FK, rule key, title/reason, due/completed timestamps,
  unique `(plant_id, rule_key, due_date)` for idempotency.
- `posts`: user/optional plant FKs, update/question/blog/sale type, content,
  image, BDT price, quantity, contact, available/sold state.
- `comments`: post/user FKs and content.
- `post_likes`: composite primary key `(post_id, user_id)`.

Foreign keys use cascade behavior for owned children; a post's optional plant
reference is set NULL if the plant is removed. Plants are intended to be
archived, not physically deleted.

### Migration order/status

Apply in order:

1. `001_initial.sql`: database and all base tables.
2. `002_seed_crops.sql`: three original demo crop rows.
3. `003_auth_profile.sql`: nullable profile fields and session cleanup.
4. `004_verified_crop_catalog.sql`: crop metadata and Cucumber source row.
5. `005_expand_verified_container_crops.sql`: Brinjal, Garlic, Pepper.
6. `006_fix_bengali_crop_names.sql`: repairs double-encoded Bengali values.

All six were applied to the local XAMPP MariaDB during development without
deleting accounts or profiles. Migration 003 intentionally deletes only
expired/revoked sessions. Do not rerun destructive database resets or replace
existing user data.

### Catalog

Current source-backed rows: Cucumber, Brinjal, Garlic, Pepper. Original
Lal shak, Chili, and Tomato rows remain `source-pending` and are excluded by
the current recommendation filter. BAMIS Dhaka calendar PDFs provide local
season windows; UMD Extension provides general sunlight/container guidance.
Exact variety/container performance remains caveated or unknown where noted.

## 5. Frontend

### Routes and state

React Router routes are defined in [src/App.jsx](../src/App.jsx):

- `/auth`
- `/dashboard`
- `/profile`

The top-level `App` loads `/auth/me` once, keeps the current user in React
state, redirects private routes when unauthenticated, and clears state on
logout/session expiry. `Dashboard` fetches recommendations on mount.
`ProfilePage` owns form state and updates the current user after save.
There is no Redux/Zustand; state is local React `useState/useEffect`.

The shared shell includes desktop sidebar, top bar, mobile bottom navigation,
profile link, and sign-out controls. [src/api.js](../src/api.js) uses Axios
with `/api/v1`, credentials, and a shared error-message helper.

### Design conventions and states

Forest green, cream, white cards, earthy accents, rounded corners, subtle
shadows, responsive grids, visible focus rings, inline validation, loading
spinners, error retry states, incomplete-profile empty state, unavailable/stale
weather state, and no-match recommendation state are implemented.

Known UI limitations:

- Community, plants, marketplace, and care-task UI is not implemented; disabled
  labels are intentional.
- No frontend test files currently exist.
- The current recommendation sunlight comparison is a strict inclusive range,
  even where a source range may be guidance rather than a biological hard
  maximum.
- Local development must use the configured frontend origin
  (`http://localhost:5173`) for write requests; using `127.0.0.1` can trigger
  the backend origin guard.
- Bengali label corruption was fixed in migration 006. Future MySQL CLI
  imports must use `--default-character-set=utf8mb4`.

## 6. Pending features and proposed next work

The following behavior is required by the original requirements but is not
implemented:

1. **My Plants:** authenticated CRUD for a user's plants, crop selection,
   nickname, planting date, propagation method, growth stage, photo upload,
   notes, archive/unarchive, ownership checks, and responsive list/detail UI.
2. **Daily care:** generate idempotent tasks from active plants and crop rules,
   use Asia/Dhaka due dates, show completed history, and keep routine tasks
   available if weather is unavailable. Weather should remain advisory.
3. **Community:** authenticated text/image post creation, update/question/blog
   types, paginated feed, public general-area profile display, comments and
   likes with ownership/rate/validation checks.
4. **Marketplace:** sale posts in the same feed with BDT price, quantity,
   general area, optional published contact, available/sold state, and filters.
5. **Plant-to-sale post:** choose one owned active plant, create a `posts` row
   with `type='sale'` and `plant_id`, expose only safe plant/crop summary
   publicly, and allow the owner to mark it sold. The schema is prepared, but
   the API, validation, plant picker, and UI do not exist.

Assumptions/proposals, not current agreements: pagination size, exact image
limits, crop-specific care rule catalog, whether archived plants may be sold,
and whether a sale post can be edited after publication must be decided before
implementation. Do not add payments, delivery, live chat, or notifications.

## 7. Run and verify

### Requirements

- Windows PowerShell
- Node.js 24+
- Go 1.26+
- XAMPP MySQL/MariaDB on `127.0.0.1:3306`

### Setup

```powershell
cd "E:\7th SemisteR\ChadKrishi"
Copy-Item .env.example .env
```

Edit only local `.env` values. Do not commit it. For default XAMPP root with
no password, the sanitized DSN is:

```text
root:@tcp(127.0.0.1:3306)/chadkrishi?parseTime=true&charset=utf8mb4
```

Start MySQL from XAMPP Control Panel, then apply migrations with the XAMPP
client:

```powershell
$mysql = "C:\xampp\mysql\bin\mysql.exe"
Get-Content migrations\001_initial.sql | & $mysql --default-character-set=utf8mb4 --protocol=tcp -h 127.0.0.1 -u root
Get-Content migrations\002_seed_crops.sql | & $mysql --default-character-set=utf8mb4 --protocol=tcp -h 127.0.0.1 -u root
Get-Content migrations\003_auth_profile.sql | & $mysql --default-character-set=utf8mb4 --protocol=tcp -h 127.0.0.1 -u root
Get-Content migrations\004_verified_crop_catalog.sql | & $mysql --default-character-set=utf8mb4 --protocol=tcp -h 127.0.0.1 -u root
Get-Content migrations\005_expand_verified_container_crops.sql | & $mysql --default-character-set=utf8mb4 --protocol=tcp -h 127.0.0.1 -u root
Get-Content migrations\006_fix_bengali_crop_names.sql | & $mysql --default-character-set=utf8mb4 --protocol=tcp -h 127.0.0.1 -u root
```

Start the API in one PowerShell:

```powershell
cd "E:\7th SemisteR\ChadKrishi\backend"
go mod download
go run .
```

Start the frontend in a second PowerShell:

```powershell
cd "E:\7th SemisteR\ChadKrishi"
npm install
npm run dev
```

Open `http://localhost:5173`. API is `http://localhost:8080`; Vite proxies
`/api` to that port. Health check:

```powershell
curl.exe http://localhost:8080/api/v1/health
```

### Checks and evidence

Passed during this handoff preparation:

```powershell
cd backend
go test ./...
go vet ./...
cd ..
npm run build
```

Previously passed: XAMPP health, migrations 001–006, real register/profile/
logout/login flows, recommendation API, live Dhaka Open-Meteo request, stale
weather tests, Bengali UTF-8 byte verification, WMO mapping tests, and
desktop/mobile browser checks for the implemented dashboard/auth/profile scope.

Not run/available: frontend automated tests (no test files), Playwright
regression suite, pending-feature API tests, and full automated browser tests
for expired sessions/back navigation. The manual browser check used a
previously retained local account.

Known blockers/reproduction:

- Starting Vite at `127.0.0.1:5173` while `FRONTEND_ORIGIN` is
  `http://localhost:5173` causes write requests to return `403 request origin
  is not allowed`; use `localhost` or change the local origin consistently.
- A fresh account without district/space/sunlight gets `422 complete your
  gardening profile first` from recommendations by design.
- If Open-Meteo is unavailable and no cache exists, the dashboard shows
  `Weather unavailable`; this is an intentional honest state.
- `node_modules`, build output, local env files, logs, and uploads are not
  source deliverables.

## 8. Recommended implementation milestones

1. Add migration/API tests and shared ownership helpers for `user_plants`;
   implement plant CRUD, archive, photo validation, and UI.
2. Add a small explicit care-rule catalog and idempotent task generation;
   implement task list/complete/history.
3. Add community posts, comments, likes, pagination, and public-safe profile
   responses.
4. Add marketplace sale posts and filters using the existing `posts` schema.
5. Add the owned-plant-to-sale-post flow, sold transition, and privacy tests.
6. Add focused frontend tests and Playwright journeys for auth, profile,
   weather unavailable/stale, plant ownership, and sale privacy.

Keep each milestone course-project sized and preserve the current no-payments,
no-delivery, no-chat boundaries.
