# Crop update — data, logic ও শেখার guide

Release: 31 October 2026. Research review: 27 October 2026.

## কী যোগ হয়েছে

১০টি crop: বেগুন, টমেটো, মরিচ, ঢেঁড়স, লালশাক, পুঁইশাক, পালংশাক, ধনেপাতা, মুলা, শসা। All crops view, বাংলা/English search, status filter, কারণ, season notes, source links, container input এবং Track a plant আছে। Original images দেওয়া হয়নি; তাই crop symbols ব্যবহার হয়েছে, downloaded photo নয়।

User-provided `crop_information.pdf`-এর evidence sheets ব্যবহার হয়েছে। PDF-এর table/JSON-এর ডান পাশ clipped ছিল; এটি lossless JSON import নয়। যাচাই করা তথ্য curated JSON-এ রাখা হয়েছে। Unknown মান null; 0 নয়।

## Evidence ও recommendation আলাদা

| Crop | Core review | সীমাবদ্ধতা |
|---|---|---|
| বেগুন | Reviewed | Cultivar ও exact sowing dates দেখুন; October window পুরো মাস নয় |
| মরিচ | Reviewed | Pepper container guidance, BARI cultivar trial নয় |
| ঢেঁড়স | Reviewed | Depth unknown; conditional guidance |
| ধনেপাতা | Reviewed | Depth সাধারণ tap-rooted herb guidance; volume unknown |
| শসা | Reviewed | General BAMIS February–March sowing; cultivar difference থাকতে পারে |
| টমেটো | Incomplete | AIS source recheck হয়নি; reported season auto-match নয় |
| লালশাক | Incomplete | Numeric sunlight/pot requirements unknown |
| পুঁইশাক | Incomplete | Full sun বলা আছে; numeric sun/pot threshold unknown |
| পালংশাক | Incomplete | BAMIS page-এ garlic-related conflicting content; বিদেশি pot guidance-ও ভিন্ন |
| মুলা | Incomplete | বিদেশি ছোট-root guidance দিয়ে long BARI radish-এর pot depth নিশ্চিত নয় |

“Reviewed” মানে সব field verified নয়, কিংবা সব setup-এ crop suitable নয়। Supplementary scientific names, harvest estimates ও কিছু support notes PDF-reported হিসেবে label করা। Observation reminders সাধারণ guidance। Source link থাকলেই সেই crop automatic match নয়।

- বাংলাদেশের seed season BAMIS থেকে; বিদেশি planting calendar ব্যবহার হয়নি।
- Container/sunlight-এর জন্য university extension sources; inches × 2.54 = cm, US gallons × 3.785411784 = litres (rounded)। Original measurements containerNote-এ আছে। Diameter থেকে depth বা volume অনুমান করা হয়নি।
- District mapping তৈরি করা হয়নি। Location weather-এর জন্য; crop seasonal scope general Bangladesh, local certification নয়।
- Current Asia/Dhaka month দিয়ে coarse shortlist। Exact day/cultivar window card-এ পড়তে হবে।
- Transplant months unknown হলে seed months reuse হয় না। বর্তমানে transplant shortlist evidence-incomplete হতে পারে।
- ৭ দিনের forecast seasonal suitability প্রমাণ করে না। Weather unavailable হলেও catalog দেখা যায়।
- Pot input একটি planned container-এর জন্য; সব existing plant-এর টব একই ধরে নেওয়া হয় না।

## Data flow: কী, কেন, কীভাবে

1. **Profile → Go API → MySQL:** location, sunlight ও space save হয়। Sensor নেই; user input-ই observation।
2. **Container form → `/garden-setup`:** logged-in user ID server থেকে নেওয়া হয়। অন্য user-এর ID দিয়ে setup overwrite করা যায় না। Unknown nullable value database-এ NULL থাকে।
3. **`/recommendations`:** profile + pot setup + Asia/Dhaka month + crop evidence দিয়ে প্রতিটি crop assess করে। Low sunlight/small pot/no drainage → not_fit; unknown input → needs_setup; missing research → needs_evidence।
4. **All crops বনাম shortlist:** সব ১০টি entry দেখা যায়। শুধু candidate/conditional shortlist-এ; unknown requirement স্পষ্ট condition হিসেবে থাকে। Preferred sunlight maximum hard upper limit নয়।
5. **React:** response থেকে cards তৈরি করে। Save-এর পর API আবার load হয়। Database insert successful হলেও refreshed results fail করলে UI error দেখায়।
6. **My Plants → care tasks:** stable crop ID দিয়ে plant যুক্ত থাকে। Generic crop observation task দৈনিক একবার তৈরি হয়; repeat GET-এ completion reset হয় না। Growth stage user-reported।

## প্রধান file

- `backend/crop_catalog.json`: canonical curated research data + sources + field mapping।
- `backend/crop_research.go`: migration CLI, setup API, matching logic।
- `backend/garden.go`: plant catalog এবং daily crop notes integration।
- `src/CropGuide.jsx`: form, cards, search, evidence view।
- `migrations/009_crop_research.sql`: phpMyAdmin-compatible alternative to CLI upgrade।
- `scripts/build_crop_migration.py`: JSON থেকে 009 SQL regenerate করে।
- `backend/crop_research_test.go`, `src/CropGuide.test.jsx`, `scripts/crop_smoke.py`: rule/UI/API tests।

## পরে crop information বদলাতে

1. JSON-এর crop `slug` ও `legacyName` স্থির রাখুন; existing plant references এদের mapping-এর সঙ্গে যুক্ত।
2. Source মূল publication খুলে relevant field যাচাই করুন। `sources`, `fieldSources`, `checkedAt`, notes update করুন। Unknown null রাখুন। Cultivar/region conflict মুছবেন না।
3. `coreVerified=true` কেবল প্রয়োজনীয় season/container/sun evidence যাচাই হলে। এই flag পুরো crop scientifically certified করে না।
4. Project root-এ `python scripts/build_crop_migration.py` চালান (Python 3 দরকার শুধু dataset editing-এ)।
5. Backend-এ `go test ./...`, frontend root-এ `npm test`, `npm run build` চালান।
6. Database backup নিয়ে backend-এ `go run . -upgrade-crops` চালান, backend restart করুন। SQL alternative: phpMyAdmin-এ শুধু regenerated 009 import। দুটো একসঙ্গে প্রয়োজন নেই।

`-upgrade-crops` existing crop IDs, custom legacy descriptions ও user plants ধরে রাখে; research JSON refresh করে। Custom crop rows untouched থাকে। JSON-এর crop remove করলে existing evidence নিজে delete হয় না—references review না করে delete করবেন না।

## ছবি পরে যোগ করতে

নিজের বা reuse-permitted ছবি `public/crops/<slug>.jpg` রাখুন। JSON-এর `imagePath` দিন `/crops/<slug>.jpg`; ছবি না থাকলে null রাখুন। Source/permission `docs/crop-research/IMAGE_CREDITS.md`-এ লিখুন। তারপর dataset upgrade ও frontend rebuild করুন। কোনো ছবির URL বা licence অনুমান করবেন না।

## সীমা

এটি source-backed planning aid, agronomist review বা sensor monitoring নয়। Regional calendars, cultivar-specific missing sizes, transplant calendars এবং harvest timing-এর independent verification এখনো বাকি। Software workflow complete; evidence gaps ইচ্ছাকৃতভাবে visible রাখা হয়েছে। Public deployment-এর Docker/TLS ও security review আলাদা কাজ; deployment guide দেখুন।
