# Compendium Tab - Implementation Plan

## Context
Add a cross-campaign "Compendium" tab to Lootsheet for browsing D&D reference data (monsters, spells, items, rules, conditions) fetched directly from D&D Beyond's API. No proxy — we call DDB services directly. The user provides a cobalt cookie at sync time (not stored) for monsters/spells/items. Rules and conditions come from the public config endpoint and need no auth.

## Architecture Decisions

1. **Direct DDB API, no proxy** — We call DDB's services directly (Go CLI, no CORS). Auth: cobalt cookie → `auth-service.dndbeyond.com` → bearer token → data endpoints.
2. **Compendium sub-tabs** (browse only): Monsters, Spells, Items, Rules, Conditions (virtual sections at 200+).
3. **Source selection & sync live in Settings** — new `SettingsTabCompendium` sub-tab under `@`. Lists 190 DDB sources with `t` toggle. `s` triggers sync (prompts cobalt cookie for monsters/spells/items; rules/conditions sync without auth).
4. **Cobalt cookie prompted on sync** — not persisted. Bearer token cached in memory for the session only.
5. **Data cached in SQLite** cross-campaign (no `campaign_id` FK), with FTS5 for search.

## DDB API Endpoints (validated with test script)

| Data | URL | Auth | Notes |
|------|-----|------|-------|
| Token exchange | `POST https://auth-service.dndbeyond.com/v1/cobalt-token` | Cookie: `CobaltSession={cobalt}` | Returns `{token: "..."}` |
| Game config | `GET https://www.dndbeyond.com/api/config/json` | **None** | 190 sources, 15 conditions, 18 basic actions, 64 rules, 51 weapon properties, plus lookup tables (CR, sizes, monster types) |
| Monsters | `GET https://monster-service.dndbeyond.com/v1/Monster?search=&skip=0&take=100&sources=...` | Bearer | Paginated (100/page), wrapped in `{data:[...]}` |
| Items | `GET https://character-service.dndbeyond.com/character/v5/game-data/items?sharingSetting=2` | Bearer | 2529 items, wrapped in `{data:[...]}` |
| Spells | `GET https://character-service.dndbeyond.com/character/v5/game-data/spells?classId={id}&classLevel=20&sharingSetting=2` | Bearer | Per-class, 550 wizard spells, wrapped in `{data:[...]}` |

### Config endpoint provides (no auth needed):
- `conditions[15]` — id, name, slug, description (HTML), type, levels
- `basicActions[18]` — id, name, description (HTML), activation type (Attack, Dash, Dodge, Help, Hide, Ready, etc.)
- `rules[64]` — id, name, description (plain text)
- `weaponProperties[51]` — id, name, description (HTML)
- `challengeRatings[34]` — lookup for monster CR display
- `creatureSizes[7]` — lookup for monster size display
- `monsterTypes[14]` — lookup for monster type display
- `sources[190]` — full source book list for selection

### Key field mappings (from sample dumps):

**Monster** (flat object):
- `id`, `name`, `armorClass`, `armorClassDescription`, `averageHitPoints`, `hitPointDice.diceString`
- `challengeRatingId` → lookup in config `challengeRatings`
- `sizeId` → lookup in config `creatureSizes`
- `typeId` → lookup in config `monsterTypes`, `tags` (e.g. "Goblinoid")
- `stats[0-5]` → STR/DEX/CON/INT/WIS/CHA (statId 1-6, value)
- `sensesHtml`, `skillsHtml`, `languageDescription` — pre-formatted strings
- `specialTraitsDescription`, `actionsDescription`, `bonusActionsDescription`, `reactionsDescription` — HTML
- `movements[].speed`, `sources[].sourceId`

**Item** (flat object):
- `id`, `name`, `type`/`filterType`, `rarity` (string: "Very Rare"), `canAttune` (bool)
- `description` — HTML
- `magic` (bool), `tags[]`, `sources[].sourceId`

**Spell** (nested under `definition`):
- `definition.id`, `definition.name`, `definition.level`, `definition.school` (string: "Conjuration")
- `definition.components` → [1,2,3] = V,S,M; `definition.componentsDescription`
- `definition.concentration` (bool), `definition.ritual` (bool)
- `definition.duration` → `{durationInterval, durationType, durationUnit}`
- `definition.range` → `{origin, rangeValue}`
- `definition.description` — HTML
- `definition.sources[].sourceId`

## Phased Delivery

### Phase 1: Tab Skeleton (UI only, no data) — DONE
Compendium tab appears in navigation with 5 sub-tabs, arcane blue theme, `8` shortcut. All tests pass.

**Files modified:**
- `src/render/model/section.go` — `SectionCompendium` iota, virtual tabs at 200+, `CompendiumTabs`, `OrderedSections`, `Title()`
- `src/render/model/action.go` — `ActionShowCompendium`, `ActionSyncCompendium`
- `src/render/model/keymap.go` — Key `8` binding, label `"1-8 jump"`
- `src/render/model/data.go` — 5 `ListScreenData` fields on `ShellData`
- `src/render/section.go` — Re-exports, `sectionStyleFor` with scatter glyphs
- `src/render/keymap.go` — Action re-exports
- `src/render/theme.go` — `SectionCompendium`/`ScatterCompendium` styles (arcane blue)
- `src/render/canvas/panel.go` + `src/render/panel.go` — `ScatterCompendium` glyphs
- `src/render/shell.go` — `compendiumTab` field, `activeCompendiumSection()`, `listSection()`, `listDataForSection()`
- `src/render/shell_action.go` — Navigation handler, sub-tab cycling, quit-to-dashboard, compose/input ignore lists
- `src/render/shell_render.go` — `renderCompendiumSection()` with tab bar
- `src/render/shell_footer.go` — Help text, header lines, glossary
- `src/render/shell_data.go` — Defaults, error data, resolve, empty checks, default list screen data
- `src/render/shell_test.go` — Updated assertions for wider tab bar

### Phase 2: Database & Domain Layer
Schema for caching compendium data. Cross-campaign (no campaign_id).

**New files:**
- `src/config/setup/migrations/013_add_compendium.sql` — All tables, FTS5 indexes, triggers
- `src/ledger/compendium/types.go` — Monster, Spell, Item, Rule, Condition, Source structs
- `src/ledger/compendium/repo.go` — List/Search, Upsert (bulk sync), ListSources, SetSourceEnabled. Uses existing `ledger.WithDB`/`WithDBResult` patterns.
- `src/ledger/compendium/repo_test.go`

**Modify:**
- `src/config/init.go` — Bump `SchemaVersion` from `"12"` to `"13"`

### Phase 3: DDB API Client
HTTP client for direct DDB API calls. No dependency on ledger — returns pure Go types.

**New package `src/ddb/`:**
- `client.go` — `Client` struct with `httpClient`, `bearerToken`. `Authenticate(ctx, cobalt) error` exchanges cobalt for bearer. In-memory token, not persisted.
- `types.go` — DDB API response types (`DDBMonster`, `DDBSpell`, `DDBItem`, `DDBCondition`, `DDBBasicAction`, `DDBRule`, `DDBSource`, `DDBConfig` with lookup tables)
- `monsters.go` — `FetchMonsters(ctx, sourceIDs) ([]DDBMonster, error)` — paginated (100/page)
- `spells.go` — `FetchSpells(ctx, classIDs) ([]DDBSpell, error)` — iterates over class IDs from config
- `items.go` — `FetchItems(ctx) ([]DDBItem, error)`
- `config.go` — `FetchConfig(ctx) (*DDBConfig, error)` — returns sources, conditions, actions, rules, and all lookup tables (CR, sizes, monster types)
- `client_test.go`

### Phase 4: App Integration
Wire DDB client → compendium repo → TUI data.

**New file:**
- `src/app/tui_compendium.go` — Item builders, detail body renderers (HTML→markdown), summary builders

**Modify:**
- `src/app/data_loader.go` — Add `ListCompendiumMonsters/Spells/Items/Rules/Conditions(ctx, query)` to `TUIDataLoader` interface + implementation
- `src/app/tui_dashboard.go` — Load compendium data in `buildTUIShellData()`. Add sync command handler:
  - Rules/conditions sync: fetch config (no auth) → upsert conditions + actions + rules
  - Full sync: prompt cobalt cookie → authenticate → fetch monsters/spells/items → upsert all

### Phase 5: Settings Tab & Sync UX
- Add `SettingsTabCompendium` (4th settings sub-tab under `@`)
- Lists 190 DDB source books from `compendium_sources` table, `t` toggles enabled/disabled
- `s` from this tab triggers sync: rules/conditions (no auth) + monsters/spells/items (prompts cobalt cookie)
- Auto-fetch source list from config endpoint on first visit (no auth)
- Pre-enable free sources (Basic Rules, SRD) by default

### Phase 6: Search & Polish
- Search integration (add compendium sub-tabs to global search)
- HTML→markdown converter for detail body rendering

## Database Schema (Migration 013)

```sql
CREATE TABLE compendium_sources (
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE compendium_monsters (
    id          INTEGER PRIMARY KEY,
    ddb_id      INTEGER UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    cr          TEXT NOT NULL DEFAULT '',
    type        TEXT NOT NULL DEFAULT '',
    size        TEXT NOT NULL DEFAULT '',
    hp          TEXT NOT NULL DEFAULT '',
    ac          TEXT NOT NULL DEFAULT '',
    source_name TEXT NOT NULL DEFAULT '',
    detail_json TEXT NOT NULL DEFAULT '{}',
    synced_at   TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE compendium_spells (
    id           INTEGER PRIMARY KEY,
    ddb_id       INTEGER UNIQUE NOT NULL,
    name         TEXT NOT NULL,
    level        INTEGER NOT NULL DEFAULT 0,
    school       TEXT NOT NULL DEFAULT '',
    casting_time TEXT NOT NULL DEFAULT '',
    range        TEXT NOT NULL DEFAULT '',
    components   TEXT NOT NULL DEFAULT '',
    duration     TEXT NOT NULL DEFAULT '',
    classes      TEXT NOT NULL DEFAULT '',
    source_name  TEXT NOT NULL DEFAULT '',
    detail_json  TEXT NOT NULL DEFAULT '{}',
    synced_at    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE compendium_items (
    id          INTEGER PRIMARY KEY,
    ddb_id      INTEGER UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL DEFAULT '',
    rarity      TEXT NOT NULL DEFAULT '',
    attunement  INTEGER NOT NULL DEFAULT 0,
    source_name TEXT NOT NULL DEFAULT '',
    detail_json TEXT NOT NULL DEFAULT '{}',
    synced_at   TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE compendium_rules (
    id          INTEGER PRIMARY KEY,
    ddb_id      INTEGER UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    category    TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    synced_at   TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE compendium_conditions (
    id          INTEGER PRIMARY KEY,
    ddb_id      INTEGER UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    synced_at   TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- FTS5 indexes (+ insert/update/delete triggers for each)
CREATE VIRTUAL TABLE compendium_monsters_fts USING fts5(name, type, source_name, content='compendium_monsters', content_rowid='id');
CREATE VIRTUAL TABLE compendium_spells_fts USING fts5(name, school, classes, source_name, content='compendium_spells', content_rowid='id');
CREATE VIRTUAL TABLE compendium_items_fts USING fts5(name, type, rarity, source_name, content='compendium_items', content_rowid='id');
CREATE VIRTUAL TABLE compendium_rules_fts USING fts5(name, category, content='compendium_rules', content_rowid='id');
CREATE VIRTUAL TABLE compendium_conditions_fts USING fts5(name, content='compendium_conditions', content_rowid='id');
```

Note: Rules and conditions store `description` directly (text/HTML from config) rather than `detail_json` since the config data is simple and complete. Monsters/spells/items use `detail_json` to preserve the rich DDB response for building detailed views.

## List View Columns

| Sub-tab | List row format | Detail body |
|---------|----------------|-------------|
| Monsters | `CR  Type  Name` | Stats block, abilities, actions, description (from detail_json) |
| Spells | `Lvl  School  Name` | Casting time, range, components, duration, concentration, ritual, description |
| Items | `Rarity  Type  Name` | Attunement, properties, tags, description |
| Rules | `Category  Name` | Full rule text (basic actions + rules + weapon properties merged) |
| Conditions | `Name` | Full condition effects description |

## Verification
- Phase 1: Tab appears, sub-tabs cycle with `h`/`l`, keyboard shortcut `8` works, empty state renders. `go test ./src/render/`
- Phase 2: `go test ./src/ledger/compendium/` passes, data can be inserted/queried/searched via FTS
- Phase 3: `go test ./src/ddb/` passes with recorded HTTP responses
- Phase 4: Config sync populates conditions/rules without cookie. Full sync with cookie populates monsters/spells/items. Data appears in Compendium tab, search works.
- Full suite: `go test ./...` passes at each phase
