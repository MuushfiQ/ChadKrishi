> Historical patch guide. For this complete release, follow START_HERE_BN.md first.

# ChadKrishi — Marketplace, public profile ও profile picture update

এই update আগের chadkrishi-source-handoff project-এর জন্য। এটি changed/new files-এর patch; আলাদা নতুন project নয়। তোমার বর্তমানে চালু project folder-এই fileগুলো বসাবে।

## কী ঠিক হয়েছে

- Marketplace প্রথমে All listings দেখাবে। Buy করলে Reserved হবে এবং listing দেখা যাবে, তবে দ্বিতীয় buyer সেটি কিনতে পারবেন না। Buyer Paid বলার পর seller payment receipt confirm করলে Sold হবে। Admin archive করলে feed/marketplace থেকে সরবে; buy/sell history থাকবে।
- Community post-এর নাম বা profile picture-এ click করলে লেখকের profile খুলবে। সেখানে ছবি, bio, area, post count এবং তার আগের post/sale দেখা যাবে। পুরোনো post দেখতে Load more চাপবে। Message button profile header-এ; নিজের profile-এ Edit my profile।
- My profile-এ JPEG/PNG profile picture upload/change/remove করা যাবে। File বেছে নিলে automatic save হয়; বাগানের profile form আলাদাভাবে Save করবে। অন্য signed-in member feed/profile/comments/chat-এ ছবিটি দেখতে পারবেন।
- Refresh posts চাপলে অন্য account-এর সাম্প্রতিক update আসবে। History-তে Refresh আছে; chat খোলা থাকলে আগের মতো polling চলে।

## প্রথমে যা করবে

1. VS Code-এর backend ও frontend terminal-এ Ctrl+C দিয়ে app বন্ধ করো।
2. বর্তমানে চালু project folder-এর একটি backup copy রাখো।
3. Download করা ZIP **অন্য একটি temporary folder-এ Extract All** করো।
4. Extract করা folder-এ backend, src, scripts, docs এবং এই guide দেখবে। নিচের fileগুলো existing project-এর একই relative path-এ copy করো। Existing file হলে Replace করবে; নতুন file হলে create/copy করবে। Folder merge করে fileগুলো বসাবে, পুরো existing src/backend folder delete করবে না।

## কোন file কোথায় বসাবে

এখানে `project folder` মানে যে folder-এ package.json এবং backend folder আছে; যেমন D:\chadkrishi।

| ZIP-এর file | Existing project-এ destination | কাজ |
|---|---|---|
| backend/garden.go | project folder/backend/garden.go | Replace |
| backend/social.go | project folder/backend/social.go | Replace |
| backend/profile_photo.go | project folder/backend/profile_photo.go | নতুন file |
| src/Avatar.jsx | project folder/src/Avatar.jsx | নতুন file |
| src/App.jsx | project folder/src/App.jsx | Replace |
| src/Garden.jsx | project folder/src/Garden.jsx | Replace |
| src/Social.jsx | project folder/src/Social.jsx | Replace |
| src/styles.css | project folder/src/styles.css | Replace |
| src/Social.test.jsx | project folder/src/Social.test.jsx | Replace test |
| scripts/api_smoke.py | project folder/scripts/api_smoke.py | Replace integration test |
| docs/API.md | project folder/docs/API.md | Replace documentation |
| docs/progress.md | project folder/docs/progress.md | Replace documentation |
| README_UPDATE_BN.md | project folder/README_UPDATE_BN.md | নতুন নির্দেশনা |

VS Code দিয়ে হাতে paste করলে: ZIP-এর file খুলে Ctrl+A → Ctrl+C, নিজের project-এর একই file খুলে Ctrl+A → Ctrl+V → Ctrl+S। নতুন দুই file সঠিক folder-এ create করবে। File Explorer দিয়ে copy/replace করলেও একই কাজ হয়। `.go` file ভুল করে `.go.txt` এবং `.jsx` file `.jsx.txt` যেন না হয়।

এই patch-এর runtime code প্রথম ৮টি file; test ও documentation-এর fileগুলোও copy করলে project-এর রেকর্ড ও verification হালনাগাদ থাকবে। আগের handoff-এর পর নিজে এসব file-এ পরিবর্তন করে থাকলে নিজের পরিবর্তনের সঙ্গে merge করবে; backup দেখে নেবে।

## Database ও পুরোনো information

এই update-এ **নতুন SQL import/migration লাগবে না**। Existing users.profile_photo column ব্যবহার হচ্ছে। আগের 001–007 ইতিমধ্যে import করা থাকলে আর import করবে না।

নিজের existing `.env` এবং `backend/uploads` রেখে দেবে। আগের account, posts, orders, chat তোমার একই MySQL database-এ আছে। এই patch-এ account data বা uploaded ছবি নেই। .env/পুরো database delete বা reset করার প্রয়োজন নেই।

যে listing আগেই Buy করা হয়েছে এবং Reserved আছে, সেগুলোও Marketplace-এর All listings view-তে পাওয়া যাবে। Admin archived করা listing সেখানে আসবে না।

## Run করবে যেভাবে

1. XAMPP Control Panel-এ MySQL Start করো।
2. VS Code-এ নিজের **existing project folder** খোলো।
3. Terminal 1 project root-এ খুলে:

```powershell
cd backend
go run .
```

Terminal আগে থেকেই backend folder-এ থাকলে শুধু `go run .`।

4. Terminal 2 project root-এ খুলে:

```powershell
npm run dev
```

Dependency বদলায়নি; আগে কাজ করা node_modules থাকলে নতুন install লাগবে না। node_modules না থাকলে প্রথমে `npm ci`, তারপর `npm run dev`।

5. Browser-এ http://localhost:5173 খোলো; Ctrl+Shift+R দাও। Vite অন্য port দিলে root .env-এর FRONTEND_ORIGIN-এর সঙ্গে frontend port মেলাতে হবে; আগের কাজ করা 5173 ব্যবহার করাই সহজ।

## দুই account দিয়ে পরীক্ষা

1. Seller দিয়ে My Profile → Profile picture-এ ছবি বেছে নাও। Automatic save-এর success দেখো। Refresh করে ছবিটি থাকে কি না দেখো। Bio/garden form বদলালে আলাদা Save চাপবে।
2. Seller একটি Community post ও sale listing publish করবে।
3. অন্য browser/Incognito-তে Buyer login করবে। Author-এর নাম/ছবিতে click করবে: profile, bio, আগের post, Message button দেখবে। Message চাপলে আগের/নতুন private conversation খুলবে। Community post-এর নিচে Message button থাকবে না।
4. Buyer Marketplace-এর All listings-এ listing Buy করবে; Buy/sell history খুলবে। Seller Marketplace-এ Refresh posts দিলে listing Reserved দেখবে। তৃতীয় buyer আবার কিনতে পারবেন না।
5. Buyer সরাসরি payment করার পরে history-তে I have paid চাপবে। Seller history-তে Refresh দিয়ে Yes, payment received চাপবে। এখন Marketplace Refresh করলে Sold।
6. Admin archive করলে marketplace/feed-এ listing থাকবে না; দুই পক্ষের history থাকবে।
7. My Profile থেকে ছবি Remove করো। Avatar-এ নামের প্রথম অক্ষর দেখাবে; browser refresh করে persistence দেখো।

## ধারণাগুলো সহজ ভাষায়

আগে Buy করার পরে backend listing-কে reserved করত, কিন্তু frontend available filter দিয়ে request পাঠাত। তাই data delete না হলেও list থেকে সরে যেত। এখন default filter All listings; reservation এবং payment rules আগের মতো server enforce করে।

ছবি upload করলে server validated file backend/uploads-এ রাখে এবং /api/media/... path দেয়। PUT /profile/photo নিজের account-এর users.profile_photo-তে path save করে। Database-এ image bytes রাখা হয় না। ছবি বদলানো পুরো garden setup save করার ওপর নির্ভর করে না।

Profile link-এ author ID দিয়ে GET /people/:id request হয়। Profile-এর timeline GET /posts?author=id দিয়ে শুধু সেই author-এর nonarchived post নেয়। Next cursor দিয়ে পুরোনো post আনা যায়। Email, password বা precise coordinates public profile-এ পাঠানো হয় না।

## Verification

এই release-এ ১৭টি frontend test, frontend production build, Go tests, go vet এবং fresh isolated MariaDB-তে real API integration suite পাস করেছে। Photo ownership/removal, payment অপেক্ষায় reserved listing visibility, seller receipt-এর পর Sold, archive/history, profile timeline এবং private chat যাচাই করা হয়েছে। এখানে তোমার Windows/XAMPP-এ সরাসরি চালানো বা desktop/mobile browser visual verification করা হয়নি।

নিজের project-এ component/unit tests চালাতে:

```powershell
npm test
```

Backend folder-এর terminal-এ:

```powershell
go test ./...
```

scripts/api_smoke.py test accounts/posts তৈরি করে; এটি চালাতে আলাদা disposable test database/server ব্যবহার করবে। নিজের দৈনন্দিন database-এ চালানোর প্রয়োজন নেই।
