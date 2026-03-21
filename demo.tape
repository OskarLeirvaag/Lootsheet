# LootSheet TUI demo — quick walkthrough of a filled campaign database.
# Run: go build -o ./lootsheet . && vhs demo.tape
#
# Safety: uses a temporary database via LOOTSHEET_DATABASE_PATH.
# Does not touch any existing local or remote database.

Output docs/demo.gif

Set Shell zsh
Set FontSize 14
Set Width 1400
Set Height 800
Set Padding 10
Set Framerate 30
Set TypingSpeed 30ms

# Build and seed a temporary demo database before recording.
Hide
Type "DB=$(mktemp /tmp/lootsheet-demo-XXXX.db) && export LOOTSHEET_DATABASE_PATH=$DB LOOTSHEET_NO_ANIMATION=1 && ./lootsheet init 2>/dev/null && sqlite3 $DB < fixtures/sample_campaign.sql && clear"
Enter
Sleep 3s

# Open the TUI — "Local" is pre-selected in the startup picker.
Type "./lootsheet tui"
Sleep 1s
Show
Sleep 1s
Enter
Sleep 500ms
Enter
Sleep 500ms
Enter

# ── Dashboard ──
Sleep 2s

# ── Journal ── browse posted entries and inspect detail.
Type "2"
Sleep 1s
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms

# ── Quests ── browse quest register.
Type "3"
Sleep 1s
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms

# ── Loot ── browse loot register.
Type "4"
Sleep 1s
Down
Sleep 500ms
Down
Sleep 500ms

# Sell a recognized loot item.
Type "s"
Sleep 600ms
Type "9GP"
Sleep 500ms
Enter
Sleep 1s

# ── Assets ── browse existing assets, then add a new one.
Type "5"
Sleep 1s
Down
Sleep 500ms

Type "a"
Sleep 600ms
Type "Dragonscale Shield"
Tab
Sleep 300ms
Type "Dragon Hoard"
Tab
Sleep 300ms
Type "1"
Tab
Sleep 300ms
Type "Ragnar"
Tab
Sleep 300ms
Type "Slightly singed on the left rim"
Sleep 500ms
Enter
Sleep 1s

# ── Codex ── browse NPCs, players, and settlements.
Type "6"
Sleep 1s
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms

# ── Notes ── browse session notes with markdown rendering.
Type "7"
Sleep 1.5s
Down
Sleep 800ms
Down
Sleep 800ms
Down
Sleep 800ms

# ── Search ── find entries across all sections.
Type "/"
Sleep 600ms
Type "ragnar"
Sleep 1s
Down
Sleep 400ms
Down
Sleep 400ms
Escape
Sleep 600ms

# ── Ledger ── open the full-screen general ledger overlay.
Type "#"
Sleep 1.5s
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
# Cycle filter to Assets.
Right
Sleep 800ms
# Cycle to Equity.
Right
Sleep 500ms
Right
Sleep 800ms
# Back to All.
Left
Sleep 400ms
Left
Sleep 400ms
Left
Sleep 600ms
# Drill into Party Cash.
Enter
Sleep 1.2s
# Back to list.
Escape
Sleep 600ms
# Close ledger overlay.
Escape
Sleep 600ms

# ── Settings ── manage accounts, codex types, campaigns.
Type "@"
Sleep 1s
Right
Sleep 500ms
Right
Sleep 500ms
Escape
Sleep 600ms

# ── Dashboard ── finish on the overview.
Type "1"
Sleep 2s

# Quit.
Type "q"
Sleep 500ms
Enter
Sleep 500ms
