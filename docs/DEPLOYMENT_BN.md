# Public deployment — Docker Compose

এই release-এ Dockerfiles, compose.yaml ও Caddy reverse proxy configuration আছে। **এই environment-এ Docker daemon নেই, তাই container build/start, TLS issuance ও public deployment যাচাই হয়নি।** Go build, frontend build ও application tests-এর ফল progress.md-তে আছে। আগে staging server-এ নিচের journey যাচাই করুন।

## নতুন server / নতুন database

প্রয়োজন: Linux VPS, Docker Engine ও Compose plugin, নিজের domain, DNS A/AAAA record সঠিক server-এ, TCP 80/443 inbound। Database ও API public port expose করা হয়নি।

1. সম্পূর্ণ source server-এ রাখুন।
2. `deploy.env.example` copy করে `deploy.env` করুন। DOMAIN-এ শুধু hostname (https:// নয়), দুই password-এ আলাদা দীর্ঘ random alphanumeric value দিন। এটি Git-এ commit করবেন না।
3. চালান:

```sh
docker compose --env-file deploy.env config --quiet
docker compose --env-file deploy.env up -d --build
docker compose --env-file deploy.env logs --tail=100 api
docker compose --env-file deploy.env exec api chadkrishi -check-db
```

4. `https://YOUR_DOMAIN/api/v1/health` পরীক্ষা করুন। App `https://YOUR_DOMAIN`। দুই account-এ registration/login, photo, order/cancel/payment, chat/notification, logout এবং refresh পরীক্ষা করুন।

Caddy static React build serve করে; /api/* Go API-তে যায়। একই HTTPS origin-এ secure session cookie ব্যবহৃত হয়। প্রথমবার empty database volume হলে MariaDB 001–009 init scripts চালায়। Existing volume-এ scripts আবার চলে না। `docker compose down -v` করবেন না—এতে persistent data মুছে যাবে।

## আপনার XAMPP data রাখতে চাইলে

Default Docker installation একটি আলাদা database তৈরি করবে; Windows database নিজে থেকে সংযুক্ত হবে না। আগে XAMPP থেকে SQL export ও uploads backup নিন। Staging database-এ backup restore করে schema check ও notifications repair পরীক্ষা করুন; তারপর production migration পরিকল্পনা করুন। Production data রেখে init migrations পুনরায় চালাবেন না। Upload files-ও persistent uploads volume-এ restore করতে হবে, UID 10001-এর write permission সহ। SQL backup restore test ছাড়া live migration করবেন না।

## Existing deployment upgrade

Source update/build করুন। Missing notifications-এর জন্য (DB user-এর CREATE permission থাকলে):

```sh
docker compose --env-file deploy.env run --rm api -upgrade-crops
docker compose --env-file deploy.env run --rm api -repair-notifications
docker compose --env-file deploy.env up -d --build
```

যদি image বদলানো হয়, আগে `docker compose --env-file deploy.env build api` চালান। অন্য missing/incompatible schema থাকলে command সুনির্দিষ্ট error দেবে; automatic destructive repair করে না।

## Operations ও release সীমা

- Database ও uploads দুটোর নিয়মিত backup এবং restore পরীক্ষা দরকার।
- Log দেখতে `docker compose --env-file deploy.env logs --tail=100 api`। Credentials/backup public share নয়।
- Migration account ও runtime account আলাদা করা যেতে পারে; runtime-এর প্রয়োজনীয় SELECT/INSERT/UPDATE/DELETE, schema check-এর metadata visibility থাকতে হবে।
- Internet-facing production-এর আগে rate limiting/abuse controls, monitoring, privacy/moderation process ও security review প্রয়োজন। এই release-কে audited production system বলা হচ্ছে না।
- Social notifications in-app polling; browser বন্ধ থাকলে push alert নেই। Payment manual; real payment gateway নেই।
- `scripts/api_smoke.py` fixtures তৈরি ও table rename করে: শুধু disposable test database-এ, আপনার বাস্তব database-এ নয়।

Official references:
- https://caddyserver.com/docs/automatic-https
- https://caddyserver.com/docs/caddyfile/directives/reverse_proxy
- https://github.com/docker-library/docs/blob/master/mariadb/README.md
