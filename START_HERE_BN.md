# ChadKrishi — সম্পূর্ণ crop update ও run guide

এই ZIP একটি সম্পূর্ণ project। আগের live-care ZIP ছিল patch; এই ZIP একাই VS Code-এ খুলতে পারবেন। Node modules install এবং database setup প্রয়োজন। সব তথ্য code-এর ভিতরে থাকে না: accounts, posts, orders, messages MySQL-এ, ছবিগুলো uploads folder-এ থাকে।

## কী সমস্যা পাওয়া গেছে

Buy, Cancel reservation, Send message-এর সঙ্গে notification একই database transaction-এ লেখা হয়। Notification schema missing হলে মূল operation-ও rollback হয়। আপনার screenshot একই failure pattern দেখায়; আপনার Windows database সরাসরি পরীক্ষা করা হয়নি, তাই আপনার ক্ষেত্রে exact SQL error এখনো নিশ্চিত নয়।

এই release-এ:
- Server startup-এ commerce tables/columns, transactional storage engine ও order/listing states পরীক্ষা হয়। Schema ভুল থাকলে server কারণসহ বন্ধ হয়।
- `-repair-notifications` missing notifications table তৈরি করে। আগের accounts/posts/messages মুছে না; থাকা incompatible table নিজে বদলায় না।
- `-check-db` connection/schema পরীক্ষা করে, data বদলায় না।
- Runtime-এ missing table/column হলে setup error দেখায়। Backend terminal-এ route ও MySQL error number থাকে, private message/password/SQL values log হয় না।
- Notification ব্যর্থ হলেও order বা message অর্ধেক save হবে না: transaction rollback অক্ষত আছে।
- Health endpoint schema/connection অসুস্থ হলে HTTP 503 দেয়।

## আপনার চলমান project থেকে এই release চালানো

1. আগের frontend/backend terminal-এ Ctrl+C চাপুন। পুরোনো project folder backup রাখুন। phpMyAdmin থেকে `chadkrishi` database Export করে SQL backup রাখুন।
2. ZIP extract করুন। যে folder-এ `package.json` এবং `backend` আছে, VS Code-এ সেই folder খুলুন। পুরোনো code-এর ওপর অন্ধভাবে folder replace করার দরকার নেই।
3. পুরোনো project-এর `.env` একই অবস্থানে নতুন project-এ copy করুন। আগে `backend/.env` ব্যবহার করলে সেখানেই রাখুন। দুই জায়গায় ভিন্ন `.env` রাখবেন না: backend-এরটি আগে load হয়। Password chat-এ পাঠাবেন না।
4. পুরোনো `backend/uploads`-এর সব ছবি নতুন `backend/uploads`-এ copy করুন। Custom `UPLOAD_DIR` থাকলে সেই folder-ই ব্যবহার করুন। Database একই `chadkrishi` থাকলে accounts, posts, history থাকবে। নতুন ZIP-এ আপনার ব্যক্তিগত ছবি/database নেই।
5. XAMPP-এ MySQL Start করুন। Backend-এর জন্য Apache প্রয়োজন নেই; phpMyAdmin ব্যবহার করলে Apache-ও Start করুন।
6. VS Code terminal-এ project root থেকে:

```powershell
cd backend
go run . -upgrade-crops
go run . -repair-notifications
go run . -check-db
go run .
```

প্রতিটি command সফল হওয়ার পর পরেরটি চালাবেন। Crop upgrade-এর expected message: `Crop catalog upgraded; existing plants preserved.`

Notifications repair-এর expected terminal message:

```text
Notification schema ready; existing data preserved.
```

`-check-db`-এর expected message:

```text
Database connection and commerce schema OK
```

`-upgrade-crops` migration 009-এর কাজ করে: research catalog এবং container setup tables যোগ করে। এটি আবার চালানো যায়; পুরোনো crop IDs ও user data রাখা হয়। Command সফল হলে 009 আলাদাভাবে Import করতে হবে না। আগে crop upgrade, পরে notifications repair চালান।

`-repair-notifications` এখানে missing migration 008-এর কাজ করে। আগে 008 import হয়ে থাকলেও valid table রেখে দেয়। 001–007 আবার import করবেন না। Existing table-এর schema mismatch দেখালে থামুন এবং শুধু error text দিন; DROP TABLE বা database delete করবেন না।

7. আলাদা terminal খুলুন, project root-এ:

```powershell
npm ci
npm run dev
```

Browser: http://localhost:5173
Health: http://localhost:8080/api/v1/health

পুরোনো browser tab refresh করুন। একই localhost cookie থাকার কারণে আগের account login থাকতে পারে। অন্য browser/incognito-তে দ্বিতীয় account ব্যবহার করে পরীক্ষা করুন।

## একেবারে নতুন installation

প্রয়োজন: Go 1.26+, Node.js 22, npm, MySQL/MariaDB (যেমন XAMPP)। `.env.example` copy করে `.env` করুন, credentials ঠিক করুন। phpMyAdmin-এ migrations/001 থেকে 009 প্রতিটি একবার ক্রমানুসারে Import করুন। তারপর `go run . -check-db`, `go run .`, এবং অন্য terminal-এ `npm ci`, `npm run dev`। নতুন database-এ পুরোনো users নিজে থেকে আসবে না।

## হাতে যাচাই

1. Account A থেকে listing; B থেকে Buy। Listing reserved অবস্থায় marketplace-এ থাকবে।
2. Seller ও buyer history-তে order থাকবে; seller notification পাবেন।
3. Waiting order buyer অথবা seller Cancel করলে listing available হবে; counterparty notification পাবেন।
4. আবার Buy → buyer Paid → seller Payment received: তখনই listing sold। Admin archive করলে feed থেকে সরে যাবে, history থাকবে।
5. দুই account পরস্পর message পাঠান; receiver inbox/notification update হবে। Refresh-এর পরও message থাকবে।
6. নিজের plant-এ custom task যোগ, complete, delete; অন্য account task দেখতে/বদলাতে পারবে না।

## Error হলে বোঝার উপায়

- Connection refused: MySQL চলছে কি না এবং DSN-এর port যাচাই করুন।
- Access denied: DSN credentials/DB user permissions যাচাই করুন।
- Unknown database: ভুল database name অথবা নতুন database এখনো import হয়নি।
- notifications missing: `go run . -repair-notifications`।
- অন্য table/column/state missing: পুরোনো schema আলাদা; error অনুযায়ী migration review দরকার। শুধু 008 দিয়ে পুরোনো 007 schema মেরামত হয় না।
- Runtime MySQL code 1146 = missing table; 1054 = missing column; 1452 = foreign-key reference mismatch। Backend terminal-এর code দিন, password বা `.env` নয়।

## গুরুত্বপূর্ণ সীমা

এটি পরীক্ষিত সংশোধিত source; “সব অবস্থায় bug-free” নিশ্চয়তা নয়। আপনার স্থানীয় database-এর exact contents/schema, Windows browser journey ও public server deployment এখানে সরাসরি পরীক্ষা হয়নি। আপনার পুরোনো project-এ আলাদাভাবে করা local edits এই source-এ নাও থাকতে পারে; backup রেখে compare করুন। Closed browser push notification নেই—এটি polling ভিত্তিক in-app notification। Crop catalog সীমিত; real weather মানেই সব crop advice validated নয়।

Deployment-এর জন্য `docs/DEPLOYMENT_BN.md` দেখুন। পুরোনো patch README-গুলো historical; এই release-এর জন্য এই নির্দেশনা অনুসরণ করুন।

## Crop update ব্যবহার করবেন যেভাবে

1. My profile-এ location, rooftop/balcony ও direct sunlight hours save করুন।
2. Dashboard-এ “Check one planned container”-এ depth (cm), volume (litres), drainage, support এবং Seeds/Seedling দিন। কোনো sensor লাগবে না। অজানা measurement ফাঁকা রাখুন; অনুমান করে 0 লিখবেন না।
3. Save container setup চাপুন। একই page-এ results refresh হবে।
4. “All crops”-এ ১০টি crop দেখবেন। “Planning shortlist”-এ শুধু বর্তমান মাস ও setup অনুযায়ী candidate/conditional crop থাকবে। Empty shortlist মানে app নষ্ট নয়; প্রতিটি card-এর কারণ দেখুন।
5. “Growing notes & evidence”-এ source ও সীমাবদ্ধতা দেখুন। “Track a plant” দিয়ে My Plants-এ যোগ করুন। Daily care-এ crop observation task পাবেন; এটি sensor measurement বা fertilizer prescription নয়।

Location বদলালে real weather বদলায়। District-specific crop calendar-এর যথেষ্ট evidence নেই বলে শুধু district বদলালে crop result একই থাকতে পারে। Seedling/transplant calendar verified না থাকলে seed calendar দিয়ে match করা হয় না।

বিস্তারিত: `docs/CROP_UPDATE_BN.md`; পরীক্ষার ফল: `docs/CROP_RELEASE_CHECKS.md`।
