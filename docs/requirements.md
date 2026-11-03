> Scope update (22 October 2026): user explicitly requested whole-listing orders, buyer Paid / seller Received confirmation, persistent buy/sell history, admin archive, four reactions and private text chat. These are included; automatic payment processing, delivery and calls remain excluded. See IMPLEMENTATION_BN.md.

# ChadKrishi requirements

ChadKrishi is a university demonstration application for gardeners in Bangladesh.
It uses a React/Vite frontend, a Go/Gin API, and MySQL. The design favors small,
explainable features over enterprise infrastructure.

## First release

1. Account and profile: register, login, logout, session, and editable gardening
   profile. Public views expose only general area, not email or exact coordinates.
2. Crop guidance: 6–8 clearly labelled demonstration crop profiles and
   rule-based recommendations using month, space, and sunlight.
3. Plants: add a crop, nickname it, record planting information, stage, photo,
   notes, and archive it.
4. Care: generate idempotent daily checklists on dashboard/plant requests.
   Weather is cached and advisory; routine care remains available when weather
   is unavailable.
5. Community: text/image posts, post types, comments, likes, pagination, and
   public profiles.
6. Marketplace: sale posts are the same posts shown in the feed. Listings have
   BDT price, quantity, general area, optional published contact, and available/
   sold state.

## Explicit boundaries

- No exact home addresses, payments, carts, delivery, chat, notifications while
  the app is closed, plant-health detection, or fixed water/fertilizer dosing.
- Uploaded images are local files with generated names and strict validation.
- Stored event timestamps use UTC; daily task dates use `Asia/Dhaka`.
- Crop and care guidance is limited demonstration data and must show its
  verification status.

## Initial location coverage

The first seed set covers Dhaka, Chattogram, Rajshahi, Khulna, Sylhet, and
Rangpur using city-level coordinates. Users select a city/general area; the
application does not request a precise home address.

