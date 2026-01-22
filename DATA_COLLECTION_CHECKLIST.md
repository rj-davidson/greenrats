# Data Collection Checklist

Track progress on expanding data collection from BallDontLie GOAT tier and eventually direct PGA Tour scraping.

---

## Phase 1: Schema Foundation ✅

### Step 1a: Season Entity ✅
- [x] Create `season.go` schema (year, start_date, end_date, is_current)
- [x] Run `go generate ./ent`
- [x] Run `atlas migrate diff add_season_entity --env local`
- [x] Review migration SQL
- [ ] Run `atlas migrate apply --env local` (deploy step)
- [ ] Seed Season records for existing season_year values (deploy step)
- [x] **🔍 CHECKPOINT: lint, test, commit "Add Season entity and migration"**

### Step 1b: New Data Collection Entities ✅
- [x] Create `course.go` schema
- [x] Create `coursehole.go` schema
- [x] Create `round.go` schema (with Round→Course edge for multi-course tournaments)
- [x] Create `holescore.go` schema
- [x] Create `golferseason.go` schema
- [x] Add edges: Tournament→course, Golfer→seasons, TournamentEntry→rounds
- [x] Run `go generate ./ent`
- [x] Run `atlas migrate diff add_data_collection_entities --env local`
- [x] Review migration SQL
- [ ] Run `atlas migrate apply --env local` (deploy step)
- [x] **🔍 CHECKPOINT: lint, test, commit "Add Course, Round, HoleScore, GolferSeason entities"**

### Step 1c: Season FK Migration ✅
- [x] Add nullable season_id FK to Tournament (keep season_year)
- [x] Add nullable season_id FK to League (keep season_year)
- [x] Add nullable season_id FK to Pick (keep season_year)
- [x] Generate and apply migration
- [x] Write data migration script (`cmd/migrate-seasons`) to populate season FKs
- [ ] Run data migration script (deploy step)
- [ ] Remove season_year fields, make season FKs required (future)
- [x] **🔍 CHECKPOINT: lint, test, commit "Migrate season_year to Season FK"**

---

## Phase 2: BallDontLie GOAT Tier Client ✅

### Step 2a: Types & Interface ✅
- [x] Add Course, CourseHole response types to `types.go` (already existed)
- [x] Add PlayerRoundResult, PlayerScorecard types (already existed in types_goat_tier.go)
- [x] Add new method signatures to `interface.go`
- [x] Add mock implementations in `mock.go`
- [x] **🔍 CHECKPOINT: lint, test, commit "Add BallDontLie GOAT tier types and interfaces"**

### Step 2b: Implement Client Methods ✅
- [x] Implement `GetCourses()` → /pga/v1/courses
- [x] Implement `GetCourseHoles(courseID)` → /pga/v1/course_holes
- [x] Implement `GetPlayerRoundResults(tournamentID)` → /pga/v1/player_round_results
- [x] Implement `GetPlayerScorecards(tournamentID)` → /pga/v1/player_scorecards
- [x] Implement `GetPlayerSeasonStats(season, statIDs)` → /pga/v1/player_season_stats
- [ ] Write unit tests for new methods (API tests deferred)
- [x] **🔍 CHECKPOINT: lint, test, commit "Implement BallDontLie GOAT tier client methods"**

---

## Phase 3: Ingest Service Jobs (Partial)

### Step 3a: Sync Service Methods ✅
- [x] Add sync service methods for upserting courses/holes (UpsertCourse, UpsertCourseHole)
- [x] Add sync service methods for upserting rounds/hole scores (UpsertRound, UpsertHoleScore)
- [x] Add sync service methods for upserting golfer season stats (UpsertGolferSeasonStat)
- [x] **🔍 CHECKPOINT: lint, test, commit "Add sync methods for courses, rounds, hole scores, and golfer stats"**

### Step 3b: Course Sync Job
- [ ] Add CourseSync job to ingest main.go (weekly)
- [ ] Wire up BallDontLie client to sync service
- [ ] Test locally
- [ ] **🔍 CHECKPOINT: lint, test, commit "Add course sync job"**

### Step 3c: Scorecard Sync Job
- [ ] Add ScorecardSync job (during active tournaments, hourly)
- [ ] Fetch player round results → create Round records
- [ ] Fetch player scorecards → create HoleScore records
- [ ] Handle partial rounds (player still on course)
- [ ] Test with active tournament
- [ ] **🔍 CHECKPOINT: lint, test, commit "Add scorecard sync job"**

### Step 3d: Season Stats Sync Job
- [ ] Identify BallDontLie stat IDs for our curated stats
- [ ] Add SeasonStatsSync job (weekly)
- [ ] Test with 2025-26 season
- [ ] **🔍 CHECKPOINT: lint, test, commit "Add season stats sync job"**

### Step 3e: Backfill Current Season
- [ ] Run course sync to populate courses
- [ ] Backfill 2024-25 season rounds and scorecards
- [ ] Backfill 2024-25 season stats
- [ ] **🔍 CHECKPOINT: verify data, commit "Backfill 2024-25 season data"**

---

## Phase 4: PGA Tour Direct Scraping (Future)

### GraphQL Discovery
- [ ] Document available PGA Tour GraphQL queries
- [ ] Map BallDontLie data to PGA Tour equivalents
- [ ] Identify data only available via PGA Tour

### Client Implementation
- [ ] Extend pgatour client with new queries
- [ ] Add types for leaderboard/scorecard data
- [ ] Add types for player stats

### Migration Strategy
- [ ] Dual-source testing (compare BallDontLie vs PGA Tour)
- [ ] Feature flag to switch data sources
- [ ] Remove BallDontLie dependency

---

## Data Source Mapping

| Data Type | BallDontLie Endpoint | PGA Tour GraphQL | Priority |
|-----------|---------------------|------------------|----------|
| Tournaments | /tournaments | schedule query | ✅ Have |
| Players | /players | (via field) | ✅ Have |
| Tournament Results | /tournament_results | leaderboardV3 | ✅ Have |
| Field | /tournament_field | field query | ✅ Have |
| Round Scores | /player_round_results | leaderboardV3 | 🎯 Phase 2 |
| Scorecards | /player_scorecards | scorecardV3 | 🎯 Phase 2 |
| Season Stats | /player_season_stats | playerStats | 🎯 Phase 2 |
| Courses | /courses | (limited) | 🎯 Phase 2 |
| Course Holes | /course_holes | (limited) | 🎯 Phase 2 |

---

## Stat IDs Reference (BallDontLie)

We need to identify the specific stat_ids for our curated stats:
- [ ] Scoring average
- [ ] Top 10 finishes
- [ ] Cuts made
- [ ] Events played
- [ ] Wins
- [ ] Earnings
- [ ] Driving distance
- [ ] Driving accuracy
- [ ] Greens in regulation
- [ ] Putting average
- [ ] Scrambling percentage

---

## Notes

- BallDontLie uses cursor-based pagination (max 100 per page)
- Rate limit: 2 requests/second with burst of 5
- Tournament IDs ≤42 are main PGA Tour events
- PGA Tour uses composite IDs like "2026002" (season + id) but we use clean FKs
- **UI changes deferred**: No frontend updates until design spec is finalized. Backend schema supports all new data; UI will be addressed separately.
- Not all tournaments use the same course for every round - Round→Course edge allows linking each round to its specific course
- Can use BallDontLie for tournament fields since we have access (don't need PGA Tour scraping for this)
