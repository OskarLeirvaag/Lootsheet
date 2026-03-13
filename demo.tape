# LootSheet TUI demo — quick walkthrough of a filled campaign database.
# Run: go build -o ./lootsheet . && vhs demo.tape

Output docs/demo.gif

Set Shell zsh
Set FontSize 14
Set Width 1400
Set Height 800
Set Padding 10
Set Framerate 30
Set TypingSpeed 20ms

# Seed a temporary demo database (hidden from recording).
Hide
Type "export LOOTSHEET_DATABASE_PATH=$(mktemp /tmp/lootsheet-demo-XXXX.db)"
Enter
Sleep 300ms
Type "./lootsheet init 2>/dev/null"
Enter
Sleep 2s
Type "sqlite3 $LOOTSHEET_DATABASE_PATH < fixtures/sample_campaign.sql"
Enter
Sleep 500ms
Type "clear"
Enter
Sleep 300ms
Show

# Open the TUI.
Type "./lootsheet tui"
Enter
Sleep 3s

# ── Dashboard ──
Sleep 1.5s

# ── Journal ── browse posted entries.
Type "2"
Sleep 800ms
Down
Sleep 150ms
Down
Sleep 150ms
Down
Sleep 150ms
Down
Sleep 150ms
Up
Sleep 150ms
Up
Sleep 150ms

# ── Quests ── browse quest register.
Type "3"
Sleep 800ms
Down
Sleep 150ms
Down
Sleep 150ms

# ── Loot ── browse loot register.
Type "4"
Sleep 800ms
# Wyvern Tooth Necklace is selected (held, first in sort).
# Move to Emerald Idol (recognized, second item).
Down
Sleep 500ms

# Sell the Emerald Idol.
Type "s"
Sleep 500ms
Type "9GP"
Sleep 300ms
Enter
Sleep 800ms

# ── Assets ── add a new asset.
Type "5"
Sleep 800ms
Type "a"
Sleep 500ms
# Name.
Type "Dragonscale Shield"
Tab
Sleep 150ms
# Source.
Type "Dragon Hoard"
Tab
Sleep 150ms
# Quantity.
Type "1"
Tab
Sleep 150ms
# Holder.
Type "Ragnar"
Tab
Sleep 150ms
# Notes.
Type "Slightly singed on the left rim"
Sleep 300ms
Enter
Sleep 800ms

# Give the new asset a template.
# The new asset should be selected after creation.
Type "f"
Sleep 500ms
# Template starts with 2 default lines: L1 debit, L2 credit.
# L1 Side is already debit — Tab to account.
Tab
Sleep 150ms
# L1 Account code.
Type "5100"
Tab
Sleep 150ms
# L1 Amount.
Type "5GP"
Tab
Sleep 150ms
# L2 Side is already credit — Tab to account.
Tab
Sleep 150ms
# L2 Account code.
Type "1000"
Tab
Sleep 150ms
# L2 Amount.
Type "5GP"
Sleep 400ms
Enter
Sleep 800ms

# Execute the template — creates a journal entry.
Type "x"
Sleep 500ms
# Date and description are pre-filled. Just submit.
Sleep 400ms
Enter
Sleep 800ms

# ── Codex ── browse NPCs and contacts.
Type "6"
Sleep 800ms
Down
Sleep 300ms
Down
Sleep 300ms
Up
Sleep 300ms

# ── Notes ── browse campaign notes.
Type "7"
Sleep 800ms
Down
Sleep 300ms
Down
Sleep 300ms

# ── Search ── find entries across all sections.
Type "/"
Sleep 500ms
Type "moonlit"
Sleep 800ms
# Browse results.
Down
Sleep 300ms
Down
Sleep 300ms
# Close search.
Escape
Sleep 500ms

# ── Journal ── see the new entries at the bottom.
Type "2"
Sleep 800ms
# Scroll to bottom to see sale + custom entries.
Down
Sleep 150ms
Down
Sleep 150ms
Down
Sleep 150ms
Down
Sleep 150ms
Down
Sleep 150ms
Down
Sleep 150ms
Down
Sleep 150ms
Down
Sleep 150ms
Down
Sleep 150ms
Sleep 500ms

# ── Dashboard ── finish on the overview.
Type "1"
Sleep 1.5s

# Quit.
Type "q"
Sleep 300ms
