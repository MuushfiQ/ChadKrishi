> Historical patch guide. For this complete release, follow START_HERE_BN.md first.

# ChadKrishi — Notification, custom care ও automatic update

এই ZIP তোমার existing project-এর cumulative update। আগের profile picture, public profile এবং marketplace fix-ও এতে আছে। আলাদা নতুন project হিসেবে এই ZIP run করবে না; existing project-এ fileগুলো বসাবে।

## কী যোগ হয়েছে

| ঘটনা | Notification কার কাছে যাবে |
|---|---|
| Buyer Buy করেন | Seller |
| Buyer I have paid দেন | Seller |
| Seller payment received confirm করেন | Buyer |
| Seller Not received জানান | Buyer |
| Buyer বা seller reservation cancel করেন | অপর পক্ষ |
| একজন আরেকজনকে message দেন | Message recipient |

Topbar-এর bell-এ unread count ও নতুন activity-এর ছোট popup থাকবে। Bell খুললেই সব read হবে না। Notification-এ click করলে সংশ্লিষ্ট order/chat খুলবে; চাইলে Mark all read করা যাবে। নতুন listing publish করলে সেটি feed-এ update হবে; সব member-কে ব্যক্তিগত notification পাঠানো হয় না।

App খোলা ও connected থাকলে notification প্রায় ৫ সেকেন্ড, chat/inbox ৪ সেকেন্ড, order status ৫ সেকেন্ড, feed ১০ সেকেন্ড অন্তর update হয়। Browser tab hidden থাকলে polling বিরতি নেয়; ফিরে এলে আবার update হয়। পুরো browser reload হয় না, draft থাকে। Browser বন্ধ থাকলে system popup আসবে না; notification database-এ জমা থাকবে এবং আবার login/app খুললে দেখা যাবে। এটি in-app notification, closed-browser push নয়।

Community/Marketplace/profile timeline-এর নিচে scroll করলে পরের postগুলো নিজে load হবে। Load more button fallback হিসেবেও আছে।

Daily care-এ প্রতিটি plant-এর আলাদা card ও completion count থাকবে। একই জাত/নামের দুটি plant-ও আলাদা থাকবে। Add custom task থেকে plant, date, title ও optional note দিতে পারবে। আজ বা ভবিষ্যতের দিনের task করা যায়; এটি একদিনের task, recurring নয়। Future date-এ automatic weather care আগাম তৈরি হবে না—সেদিনের জন্য নিজের scheduled task দেখবে। নিজের custom task delete করা যায়; system checklist নয়।

## Existing project-এ update বসানো

1. VS Code-এর frontend/backend terminal-এ Ctrl+C দাও। Existing project folder-এর backup copy রাখো।
2. ZIP আলাদা temporary folder-এ Extract All করো।
3. ZIP-এর src, backend, scripts, docs এবং migrations-এর **ভেতরের fileগুলো** নিজের project-এর একই path-এ copy করো। Existing file হলে Replace, নতুন file হলে add। পুরো existing folder delete করবে না।
4. নিজের `.env` ও `backend/uploads` রাখবে। এই ZIP-এ credentials বা user upload নেই।

Project root মানে package.json আছে যে folder-এ, যেমন D:\chadkrishi। যদি আগের handoff-এর পরে নিজে code বদলে থাকো, backup দেখে নিজের পরিবর্তনের সঙ্গে merge করবে।

| File | বসানোর নিয়ম |
|---|---|
| backend/garden.go, backend/social.go | একই file Replace |
| backend/profile_photo.go | আগের photo patch থাকলে Replace; নইলে নতুন |
| backend/notifications.go, backend/custom_tasks.go, backend/custom_tasks_test.go | নতুন file |
| src/App.jsx, src/Garden.jsx, src/Social.jsx, src/styles.css | একই file Replace |
| src/Avatar.jsx | আগের photo patch থাকলে Replace; নইলে নতুন |
| src/Care.jsx, src/Notifications.jsx, src/live.js, src/Activity.test.jsx | নতুন file |
| src/Social.test.jsx, scripts/api_smoke.py | test file Replace |
| docs/API.md, docs/progress.md | documentation Replace |
| migrations/008_notifications.sql | নতুন migration file |
| README_LIVE_CARE_BN.md | project root-এ রাখো |

VS Code দিয়ে paste করলে পুরো file-এর content Ctrl+A → Ctrl+C, নিজের একই file-এ Ctrl+A → Ctrl+V → Ctrl+S। নতুন file সঠিক folder-এ তৈরি করবে। File Explorer দিয়ে copy/replace করাই সহজ।

## এইবার database-এ শুধু 008 Import লাগবে

আগের profile-picture update-এ SQL লাগেনি। **এই নতুন notification update-এ লাগে।** তোমার existing chadkrishi database-এ আগের 001–007 আছে ধরে নিচের ধাপ:

1. XAMPP-এ MySQL এবং phpMyAdmin খোলার জন্য Apache চালু করো।
2. Browser-এ http://localhost/phpmyadmin খোলো।
3. বাঁ পাশে existing **chadkrishi** database নির্বাচন করো। নতুন database বানাবে না।
4. Import tab → Choose File → `migrations/008_notifications.sql` নির্বাচন করো।
5. Import/Go চাপো। সফল হলে database-এ notifications table যোগ হবে। আগের account/post/order/chat থাকবে।

001–007 আবার চালাবে না। 008 table না থাকলে তৈরি করে; আগে থাকলে একই table রেখে দেয়। Import error এলে error text দেখে ঠিক করে তারপর backend চালাবে।

## Run

XAMPP MySQL চালু রাখো। VS Code-এ existing project folder খোলো। Terminal 1 project root-এ:

```powershell
cd backend
go run .
```

Terminal আগে থেকেই backend folder-এ থাকলে শুধু `go run .`। Terminal 2 project root-এ:

```powershell
npm run dev
```

আগে dependencies install করা থাকলে আবার npm install লাগবে না। node_modules না থাকলে `npm ci` চালাবে। Browser-এ http://localhost:5173 খোলো এবং update বসানোর পরে একবার Ctrl+Shift+R দাও। এরপর নিয়মিত message/feed দেখতে refresh চাপতে হবে না।

## দুই browser দিয়ে পরীক্ষা

1. Browser A-তে seller, Browser B/Incognito-তে buyer login করো।
2. Buyer seller-এর profile খুলে Message পাঠাও। Seller Dashboard-এ থাকলেও কয়েক সেকেন্ডের মধ্যে bell count বাড়বে। Notification-এ click করলে chat খুলবে।
3. Seller chat-এ থাকলে buyer-এর পরের message নিজে আসবে। Message draft লিখে রেখে দেখো—incoming message draft মুছে ফেলবে না।
4. Buyer listing Buy করলে seller notification পাবেন। Buyer Paid দিলে seller-এর order history নিজে update হবে। Seller received confirm করলে buyer notification পাবেন এবং listing Sold হবে।
5. Daily care → Add custom task → নিজের plant, date, title → Save custom task। নির্দিষ্ট plant-এর card-এ task দেখবে। Complete করো; refresh-এর পরও থাকবে। Future task দিলে checklist date-ও সেই দিনে যাবে।
6. একাধিক plant-এর checklist আলাদা card-এ দেখবে। Plant-এর নামের পাশে ID একই নামের গাছ আলাদা করতে সাহায্য করে।
7. ২০টির বেশি post থাকলে feed-এর নিচে scroll করো; পরের page নিজে load হবে। আরেক browser-এ নতুন post করলে feed-এ প্রায় ১০ সেকেন্ডের মধ্যে আসবে।
8. Seller logout থাকা অবস্থায় buyer message পাঠাও। Seller আবার login করলে unread notification থাকবে। Update-এর আগের পুরোনো activity থেকে notification তৈরি করা হয়নি।

## Data flow সহজভাবে

Buyer Buy/Paid বা seller confirmation দিলে Go backend প্রথমে permission ও current state যাচাই করে। একই database transaction-এ action এবং recipient-এর notification save হয়। একটি fail হলে পুরো transaction rollback হয়, ফলে successful action-এর notification বাদ পড়ে না।

React ছোট interval-এ API থেকে নতুন data আনে এবং UI state update করে। Scroll sentinel দৃশ্যমান হলে cursor ব্যবহার করে পরের page আনে। প্রতিটি post ID দিয়ে duplicate সরানো হয়। Filter বদলালে পুরোনো request-এর response বাদ দেওয়া হয়।

Custom task-এর সঙ্গে plant_id save হয়। Backend নিশ্চিত করে plant তোমার এবং active। UI plant_id অনুযায়ী task group করে, তাই একই crop-এর একাধিক plant মিশে যায় না।

## পরীক্ষা

২৭টি frontend test, frontend build, Go tests ও go vet পাস করেছে। আলাদা test MariaDB-তে migration, real API notification/care/payment/chat flow পাস করেছে। Notification save ব্যর্থ করেও পরীক্ষা করা হয়েছে—message/purchase-ও rollback হয়। তোমার নিজের XAMPP-এ বা desktop/mobile visual browser session-এ সরাসরি পরীক্ষা করা হয়নি।

নিজে unit/component test করতে project root-এ `npm test`; backend folder-এ `go test ./...`। `scripts/api_smoke.py` আলাদা disposable test database-এর জন্য, কারণ এটি test accounts/posts তৈরি করে এবং test table সাময়িক rename করে rollback যাচাই করে। নিজের দৈনন্দিন database-এ চালাবে না।
