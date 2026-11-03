# ChadKrishi — crop research update (31 October 2026)

Complete React + Go + MySQL/MariaDB source, including existing authentication, social community, marketplace, messaging, notifications and plant care.

**Start here:** [বাংলা run/upgrade guide](START_HERE_BN.md). Existing installations: back up database/uploads, retain your `.env`, run `go run . -upgrade-crops`, then `go run . -repair-notifications`, then `go run . -check-db` inside `backend`. Do not rerun migrations 001–007 on an existing database.

- [Crop features, evidence limits and editing guide](docs/CROP_UPDATE_BN.md)
- [Current release validation](docs/CROP_RELEASE_CHECKS.md)
- [Deployment guide](docs/DEPLOYMENT_BN.md)

The sections below describe the earlier recovery baseline. The current upgrade instructions above take precedence.

# ChadKrishi

React + Go/Gin + MySQL/MariaDB urban gardening course project for Bangladesh.

Includes authentication/profile, weather/recommendations, plants/care, photo community feed with four reactions/comments, marketplace reservations and manual payment confirmation, private buy/sell history, admin archive, and private text chat.

**Start here for this full release:** [START_HERE_BN.md](START_HERE_BN.md).

**বাংলা setup, upgrade, learning guide:** [docs/IMPLEMENTATION_BN.md](docs/IMPLEMENTATION_BN.md).
**Current status / validation:** [docs/progress.md](docs/progress.md).

## Upgrade an existing installation

1. Back up source and database. Merge updated source, preserving your `.env` and `backend/uploads`.
2. If migrations 001–006 are already applied, apply **007_social_commerce.sql once**, followed by **008_notifications.sql** through phpMyAdmin. Do not rerun earlier ALTER migrations.
3. Run `npm ci` at project root. Start the backend with `cd backend` then `go run .`; start frontend in another terminal with `npm run dev` at project root.
4. Open http://localhost:5173. Existing accounts are retained. Refresh after upgrading.

## First installation

Requirements: Node.js 20.19+ (22 LTS also suitable), Go 1.26+, MySQL 8 / MariaDB 10.4+ (XAMPP).

- Copy `.env.example` to `.env` and configure database credentials.
- Apply migrations 001 through 008 in order, once each.
- `npm ci` then `npm run dev`.
- In a second terminal: `cd backend` then `go run .`.
- Health: http://localhost:8080/api/v1/health

No seeded password or automatic admin. Register, then the DB owner may grant admin with `UPDATE users SET is_admin=TRUE WHERE email='your-actual-email@example.com';`. Refresh to see Admin archive.

## Verification

```sh
npm test
npm run build
cd backend
go test ./...
go vet ./...
```

`python scripts/api_smoke.py --test-only` exercises a running isolated test installation, creating test accounts/posts/orders. Never point test scripts at a live community database. Browser checks are documented in progress.md; the supplied Playwright journey is unverified here because Chrome could not start under sandbox socket restrictions. After installing Playwright and Chromium on your PC, run `node scripts/browser_smoke.cjs` after the API smoke test against the same isolated installation.

## Product rules

- A Buy reserves **the entire lot** at its advertised total price. No partial purchase.
- Buyer reports Paid after an external cash/bKash/other payment. Only seller receipt confirmation marks Sold. There is no payment gateway or automatic verification.
- Waiting reservations may be cancelled by either participant. Buyer-paid claims may be returned to waiting by seller if payment was not received. Completed orders cannot be reopened by either member.
- Admin archive hides sold listings while preserving transaction records. Ordered posts cannot be deleted.
- One reaction per member/post: like, love, angry, sad. Photos support JPEG/PNG up to 5 MB and 20 MP, with metadata stripped by re-encoding.
- Private text chat uses 4-second polling while the thread is visible, with pagination and read receipts.
- Generic observation-based care, not sensor readings or agronomic diagnosis. Existing verified crop catalog is limited.

## Deployment boundaries

This is a functional course-project release, not a hardened public marketplace. Before public deployment, configure HTTPS/secure cookies, persistent backed-up upload storage and database, production reverse proxy, rate limits, reporting/blocking/moderation, operational monitoring and recovery, and a payment/dispute policy. No orders are fulfilled or payments moved by this code. Uploaded files persist locally; unused uploads need periodic operator cleanup. Preserve uploads during upgrades.
