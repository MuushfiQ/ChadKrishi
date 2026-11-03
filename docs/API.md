# API overview

Prefix `/api/v1`; session cookie required except register/login/health. Write requests from browser must match FRONTEND_ORIGIN. JSON errors: `{error:{message}}`.

| Method / route | Purpose |
|---|---|
| POST /auth/register, /auth/login, /auth/logout | Account/session |
| GET /auth/me; PUT /profile | Private profile, including read-only isAdmin |
| PUT /profile/photo | `{photoPath:"/api/media/owner-file.png"}`; empty string removes, returns updated private user |
| GET /weather, /recommendations | Existing weather and source-backed recommendations |
| GET /crops | Catalog for manual plant entry (includes verification status) |
| GET, POST /plants; PUT /plants/:id | Owned plants; archive/restore in PUT body |
| GET /tasks?date=YYYY-MM-DD; PUT /tasks/:id | Dhaka checklist and `{completed:true/false}` |
| POST /uploads | Multipart `image`, valid JPEG/PNG ≤5 MB, ≤20 MP |
| GET /api/media/:name (outside v1) | Owner, member-visible post image or saved profile picture |
| GET /posts | Filters type, status, area, author, q, before; 20 items/cursor |
| POST /posts; PUT, DELETE /posts/:id | Caption/image community or sale; ownership enforced |
| PUT /posts/:id/reaction | `{reaction:"like|love|angry|sad"}`; empty removes |
| GET, POST /posts/:id/comments; DELETE /comments/:id | Comment feed and owned deletion |
| GET /people/:id | Safe public name, general area, bio, profilePhoto, postCount, isSelf |
| POST /posts/:id/buy | Reserve whole lot; returns order ID, 409 if unavailable |
| GET /orders?before= | Current member's purchases/sales, newest 20, cursor |
| PUT /orders/:id | `{action:"paid",method:"cash|bkash|other",reference:"optional"}` or received/not_received/cancel |
| GET /admin/listings?before= | Admin-only sold/archive inventory |
| PUT /admin/listings/:id/archive | Admin-only, sold items only |
| GET, POST /conversations | Private inbox; POST `{userId:123}` opens/reuses thread |
| GET /conversations/:id/messages?before= | Last 50, descending; older cursor |
| POST /conversations/:id/messages | `{content:"text"}`, max 2000 characters |
| PUT /conversations/:id/read | `{through:messageId}` marks displayed incoming messages read |

Order state rules: waiting_payment + buyer paid → buyer_paid; buyer_paid + seller received → completed/Sold; buyer_paid + seller not_received → waiting_payment; waiting_payment + participant cancel → cancelled/Available. All other transitions return 409; nonparticipants receive 404. No API accepts a client-supplied buyer/seller ID or amount for purchases.

The legacy direct sale-status route returns 409: payment confirmation is required. Legacy like remains supported, while new UI uses the reaction endpoint. General feed hides archived posts. Listing editing is blocked while reserved/completed; deletion is blocked if any order exists. Sessions/ownership/admin checks happen on the backend.


Marketplace UI defaults to All listings (no status filter). Reserved and sold listings remain visible until an admin archives the sold listing. Explicit Available/Reserved/Sold filters narrow the results by choice. A reserved listing cannot be bought again. Refresh posts/history to see another member's latest actions.

`PUT /profile/photo` only accepts the current account's existing local upload. Missing/null photoPath is rejected; `""` removes it. The endpoint uses the existing profile_photo column and works without completed garden setup. Current saved profile pictures may be fetched by other signed-in members; private plant photos and orphan uploads retain their previous restrictions. Posts/comments carry authorPhoto; conversations carry profilePhoto. Public profiles never include email or precise coordinates. User profile timelines paginate through `GET /posts?author=<id>`; archived listings are excluded.


## Notifications, custom tasks and live updates (2026-10-22)

Migration: apply `008_notifications.sql` after 001–007. It adds the durable notifications table; custom tasks reuse care_tasks with a random custom_ rule key.

| Method / route | Contract |
|---|---|
| GET /notifications?before= | Current recipient only; notifications, nextCursor, unreadCount and latestId |
| PUT /notifications/:id/read | Acknowledge own notification; others return 404 |
| PUT /notifications/read-all | `{through:latestId}`; never marks later arrivals read |
| GET /orders?id= | Optional exact order filter, still restricted to buyer/seller; used by notification links |
| POST /tasks | `{plantId,title,reason,date}`; owned active plant, title 1–160, note 0–300, Dhaka date within one year and not before planting |
| DELETE /tasks/:id | Delete own custom task only; system tasks and others' tasks return 404 |

GET /tasks accepts dates within one year before/after today. Returned tasks include plantId, plantName, cropName and custom. Future days show scheduled custom tasks only; automatic tasks are generated on their day. Today's/future tasks for archived plants are hidden; past history remains available. Custom tasks are one-time entries; completion uses the existing PUT endpoint.

New purchase: notify seller. Buyer Paid: notify seller. Seller received/not_received: notify buyer. Cancellation by either participant: notify the other. Message: notify the other conversation participant. These writes share the order/message transaction; notification failure rolls back the action. Reading displayed messages also acknowledges their notifications through the displayed message ID. Opening the bell alone does not acknowledge unread items. Notification read status and chat read receipts are separate except for this chat-read synchronization. Existing events from before 008 are not backfilled.

Frontend: notification polling 5 seconds, messages/inbox 4 seconds, orders 5 seconds, community/market/profile timelines 10 seconds. Feed pagination uses IntersectionObserver plus a manual fallback. Polling pauses for hidden tabs and resumes on visibility/focus/online. It updates React data, never reloads the document or clears drafts. Loaded list pages refresh together; stale filter responses are discarded. Chat catches up across message pages after returning from a hidden/offline tab.

These are persistent in-app notifications, not OS push notifications. A closed browser receives no popup; unread activity appears on the next signed-in visit. Polling depends on browser scheduling and connectivity. Publishing a new listing does not broadcast a notification to every member; it becomes visible through the live feed.


## Crop research update (2026-10-31)

Apply migration 009 or run `go run . -upgrade-crops` before starting this version. This command is repeatable, preserves legacy crop IDs and existing user data, and updates the research JSON. It does not need a weather API request.

| Method / route | Contract |
|---|---|
| GET /garden-setup | Current user's planned pot: depthCm, volumeL, drainage, support, method; unknowns are null |
| PUT /garden-setup | Same fields; method seed or seedling; positive depth 0.01–1000 cm, volume 0.001–1000 L, or null. Values normalize to 2/3 decimal places. User ID always comes from session |
| GET /recommendations | Returns location, date, month (Asia/Dhaka), weather, setup, sources, cropCatalog (all 10) and crops (candidate/conditional subset). Needs complete profile, otherwise 422 |

Crop statuses: candidate, conditional, needs_setup, needs_evidence, out_of_season, not_fit. Missing user input has priority over a conditional unknown crop requirement. Verified seed months are never substituted for missing transplant months. Preferred sunlight maximum is not a hard exclusion. Seasonal guidance is month-level; the season/cultivar note must be checked for exact dates. Location selects weather; this release has no invented district-to-crop calendar mapping.

GET /crops and GET /plants now prefer the current evidence notes/status when available; legacy custom crops remain accessible. GET /tasks adds one crop-observation reminder per active plant/day from the new care notes; completion is retained on repeated requests. This is general observation guidance, not plant-health measurement or prescribed treatment.

`crop_evidence` stores research rows linked to existing crops.id. `garden_setup` stores an independently owned planned container per user. Existing plants retain their own planting date and growth stage.
