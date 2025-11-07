# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Warscry API** - A RESTful API providing queryable access to Warhammer: Age of Sigmar Warcry game data. This is an unofficial fan project that serves fighter and ability data through HTTP endpoints with advanced filtering capabilities including comparison operators.

- **Language**: Go 1.21.9 (pure standard library, zero dependencies)
- **Data Source**: External JSON from https://krisling049.github.io/warcry_data/
- **Deployment**: Google Cloud App Engine (https://warscry.nw.r.appspot.com)
- **Architecture**: Stateless REST API with in-memory data store

## Essential Commands

### Local Development
```bash
# Run locally (default port 4424)
go run ./cmd/main.go

# Run with custom port
WARSCRY_PORT=8080 go run ./cmd/main.go

# Build executable
go build -o cmd/build/server.exe ./cmd

# Build for Linux (App Engine deployment)
GOOS=linux GOARCH=amd64 go build -o cmd/build/server ./cmd
```

### Deployment
```bash
# Deploy to Google Cloud App Engine
gcloud app deploy app.yaml --project warscry
```

### Testing
**No automated tests exist.** Test manually via API endpoints:
```bash
# Basic query
curl http://localhost:4424/fighters?warband=stormcast-eternals

# Operator query
curl http://localhost:4424/fighters?attacks__gte=5&wounds__gt=20
```

## Code Architecture

### Directory Structure
```
cmd/
  main.go              # Entry point: loads data, registers routes, starts server
warscry/              # Core package
  api.go              # HTTP handlers (RootHandler, FighterHandler, AbilityHandler)
  models.go           # Data structures (Fighter, Weapon, Ability, Warband) and loading logic
  fighters.go         # Fighter query matching logic
  abilities.go        # Ability query matching logic
  warbands.go         # Warband aggregation logic
app.yaml              # GCP App Engine configuration
openapi.yaml          # API specification
```

### Request Flow
1. `main()` loads all fighter/ability data from GitHub Pages into memory
2. Creates global collections: `AllFighters`, `AllAbilities`, `AllWarbands`
3. Registers HTTP handlers:
   - `/` → RootHandler (usage instructions)
   - `/fighters` → FighterHandler (concurrent query processing)
   - `/abilities` → AbilityHandler (query processing)
4. Starts HTTP server on port from `WARSCRY_PORT` env var

### Key Design Patterns

**Concurrent Processing** (warscry/api.go:219-282):
- Fighter queries use goroutines + channels for parallel filtering
- Each fighter checked in separate goroutine
- Results aggregated via channel and `sync.WaitGroup`

**Query Operators** (warscry/api.go:55-104):
- Operators (`__gt`, `__gte`, `__lt`, `__lte`) implemented as functions
- `GetOperator()` parses query parameter suffix to select operator
- Applied to numeric characteristics: points, movement, toughness, wounds, weapon stats

**Weapon Matching** (warscry/fighters.go:106):
- Weapon queries succeed if **ANY** weapon on a fighter matches (not all)
- Uses `Any(weaponConditions)` helper function

## API Query System

### Fighter Endpoint: GET `/fighters`

**String exact match**: `name`, `_id`, `warband`, `subfaction`, `grand_alliance`
**String array match**: `runemarks` (must match ALL provided runemarks)
**Numeric with operators**: `points`, `movement`, `toughness`, `wounds`
**Weapon numeric with operators**: `attacks`, `strength`, `dmg_hit`, `dmg_crit`, `min_range`, `max_range`

**Operators**:
- `__gt` - greater than
- `__gte` - greater than or equal to
- `__lt` - less than
- `__lte` - less than or equal to

Examples:
```
/fighters?attacks__gte=5                    # Fighters with 5+ attacks on any weapon
/fighters?wounds__gt=20&toughness__gte=5   # High toughness, high wound fighters
/fighters?warband=rotbringers&runemarks=hero  # Rotbringers heroes
```

### Ability Endpoint: GET `/abilities`

**Exact match**: `_id`, `name`, `warband`, `cost`, `runemarks`
**Substring search**: `description`

Examples:
```
/abilities?warband=stormcast-eternals
/abilities?description=wounds              # Find abilities mentioning "wounds"
```

## Data Models

All models defined in `warscry/models.go`:

**Fighter**: Main fighter data with characteristics (movement, toughness, wounds, points) and weapons array
**Weapon**: Weapon characteristics (range, attacks, strength, damage on hit/crit)
**Ability**: Ability data with cost, runemarks, and description
**Warband**: Aggregated collection of fighters, abilities, and battle traits for a faction

See `openapi.yaml` for complete schema references.

## Important Implementation Notes

1. **All data loaded at startup** - No database, all JSON data fetched from GitHub Pages and held in memory
2. **CORS enabled** - `Access-Control-Allow-Origin: *` on all responses
3. **Zero dependencies** - Uses only Go standard library
4. **Concurrent filtering** - Fighter queries use goroutines for performance
5. **Stateless** - No session state, authentication, or persistence

## Environment Configuration

- `WARSCRY_PORT`: Server port (default: 4424, App Engine uses 8080)
- App Engine settings in `app.yaml`: F1 instance class, max 1 instance, auto-scaling config
