# efs-ical

CLI that generates subscribable `.ics` calendars (Outlook, Google Calendar,
Android) for upcoming blood donation collections from EFS
(Établissement Français du Sang), for one or more locations.

## Commands

### 1. Geocode a location

```
go run . geocode <city or zip code>
```

Example:

```
go run . geocode 44130
```

Lists matching communes, prompts you to pick one by number, then saves it
to `dist/locations.json`. This only records the location — it does not fetch
events or write any `.ics` file yet.

### 2. Build (or rebuild) the ICS file(s)

```
go run . build [normalized-name ...]
```

- No arguments → rebuilds every location saved in `dist/locations.json`.
- One or more names → rebuilds only those (the normalized name printed by
  `geocode`, e.g. `blain`).

Example:

```
go run . build blain
```

Fetches upcoming collections for that location and writes `dist/blain.ics`.
Re-run anytime to refresh with new/updated events — event UIDs are stable,
so calendar apps update in place rather than duplicating.

## Subscribing

- **Outlook**: Add calendar → From internet → paste URL
- **Google Calendar**: Other calendars → + → From URL → paste URL
- **Android**: auto-syncs once added to the linked Google account

## Files

- `dist/locations.json` — saved locations (append-only via `geocode`)
- `dist/<normalized-name>.ics` — generated calendar per location (via `build`)
