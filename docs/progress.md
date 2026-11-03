# ChadKrishi progress — 22 October 2026

## Current release

Functional course-project implementation of the requested gardening, community, manual marketplace and text-chat flows. Source implementation is complete for the scope below. A final desktop/mobile visual browser check remains unverified because this execution environment blocks the Unix socket required to launch Chrome.

## Implemented

- Existing register/login, HttpOnly server sessions, profile, sign-out and private routes retained. API distinguishes unavailable database from expired session. Private profile now includes ID and server-controlled admin flag.
- Existing Open-Meteo current/seven-day weather, stale/unavailable states, source attribution and source-backed recommendations retained. Forecast rain probability field corrected in UI.
- My Plants: add/edit crop, date, seed/seedling, observed stage, photo, private notes, archive/restore; recommendation-to-plant and plant-to-sale actions.
- Daily checklist: Dhaka dates; idempotent task identity; completion and past-date history; general observations and stage prompts. Fresh weather may add rain/wind precautions; no sensor claims or fake weather.
- Community: update/question/blog/sale posts; JPEG/PNG photo plus caption; own edit/delete; comments with own deletion; like/love/angry/sad, one reaction per user/post; search/filter/cursor pagination; safe member profiles.
- Shared feed/marketplace listings: general area, total lot price, quantity, optional chosen public contact, photo and linked crop. Private plant notes/account email are not automatically published.
- Buy reserves the whole lot atomically. Concurrent buyers cannot both reserve. Buyer declares external payment; seller confirms receipt to mark Sold. Waiting cancellation and seller Not received return flow implemented.
- Private buy/sell history with immutable title/amount/quantity snapshots; actor/time transition audit in `order_events`; participant-only state changes. Ordered posts cannot be deleted; active/completed listings cannot be edited.
- Admin-only archive of sold listings. Archived listings leave public member feeds; both parties retain history. Admin provisioned only by DB owner, not signup/profile fields.
- Private two-person text conversations, persisted messages, inbox unread counts, read receipts, older-message pagination; 4-second polling while open. Start chat from profile/post/order. Third-party accounts cannot read/write a thread.
- Botanical SVG rooftop/sprout logo, favicon, redesigned feed/reaction controls, listing/history/payment cards and responsive chat layout.
- Local image validation: input ≤5 MB, decoded ≤20 MP, JPEG/PNG only, random owner-prefixed name, metadata stripped by re-encoding. Private/orphan uploads cannot be fetched by unrelated members.
- Bengali setup/upgrade guide and learning/data-flow explanation in IMPLEMENTATION_BN.md. API/schema documentation updated.

## Validation actually run in this release

- `npm test`: **9 frontend component tests passed**. Covers plant request payload, care completion, honest errors, reaction switching, buyer/seller payment UI, private chat send/read rendering, admin archive UI.
- `npm run build`: passed, 103 modules; production output compiled.
- `go test ./...`: **13 top-level tests passed**, including existing auth/recommendation/UTF-8/weather tests, plant/sale/date/stage rules and order state permissions.
- `go vet ./...`: passed.
- All migrations 001–007: applied successfully to a new isolated MariaDB 10.11 instance. No user database was accessed or changed here. 007 adds fields/tables without deleting earlier rows; an actual upgrade on the user's XAMPP remains to be performed.
- `python scripts/api_smoke.py --test-only`: passed against the actual Go server and isolated MariaDB. Validated register/profile/relogin/logout; plant ownership, task idempotency/completion/archive; upload validation/media access; Unicode captions; reaction replacement/comment ownership; safe public profile; concurrent Buy (one 201, one 409); seller/buyer/nonparticipant permissions; paid/retry/receipt/Sold; cancellation; retained history; unique/private chat, read receipts, 50-message cursor pagination; admin archive/unauthorized admin denial; expired sessions.
- Integration weather provider deliberately unavailable to verify routine checklist resilience. Existing weather parser/cache tests passed; no fresh live provider claim is made for this release.

## Not verified / limitations

- Playwright desktop/mobile journey script supplied in `scripts/browser_smoke.cjs`, but **did not execute its journey here**: Chrome startup failed with `socket() failed: Operation not permitted` in this sandbox. Do not interpret build/component tests as visual browser verification. Run the supplied script or Bengali manual checklist locally.
- No live XAMPP/Windows test on the user's PC, all-city live weather test, real cash/bKash transfer or delivery test was performed.
- Payments are manual declarations/acknowledgements, not gateway verification. Whole-lot purchases only; no partial stock, automatic timeout, escrow, refunds or delivery automation.
- Chat is text-only polling, not WebSocket/push, file sharing, group calling or end-to-end encryption. Inbox refresh is explicit; open conversation updates periodically. History refresh is explicit after another participant acts.
- Care is generic observation with simple fresh-weather precautions, not independently validated crop-specific dosing or disease advice. Growth stage is user-reported.
- Existing catalog remains limited (four source-backed crops plus pending entries for manual records); no extra agronomic sources were verified during this implementation.
- Public-launch hardening remains: abuse rate controls, reporting/blocking/content moderation beyond sold-listing archive, backup/restore operations, HTTPS deployment, upload cleanup, notification delivery and operational monitoring.

## Upgrade action

Back up source/database; preserve `.env` and `backend/uploads`; apply **007 once** if 001–006 already exist; `npm ci`; backend `go run .`; frontend `npm run dev`. Read IMPLEMENTATION_BN.md for exact steps and admin setup. The new source ZIP intentionally excludes real credentials, uploads, database dumps, node_modules and build output.


## Social/profile update — 2026-10-22 (Bangladesh)

Implemented:
- Fixed listing disappearance: Marketplace default is now All listings rather than Available. Buy retains atomic reservation and prevents another purchase. Reserved listings remain visible; only buyer Paid followed by seller receipt marks Sold. Sold remains visible until administrator archive. Added explicit Refresh posts.
- Removed post-bottom Message actions. Author name/photo opens /people/:id with avatar, bio, area, post count and that author's history (cursor pagination). Message is in the profile header; own profile offers Edit instead of self-chat. Post reactions/comments and purchase history chat remain functional.
- Added account-owned JPEG/PNG profile picture upload/replacement/removal in My Profile, independent of garden setup. Existing profile_photo column reused; no new migration. Avatar appears in sidebar/topbar, author links, comments, member profile, inbox and selected chat header.
- Added authenticated PUT /profile/photo and explicit media authorization for published profile pictures. A member cannot attach another member's upload. Private plant/orphan media are not made broadly accessible.
- Added README_UPDATE_BN.md with exact files to copy into the existing project and Windows run/manual verification steps. Patch ZIP includes changed/new files only; no credentials, uploads, account exports or migrations.

Verification for this update:
- npm test: 17 frontend component tests passed, including 8 additional cases for default reserved-listing visibility/filter, author navigation, public profile/chat, own profile, upload/save/remove, invalid upload and failed photo save.
- npm run build: passed.
- go test ./... and go vet ./...: passed.
- Expanded scripts/api_smoke.py passed against the actual updated Go server and a newly initialized isolated MariaDB 10.11 database with migrations 001–007 only. Covers profile-photo ownership/persistence/removal/public visibility/incomplete profile; author post filter/count/photos; reserved visibility both before and after buyer Paid; sold visibility until archive; concurrent purchase guard; cancellation; admin archive; history; private chat; auth logout/relogin and expired sessions.
- No user's XAMPP database was accessed here. No new agronomic research/weather-provider verification in this update. Desktop/mobile visual browser verification remains unperformed; component/API tests do not establish screenshot fidelity.

Remaining behavior: feeds and history refresh explicitly; open chats poll. An Available-only filter intentionally excludes reserved listings. Member history shows existing nonarchived posts, with Load more for older posts. Historical archived sales remain in private buy/sell history.


## Notification, care and live-feed update — 2026-10-22

Completed:
- Durable recipient-only notifications with badge, new-event toast, cursor history, individual read and bounded Mark all read. Order/message writes and notification inserts are atomic.
- Both directions covered: purchase/paid notify seller; receipt/payment retry notify buyer; cancellation notifies counterparty; messages notify recipient. Links open the exact private order or chat; old orders remain reachable using a participant-checked ID filter.
- Bell polls on all signed-in pages. Inbox/open messages/orders/feed update without full reload. Hidden tabs pause; focus/visibility/network restoration resumes polling. Message drafts remain intact. Chat catches up across multiple pages when required.
- Community, marketplace and profile timelines load older posts on scroll, with button fallback and retry handling. Background refresh preserves loaded history and checks older rows too. Out-of-order responses from previous filters cannot replace the current feed.
- Custom tasks: own active plant, chosen date, title and optional note, persistent completion and deletion of custom tasks. One-time scheduling up to one year ahead; no recurring automation implied.
- Care grouped by plant ID (same crop or same nickname stays separate), with progress per plant, custom labels, per-plant Add task and dashboard grouping.
- Added 008_notifications.sql and cumulative patch README_LIVE_CARE_BN.md. Existing auth, profile photos, sale reservation/payment/history, admin archive and weather behavior retained.

Verification:
- 27 frontend tests passed: prior 17 plus notification polling/read navigation, grouped plants, custom creation, older-page live refresh/deduplication, stale-filter protection, infinite scroll, connection recovery, live inbox/chat with draft preservation, automatic seller payment controls.
- Go tests and go vet passed; frontend build passed.
- Expanded real API integration suite passed against updated Go server + isolated MariaDB 10.11 using migrations 001–008. 008 reapplied successfully. Includes custom task ownership/validation/future/archive; order notification recipient and role checks; message notification pagination/read receipts; mark-all does not consume later arrivals; exact-order privacy; concurrent purchase; notification storage failure rolls back message and purchase; archive/history; existing profile photo, authentication, plant and community checks.
- No user XAMPP/database touched. Desktop/mobile visual browser testing and closed-browser OS push were not performed/implemented. No new live weather or agronomic verification in this update.

Upgrade: use cumulative patch source, preserve .env and backend/uploads, import only 008 into the existing chadkrishi database (001–007 already installed), restart backend and frontend. No dependency additions. The earlier 'no new migration' note applied only to the profile-picture release; this release needs 008.

Behavior limits: in-app polling (roughly 4–10 seconds with a visible connected tab), durable unread history while away, no OS push when browser closed; historical pre-migration events not backfilled; no blanket notification to every user for new listings; custom tasks are one-time and not notification reminders.


## 2026-10-27 — full source recovery release

Reported: buy, cancel reservation, and send message return generic database operation failed after live-care upgrade. Exact Windows DB error was not supplied. All three share transactional notification writes. Reproduced missing notifications-table failure on disposable MariaDB and verified repair. This is a confirmed failure mode, not a claim that every possible local DB issue has been diagnosed.

Changes:
- Full source package (not a patch), preserving original module and all frontend/backend files.
- Explicit `-repair-notifications` command creates missing migration-008 table without deleting existing data; incompatible existing schemas are reported, not guessed/rebuilt.
- `-check-db` and startup preflight verify connection, commerce columns, InnoDB engines, and required order/listing enum states.
- Missing table/column runtime errors return actionable HTTP 503, with private data kept out of responses/logs; other DB failures log MySQL code and route.
- Readiness includes commerce schema and returns 503 on unhealthy state.
- HTTP server timeouts and production secure-cookie/HTTPS-origin configuration validation.
- Full Bengali recovery/run instructions and Docker Compose/Caddy deployment configuration, persistent DB/uploads volumes; no secrets or personal uploads in deliverable.

Validation this release:
- Frontend: 27 tests passed; production Vite build passed (107 modules).
- Backend: go test ./... and go vet ./... passed; Go executable built.
- Disposable MariaDB 10.11: 001-007 import, missing 008 startup rejection, repair, repeat repair, check-db, existing incompatible-column rejection and recovery all passed.
- Real authenticated integration suite passed: profile/photos, community/reactions/comments, plants/custom tasks, simultaneous buyers, cancellation, payment transitions, sold/archive/history, private chat, durable notifications/read state, ownership, logout/relogin and session expiry.
- Forced missing notification table during runtime: message, purchase, and cancel all failed without partial commits; cancellation succeeded after table restoration.
- compose.yaml parsed as YAML; Docker daemon unavailable, so container build/runtime/TLS/public staging deployment NOT tested.

Remaining validation before public release:
- User's Windows database/schema and browser end-to-end journey.
- Container image builds, real domain HTTPS, backup/restore and persistent-volume permissions on target host.
- Internet abuse/rate-limit controls, monitoring and security review. No claim of universally bug-free software or audited production readiness.
- Crop catalog breadth and closed-browser push remain limited as documented in previous milestones.


## 2026-10-31 — crop research update completed

Completed software scope:
- Ten research crop entries from the supplied report, with independently reviewed core fields separated from report-only/unknown fields. Five core-reviewed entries can be planning candidates subject to season/setup; five stay evidence-incomplete.
- Dashboard all-crop catalog, Bangla/English search, status filters, season/variety conditions, readable notes and field source links. No new crop photographs supplied; symbols and optional local image support with fallback.
- Per-user nullable planned-container setup: depth, volume, drainage, support, seed/transplant method. Ownership enforced; stored precision and response agree.
- Month-level seasonal assessment, distinct transplant windows, minimum sunlight without incorrect preferred-maximum exclusion, pot checks, missing-setup priority over unknown-requirement conditions.
- Existing real weather integration retained and separated from seasonal crop guidance; crop catalog works when weather is unavailable. No fake district specificity.
- My Plants and care notes use the new evidence catalog while preserving stable crop IDs; daily crop observations remain idempotent with completion retained.
- Repeatable `-upgrade-crops` command and equivalent migration 009; existing records and custom legacy crop notes preserved. Full standalone source package, not a patch.
- Bengali run/upgrade/data-flow/source-editing/photo instructions; API and deployment docs updated for 009.

Validation: Go tests/vet/build passed; 32 frontend tests passed; production build passed (108 modules). Disposable MariaDB/real API crop and social-commerce regression tests passed, including repeated upgrades, existing-data retention, transactional notification rollback, payment/cancel/history, messaging, photos and session checks. See CROP_RELEASE_CHECKS.md for scope and reproduction.

Remaining: browser visual/end-to-end verification in user's environment, Docker/public deployment validation and production security/operations; missing agronomic evidence (district/cultivar/transplant/pot/harvest details), original crop photos, closed-browser push. These are not claimed complete. The earlier submission PDF describes the recovery baseline, not this later crop release.
