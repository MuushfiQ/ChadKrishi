# ChadKrishi database schema

All tables use MySQL `utf8mb4`, InnoDB, foreign keys, and UTC timestamps.
Migrations are versioned and seeds are rerunnable without destructive resets.

## Tables

- `users`: account identity, password hash, profile fields, selected city,
  city-level latitude/longitude, gardening space, sunlight, rain exposure, bio.
- `sessions`: hashed random session token, user relation, expiry, and revocation.
- `crops`: seeded crop catalog, planting months, sunlight range, spaces,
  container/care guidance, source reference, and verification status.
- `user_plants`: a user's crop instance, nickname, planting date, propagation
  method, growth stage, photo, note, and archive timestamp.
- `care_tasks`: generated plant checklist rows. A unique
  `(plant_id, rule_key, due_date)` prevents duplicate generation while keeping
  completed history.
- `posts`: community content. `type = sale` stores sale title, price (DECIMAL),
  quantity, contact method, availability, and optional plant relation.
- `comments`: text replies owned by users and attached to posts.
- `post_likes`: one row per user/post pair, enforced by a unique key.

## Important relationships

`users 1--N user_plants`, `users 1--N posts`, `user_plants 1--N care_tasks`,
`posts 1--N comments`, and `posts N--N users` through `post_likes`.

Plants are archived rather than physically deleted. A sale post keeps its
internal plant reference, but its public response never includes private plant
notes or care history. Deleting a user is outside the first release and will be
specified before it is implemented.


## Social commerce extension — migration 007

- `users.is_admin`: operator-granted role, default false; never accepted from registration/profile JSON.
- `posts.sale_status`: available → reserved → sold. `archived_at` hides a sold listing from public member feeds. Prices are total lot prices.
- `post_likes.reaction`: like/love/angry/sad; existing composite key enforces one member reaction per post.
- `orders`: seller/buyer/listing IDs, immutable title/amount/quantity snapshot, state, reported method/reference, buyer-paid and received times. FK restrictions retain history.
- `order_events`: transition actor, action and timestamp; append-only from API state changes.
- `conversations`: sorted participant pair with a unique key; no duplicate thread for a pair.
- `messages`: participant sender, text, timestamp, read timestamp; indexed by conversation/id for cursor paging.

Reservation and order transitions lock the listing before the order. Orders and listing state update in one transaction. Cancelled reservations release the lot; a buyer-paid claim must be confirmed or returned to waiting by the seller. Completed transactions are not reopened. Archived listings retain orders.


## Crop research extension — migration 009

- `crop_evidence(slug PK, crop_id UNIQUE FK → crops.id, data JSON)`: reviewed evidence and explicit unknowns; preserves original crop IDs and user_plants references.
- `garden_setup(user_id PK/FK → users.id ON DELETE CASCADE, depth_cm DECIMAL(7,2) NULL, volume_l DECIMAL(8,3) NULL, drainage BOOLEAN NULL, support BOOLEAN NULL, method ENUM(seed,seedling))`: one planned container per user, separate from existing plants.
- Both use InnoDB. `-upgrade-crops` creates these tables and upserts the bundled catalog in a transaction; legacy custom crop descriptions are preserved. DDL creation itself follows normal MySQL implicit-commit behavior; a failed data upgrade can be retried.
