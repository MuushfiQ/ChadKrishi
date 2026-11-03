# ChadKrishi — নতুন features ও data flow

## ১. Existing project update (Windows / XAMPP)

আগের project folder এবং XAMPP-এর `chadkrishi` database export করে backup রাখুন। ZIP-এর `chadkrishi` folder থেকে source files আপনার project-এ copy করুন। নিজের `.env` এবং `backend/uploads` রাখুন—এগুলো নতুন ZIP-এ নেই।

আপনার database-এ আগেই migrations 001–006 চালানো থাকলে **শুধু `migrations/007_social_commerce.sql` একবার চালাবেন**। XAMPP থেকে MySQL Start → phpMyAdmin → Import → 007 file নির্বাচন করুন। Duplicate column error এলে একই migration বারবার চালাবেন না; আগে schema পরীক্ষা করুন। নতুন database হলে 001 থেকে 007 ক্রমানুসারে চালান। Migration 007 পুরোনো user, crop, post বা plant মুছে না; পুরোনো likes `like` reaction হয়।

VS Code-এ প্রকৃত project folder খুলুন → Terminal → New Terminal। এখানে `package.json` থাকতে হবে।

Terminal 1:

```powershell
cd backend
go run .
```

**`go run main.go` ব্যবহার করবেন না**—নতুন backend একাধিক Go file ব্যবহার করে। Go 1.26+ লাগবে।

Terminal 2, project root থেকে:

```powershell
npm ci
npm run dev
```

Browser: http://localhost:5173 । API health: http://localhost:8080/api/v1/health । নিজের `.env` না থাকলে root `.env.example` copy করে `.env` করুন এবং MySQL password/port ঠিক করুন। Existing `.env` overwrite করবেন না।

## ২. Admin account

প্রথমে সাধারণভাবে নিজের account register করুন। এরপর phpMyAdmin-এ database owner হিসেবে নিচের SQL-এ নিজের **সঠিক email** বসিয়ে চালান:

```sql
UPDATE users SET is_admin = TRUE WHERE email = 'your-actual-email@example.com';
```

Page refresh করলে sidebar-এ Admin archive আসবে। Registration/profile form থেকে admin হওয়া যায় না। এটি database owner-এর কাজ; সাধারণ ব্যবহারকারী এই permission পায় না।

## ৩. Buy → Paid → Payment received → Sold → Archive

- Seller একটি listing প্রকাশ করেন। Price হলো **পুরো listing-এর মোট দাম**, quantity হলো একসঙ্গে দেওয়া item-এর সংখ্যা। Partial quantity purchase এই version-এ নেই।
- অন্য user Buy চাপলে একটি order তৈরি হয়, listing Reserved হয়। Seller-এর Buy/sell history-তে `Waiting for payment` এবং `Payment received?` দেখা যায়।
- Buyer seller-এর সঙ্গে chat করে cash/bKash/অন্য উপায়ে বাস্তবে টাকা দেন। তারপর নিজের history-তে method ও optional transaction reference দিয়ে `I have paid` চাপেন। PIN, OTP বা password দেবেন না।
- Buyer Paid বললেই listing Sold হয় না। Seller নিজের cash/payment account যাচাই করে `Yes, payment received` চাপলে order Completed ও listing Sold হয়।
- Payment না এলে seller `Not received` চাপতে পারেন; buyer আবার যাচাই করে Paid দেবেন। Waiting অবস্থায় দুপক্ষের যেকেউ reservation cancel করতে পারেন। Paid অবস্থায় cancel বন্ধ থাকে।
- Admin Sold listing archive করলে Community/Marketplace থেকে লুকানো হয়। Buyer ও seller-এর history অক্ষত থাকে।
- এটি manual payment acknowledgement; bKash gateway, automatic verification, delivery tracking, escrow বা automatic refund নেই।

### কেন transaction ও row lock?

একই সময়ে দুই buyer Buy চাপতে পারেন। Backend `SELECT ... FOR UPDATE` দিয়ে listing lock করে। প্রথম order commit হলে পরের buyer Reserved status দেখবেন। তাই এক listing একসঙ্গে দুজন reserve করতে পারেন না। Order update ও Sold update একই database transaction-এ হয়—যেকোনো অংশ fail হলে পুরো পরিবর্তন rollback হয়।

Order-এ title, total amount ও quantity-এর snapshot থাকে। Listing edit হলেও পুরোনো transaction-এর তথ্য বদলায় না। Active/completed order থাকলে listing edit বন্ধ; কোনো order থাকলে post delete বন্ধ। `order_events`-এ transition actor ও সময় থাকে।

## ৪. Community ও reactions

Image upload → backend JPEG/PNG decode/size যাচাই → image re-encode → local uploads → API image URL → post caption ও URL database-এ save। Re-encoding পুরোনো metadata বাদ দেয়। ছবি সর্বোচ্চ 5 MB / 20 megapixels; server-generated filename ব্যবহার হয়।

Community update, প্রশ্ন, blog ও sale একই feed-এ আসে। Feed-এর photo post-এ caption, Like/Love/Angry/Sad ও comments আছে। একজন এক post-এ এক reaction দিতে পারেন; আরেকটি চাপলে বদলায়, একইটি চাপলে সরানো হয়। Database-এ `(post_id,user_id)` unique হওয়ায় duplicate reaction হয় না।

React server থেকে state নিয়ে UI দেখায়। Button → API call → session/ownership check → SQL → response → UI update। Backend-ই permission ঠিক করে; শুধু button লুকিয়ে security তৈরি হয় না।

## ৫. Private chat

কারও post/profile/order থেকে Message → দুজনের একটি conversation → text messages SQL-এ save। শুধু ওই দুই member thread পড়তে/লিখতে পারেন। Chat screen প্রতি 4 সেকেন্ডে নতুন messages আনে; পুরোনো messages pagination দিয়ে দেখা যায়। Read receipt আছে। Text-only chat; attachment/call/group chat নেই। এটি end-to-end encrypted messenger নয়।

## ৬. My Plants ও daily care

নিজের crop, nickname, planting date, seed/seedling, observed stage, photo, private notes save/edit/archive করা যায়। Recommendation card থেকে Add to my plants এবং plant থেকে Sell this plant আছে।

Stage আপনি দেখে দেন; sensor নেই। Plant-এর private notes sale caption-এ স্বয়ংক্রিয়ভাবে যায় না। Daily care হলো soil/drainage/leaf observation, সঙ্গে stage অনুযায়ী prompt। `(plant_id,rule_key,due_date)` unique—refresh করলে একই task duplicate হয় না।

Fresh weather পাওয়া গেলে app-এর precaution thresholds অনুযায়ী rain/wind check যোগ হতে পারে। Weather unavailable হলেও সাধারণ checklist পাওয়া যায়। এগুলো official alerts, disease diagnosis বা crop-specific fertilizer dose নয়।

## ৭. Weather ও recommendations

আগের Open-Meteo current/7-day weather integration ও source-backed catalog রাখা হয়েছে। API fail করলে unavailable/stale label আসে, demo weather বানানো হয় না। Recommendation profile area, season, sunlight ও gardening space ব্যবহার করে। Catalog সীমিত; matching crop না থাকলে empty state সঠিক। নতুন crop-এর agronomic তথ্য গবেষণা ছাড়া যোগ করা হয়নি।

## ৮. হাতে যাচাই করুন

দুই browser/profile ব্যবহার করে Seller ও Buyer register করুন। Seller-এর profile area save করুন। Plant এবং photo-caption post তৈরি করুন; Buyer reaction/comment দিন। Seller sale listing দিন; Buyer Buy → Paid; Seller history Refresh → Payment received; Marketplace Sold filter-এ ফল দেখুন। Admin archive করুন; দুজনের history refresh করে record আছে দেখুন। Chat-এ দুইদিকে message দিন; page refresh-এ message থাকবে। Sign out করে private URL খুললে login আসবে।

## ৯. শেখার জন্য ছোট অনুশীলন

1. DevTools Network-এ Buy request এবং response দেখুন; `social.go`-এর `buy` function-এর SQL-এর সঙ্গে মিলিয়ে নিন।
2. `nextOrderState` পড়ুন: buyer কেন received করতে পারে না? একই rule-এর Go test দেখুন।
3. Reaction বদলে Network request ও database-এর একটি row পর্যবেক্ষণ করুন।
4. এক task complete করে refresh দিন; unique key ও completed timestamp-এর ভূমিকা ব্যাখ্যা করুন।
5. Chat-এর thread ID বদলালে অন্য account কেন 404 পায়, ownership query দেখে বুঝুন।

## ১০. কী পরীক্ষা হয়েছে, কী বাকি

৯টি frontend test, ১৩টি Go test, build ও go vet pass করেছে। আলাদা MariaDB-তে actual API দিয়ে দুই buyer-এর concurrent reservation, payment confirmation, chat privacy, archive/history এবং expired session পরীক্ষা pass করেছে। এখানে Chrome চালানোর Unix socket নিষেধ থাকায় desktop/mobile visual test শেষ হয়নি। আপনার PC-তে উপরের manual verification করুন; `docs/progress.md`-এ সীমাবদ্ধতা দেওয়া আছে।
