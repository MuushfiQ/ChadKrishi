# Render + Aiven deployment

This deployment uses one Render web service. The service builds React, serves the
React app with SPA fallback, and runs the Go API on the same HTTPS origin.
Existing local Vite/XAMPP development remains unchanged.

## Aiven database

The migrations use `defaultdb`, which matches the Aiven database shown in the
connection details. Migration `001_initial.sql` also contains
`CREATE DATABASE IF NOT EXISTS chadkrishi`; it does not switch away from
`defaultdb`, so no database reset is required.

Before applying anything, run the following in MySQL Workbench:

```sql
SELECT DATABASE();
SELECT VERSION();
SHOW TABLES;
SELECT TABLE_NAME, ENGINE
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'defaultdb'
ORDER BY TABLE_NAME;
```

If `SHOW TABLES` is empty, run these files in order in Workbench:

1. `migrations/001_initial.sql`
2. `migrations/002_seed_crops.sql`
3. `migrations/003_auth_profile.sql`
4. `migrations/004_verified_crop_catalog.sql`
5. `migrations/005_expand_verified_container_crops.sql`
6. `migrations/006_fix_bengali_crop_names.sql`
7. `migrations/007_social_commerce.sql`
8. `migrations/008_notifications.sql`
9. `migrations/009_crop_research.sql`

Do not rerun `003`–`007` if those changes already exist. For an existing
database, inspect the table and column state first and apply only the missing
migration after a backup. The backend `-check-db` command is also non-destructive.

## Render environment variables

Create a Render Web Service from this repository using `render.yaml`, or set:

```text
APP_ENV=production
PORT=10000
SESSION_COOKIE_SECURE=true
FRONTEND_ORIGIN=https://YOUR-RENDER-SERVICE.onrender.com
DATABASE_DSN=USERNAME:PASSWORD@tcp(HOST:PORT)/defaultdb?parseTime=true&charset=utf8mb4&tls=aiven
DATABASE_TLS_SERVER_NAME=YOUR-AIVEN-HOST
DATABASE_TLS_CA_PEM=<the Aiven CA certificate, stored only in Render>
WEATHER_CACHE_MINUTES=30
```

Replace `DATABASE_DSN` with the Aiven connection string using the private
username, password, host, port, and `defaultdb` database. URL-escape special
characters in the password. Do not commit this value.

`DATABASE_TLS_CA_PEM` must contain the CA certificate downloaded from Aiven.
Certificate verification is enabled; never replace it with `tls-skip-verify`.
The code also supports `DATABASE_TLS_CA_FILE` for local testing.

## Local Aiven connection test

Keep the local `.env` private and set:

```text
APP_ENV=development
DATABASE_DSN=USERNAME:PASSWORD@tcp(HOST:PORT)/defaultdb?parseTime=true&charset=utf8mb4&tls=aiven
DATABASE_TLS_SERVER_NAME=YOUR-AIVEN-HOST
DATABASE_TLS_CA_FILE=C:\private\aiven-ca.pem
```

Replace the placeholders with your private Aiven values, then run:

```powershell
cd backend
go run . -check-db
```

Your local XAMPP setup can continue using its existing `.env` by removing the
three Aiven TLS variables and restoring the local `DATABASE_DSN`.

## Render free storage

Render Free has an ephemeral filesystem. Uploaded profile, plant, and post
photos can disappear after a restart or redeploy. The database remains external
in Aiven, but image storage is not durable. For production, move uploads to
private object storage (for example S3-compatible storage) and keep signed or
authenticated media access; do not make private plant photos public. Until that
integration is implemented, treat Render uploads as demo-only.

## Manual Bangla steps

1. Aiven থেকে CA certificate download করুন; password বা certificate chat-এ পাঠাবেন না।
2. Workbench-এ আগে `SHOW TABLES;` চালিয়ে database empty কি না দেখুন।
3. Empty হলে উপরকার 001–009 files ধারাবাহিকভাবে চালান; existing table থাকলে blindly rerun করবেন না।
4. Render-এ `render.yaml` দিয়ে service তৈরি করুন এবং environment variables privately দিন।
5. Deploy-এর পর `/api/v1/health` খুলে `status=ok` এবং `database=ok` দেখুন।
