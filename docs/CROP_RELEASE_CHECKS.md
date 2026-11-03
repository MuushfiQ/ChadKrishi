# Crop release validation

Release date: 31 October 2026. Research evidence review dates remain 27 October 2026.

## Passed on the final source

- Go 1.26: `go test ./...`, `go vet ./...`, and executable build.
- Frontend: `npm test` — **32 tests in 4 files**; `npm run build` — **108 modules**, successful production bundle.
- Rules: season/light/space, seed vs transplant, unknown requirements vs missing user setup, small pot/no drainage, storage precision boundaries, missing evidence exclusion, 10-entry source mapping integrity.
- React: unknown values remain null, explicit false drainage, chosen propagation method, filtering, failed save feedback, valid local photo and broken-photo fallback, rejecting arbitrary remote crop images; existing social/activity/garden component tests retained.
- Real Go API + isolated MariaDB 10.11.14: migrations 001–007, repeated CLI crop upgrade, SQL 009 reapplication with unchanged stable IDs, notification repair twice, schema check. Existing custom crop description preserved.
- Crop API: all 10 entries, incomplete-evidence labels, null values, pot validation/rounding, ownership and persistence, transplant separation, crop notes in My Plants, one daily crop reminder, completed task preserved on refresh.
- Existing API suite: registration/profile/photo ownership, community captions/reactions/comments, plant ownership, custom tasks, simultaneous buyers, reservation/cancellation, buyer Paid/seller receipt, sold/archive/history, private chat, recipient notifications and pagination, logout/relogin and expiry.
- Forced notification-table failure: message, purchase and cancellation rolled back without partial commits; recovery succeeded after restoration.
- Repeated crop upgrade after user records existed preserved counts of users/plants/posts/orders/messages/tasks and exact saved pot setup.

All integration tests used a newly initialized disposable test database. No user XAMPP database was accessed or changed. Weather was intentionally unavailable in integration tests to verify crop guidance works without a live forecast; existing weather unit tests passed.

## Not verified in this release

- Desktop/mobile visual browser journeys: attempted browser setup could not produce a runnable browser in this environment. Component tests and API tests do not establish visual fidelity.
- Windows/XAMPP user's existing schema or custom changes, all supported cities' live Open-Meteo requests, public hosting, Docker build/runtime, real HTTPS and production security review.
- Five entries still lack core recommendation evidence; cultivar/district/transplant gaps remain explicit. The catalog is not agronomic certification.

## Reproduce locally

Backend: `go test ./...` and `go vet ./...` inside backend.
Frontend: `npm ci`, `npm test`, `npm run build` at project root.

Only on a separate disposable database with 001–009 applied and backend running:

```sh
python scripts/crop_smoke.py --test-only --base http://127.0.0.1:8081/api/v1
python scripts/api_smoke.py --test-only --base http://127.0.0.1:8081/api/v1
```

For admin/session-expiry/rollback branches, api_smoke also accepts TEST_MYSQL_BIN and TEST_DB_PORT. It creates test accounts and deliberately renames the notification table temporarily. Never point it at real user data.

## User acceptance checklist after upgrade

1. Existing account signs in; profile and uploaded images remain.
2. Dashboard shows ten crop entries; search/filter and evidence details work.
3. Save pot details, refresh, confirm persisted values; compare shortlist and All crops.
4. Change starting method to Seedling: unknown transplant windows do not reuse seed seasons.
5. Track a crop, add custom care task, complete it, refresh; each plant has its own section.
6. Two accounts: create listing → Buy → Cancel → Buy → Paid → Payment received. Reserved stays visible; sold follows seller confirmation.
7. Send messages in both directions; verify notifications and persistence.

See START_HERE_BN.md for the exact upgrade order. No promise of universal bug-free behavior or audited production readiness is made.
