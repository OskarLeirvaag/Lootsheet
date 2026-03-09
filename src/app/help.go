package app

const rootHelpText = `LootSheet CLI

Usage:
  lootsheet COMMAND [SUBCOMMAND] [FLAGS]
  lootsheet help [COMMAND] [SUBCOMMAND]
  lootsheet COMMAND help
  lootsheet COMMAND SUBCOMMAND -h
  lootsheet COMMAND SUBCOMMAND --help

Command groups:
  db         inspect database state and run schema migrations
  init       initialize a fresh LootSheet database
  tui        open the full-screen TUI shell
  account    create, rename, activate, deactivate, delete, and inspect accounts
  entry      create guided expense, income, and custom journal entries
  journal    post and reverse balanced journal entries
  quest      track promised, earned, collected, and written-off quest rewards
  loot       track loot appraisal, recognition, and sale workflows
  report     run read-only accounting and register reports

Examples:
  lootsheet init
  lootsheet tui
  lootsheet account list
  lootsheet entry expense --account 5100 --amount 2SP5CP --description "Restock arrows"
  lootsheet journal post --date 2026-03-08 --description "Restock arrows" --debit 5100:2SP5CP --credit 1000:2SP5CP
  lootsheet quest create --title "Goblin Bounty" --patron "Mayor Rowan" --reward 25GP
  lootsheet report trial-balance

Help examples:
  lootsheet help
  lootsheet account help
  lootsheet account list help
  lootsheet entry expense --help
  lootsheet journal post --help
`

const dbHelpText = `LootSheet CLI

Usage:
  lootsheet db status
  lootsheet db migrate

Subcommands:
  status   inspect whether the configured SQLite database is uninitialized, current, upgradeable, foreign, or damaged
  migrate  apply pending embedded schema migrations to an existing LootSheet database

Examples:
  lootsheet db status
  lootsheet db migrate
`

const dbStatusHelpText = `LootSheet CLI

Usage:
  lootsheet db status

Displays:
  Database path
  Existence and lifecycle state
  Detail for foreign or damaged databases
  Current and target schema versions
  Applied and pending migrations
`

const dbMigrateHelpText = `LootSheet CLI

Usage:
  lootsheet db migrate

Applies pending schema migrations to the configured LootSheet database.
Foreign and damaged databases are reported and left untouched.

Examples:
  lootsheet db migrate
`

const initHelpText = `LootSheet CLI

Usage:
  lootsheet init

Bootstraps a fresh SQLite database from the embedded schema and seed accounts.
If the configured database already contains LootSheet metadata, init reports that it is already initialized and does not reseed it.
`

const tuiHelpText = `LootSheet CLI

Usage:
  lootsheet tui

Opens the full-screen LootSheet TUI using tcell.
The current slice is interactive: a boxed dashboard plus Accounts, Journal, Quest, and Loot screens backed by app-facing adapters. List screens keep a selected row and detail pane visible, the shell redraws cleanly on resize, the dashboard exposes guided expense, income, and custom journal-entry launchers, the Accounts screen supports add/remove/toggle actions, the Journal screen exposes guided expense/income launchers plus reversal, the Quest screen supports add/edit plus collect/write-off actions, and the Loot screen supports add/edit plus recognize/sell actions.

Keys:
  Left/Right, Tab/Shift+Tab  move between top-level sections
  1-5                        jump directly to dashboard/accounts/journal/quests/loot
  Up/Down, j/k               move the selected row on list screens
  PgUp/PgDn, Home/End        jump through longer list screens
  e                          open guided expense entry creation on the Dashboard or Journal screen
  i                          open guided income entry creation on the Dashboard or Journal screen
  a                          open guided custom entry creation on the Dashboard; add accounts, quests, or loot on their screens
  u                          edit the selected quest or loot item on its screen
  d                          remove the selected account on the Accounts screen
  t                          toggle the selected account active/inactive on the Accounts screen
  r                          reverse the selected posted journal entry on the Journal screen
  c                          collect the full outstanding balance for the selected quest on the Quest screen
  w                          write off the full outstanding balance for the selected quest on the Quest screen
  n                          recognize the selected latest loot appraisal on the Loot screen
  s                          sell the selected recognized loot item on the Loot screen
  Loot create/edit           does not set value; appraisal happens later as a separate workflow
  ?                          open a glossary modal for the current screen's accounting terms
  Enter                      confirm the open modal
  Enter                      submit the guided entry composer when a guided entry form is open
  Esc                        cancel the open modal, or quit when no modal is open
  q                          cancel the open modal, or quit when no modal is open
  Ctrl+L                     reload data and force a full redraw
`

const entryHelpText = `LootSheet CLI

Usage:
  lootsheet entry expense --account CODE --amount AMOUNT --description TEXT [--paid-from CODE] [--date YYYY-MM-DD] [--memo TEXT]
  lootsheet entry income --account CODE --amount AMOUNT --description TEXT [--deposit-to CODE] [--date YYYY-MM-DD] [--memo TEXT]
  lootsheet entry custom --description TEXT [--date YYYY-MM-DD] --debit CODE:AMOUNT[:MEMO] --credit CODE:AMOUNT[:MEMO]

Subcommands:
  expense  record a guided two-line expense entry
  income   record a guided two-line income entry
  custom   record a guided multi-line journal entry
`

const entryExpenseHelpText = `LootSheet CLI

Usage:
  lootsheet entry expense --account CODE --amount AMOUNT --description TEXT [--paid-from CODE] [--date YYYY-MM-DD] [--memo TEXT]

Defaults:
  --paid-from  1000
  --date       today

Rules:
  --account must be an active expense account
  --paid-from must be an active asset or liability account

Examples:
  lootsheet entry expense --account 5100 --amount 2SP5CP --description "Restock arrows"
  lootsheet entry expense --account 5300 --paid-from 2100 --amount 8GP --description "Inn charged to tab"
`

const entryIncomeHelpText = `LootSheet CLI

Usage:
  lootsheet entry income --account CODE --amount AMOUNT --description TEXT [--deposit-to CODE] [--date YYYY-MM-DD] [--memo TEXT]

Defaults:
  --deposit-to  1000
  --date        today

Rules:
  --account must be an active income account
  --deposit-to must be an active asset account

Examples:
  lootsheet entry income --account 4000 --amount 25GP --description "Goblin bounty"
  lootsheet entry income --account 4100 --deposit-to 1100 --amount 5GP --description "Bonus receivable"
`

const entryCustomHelpText = `LootSheet CLI

Usage:
  lootsheet entry custom --description TEXT [--date YYYY-MM-DD] --debit CODE:AMOUNT[:MEMO] --credit CODE:AMOUNT[:MEMO]

Defaults:
  --date  today

Amounts accept D&D 5e denominations: PP, GP, EP, SP, CP (case insensitive).

Examples:
  lootsheet entry custom --description "Gear transfer" --debit 1300:5GP --credit 1000:5GP
  lootsheet entry custom --description "Split payout" --debit 1000:10GP --credit 4000:8GP --credit 4100:2GP
`

const accountHelpText = `LootSheet CLI

Usage:
  lootsheet account list
  lootsheet account create --code CODE --name NAME --type TYPE
  lootsheet account rename --code CODE --name NAME
  lootsheet account deactivate --code CODE
  lootsheet account activate --code CODE
  lootsheet account delete --code CODE
  lootsheet account ledger --code CODE

Subcommands:
  list        show the chart of accounts
  create      add a new account to the chart
  rename      change an account name without changing its immutable code
  deactivate  mark an account inactive without deleting history
  activate    reactivate an inactive account
  delete      remove an unused account that has no postings
  ledger      print the posting history for a single account
`

const accountListHelpText = `LootSheet CLI

Usage:
  lootsheet account list

Shows account code, type, active state, and name for the current chart of accounts.
`

const accountCreateHelpText = `LootSheet CLI

Usage:
  lootsheet account create --code CODE --name NAME --type TYPE

Types:
  asset
  liability
  equity
  income
  expense

Examples:
  lootsheet account create --code 5120 --name "Wizard Magic Ink" --type expense
  lootsheet account create --code 1210 --name "Dragon Bond Deposits" --type asset
`

const accountRenameHelpText = `LootSheet CLI

Usage:
  lootsheet account rename --code CODE --name NAME

Example:
  lootsheet account rename --code 5120 --name "Wizard Ink & Paper"
`

const accountDeactivateHelpText = `LootSheet CLI

Usage:
  lootsheet account deactivate --code CODE

Example:
  lootsheet account deactivate --code 5120
`

const accountActivateHelpText = `LootSheet CLI

Usage:
  lootsheet account activate --code CODE

Example:
  lootsheet account activate --code 5120
`

const accountDeleteHelpText = `LootSheet CLI

Usage:
  lootsheet account delete --code CODE

Deletes an account only if it has no journal postings.

Example:
  lootsheet account delete --code 5120
`

const accountLedgerHelpText = `LootSheet CLI

Usage:
  lootsheet account ledger --code CODE

Examples:
  lootsheet account ledger --code 1000
  lootsheet account ledger --code 5100
`

const journalHelpText = `LootSheet CLI

Usage:
  lootsheet journal post --date YYYY-MM-DD --description TEXT --debit CODE:AMOUNT[:MEMO] --credit CODE:AMOUNT[:MEMO]
  lootsheet journal reverse --entry-id UUID --date YYYY-MM-DD [--description TEXT]

Subcommands:
  post     create a new balanced posted journal entry
  reverse  create an immutable reversing entry for a posted journal entry
`

const journalPostHelpText = `LootSheet CLI

Usage:
  lootsheet journal post --date YYYY-MM-DD --description TEXT --debit CODE:AMOUNT[:MEMO] --credit CODE:AMOUNT[:MEMO]

Amounts accept D&D 5e denominations: PP, GP, EP, SP, CP (case insensitive).
  Mixed:   2GP5SP, 1PP 2GP 3SP 5CP
  Decimal: 5.5GP, 0.5SP
  Bare integer (treated as CP): 100

Examples:
  lootsheet journal post --date 2026-03-08 --description "Restock arrows" --debit 5100:2SP5CP:Quiver refill --credit 1000:2SP5CP
  lootsheet journal post --date 2026-03-08 --description "Quest reward earned" --debit 1100:1GP --credit 4000:1GP
`

const journalReverseHelpText = `LootSheet CLI

Usage:
  lootsheet journal reverse --entry-id UUID --date YYYY-MM-DD [--description TEXT]

Examples:
  lootsheet journal reverse --entry-id abc-123 --date 2026-03-09
  lootsheet journal reverse --entry-id abc-123 --date 2026-03-09 --description "Correcting duplicate entry"
`

const questHelpText = "LootSheet CLI\n\nUsage:\n  lootsheet quest create --title TEXT [--patron TEXT] [--description TEXT] [--reward AMOUNT] [--advance AMOUNT] [--bonus TEXT] [--status offered|accepted] [--accepted-on DATE]\n  lootsheet quest list\n  lootsheet quest accept --id ID --date YYYY-MM-DD\n  lootsheet quest complete --id ID --date YYYY-MM-DD\n  lootsheet quest collect --id ID --amount AMOUNT --date YYYY-MM-DD [--description TEXT]\n  lootsheet quest writeoff --id ID --date YYYY-MM-DD [--description TEXT]\n\nSubcommands:\n  create    register a promised quest reward off-ledger\n  list      show current quests and statuses\n  accept    move an offered quest to accepted\n  complete  recognize a quest as earned\n  collect   cash against an earned quest reward\n  writeoff  write off an uncollectible earned quest"

const questCreateHelpText = `LootSheet CLI

Usage:
  lootsheet quest create --title TEXT [--patron TEXT] [--description TEXT] [--reward AMOUNT] [--advance AMOUNT] [--bonus TEXT] [--status offered|accepted] [--accepted-on DATE]

Amounts accept D&D denominations such as 25GP, 7SP5CP, or bare CP integers.

Examples:
  lootsheet quest create --title "Goblin Bounty" --patron "Mayor Rowan" --reward 25GP
  lootsheet quest create --title "Escort the Caravan" --reward 15GP --advance 5GP --bonus "Extra 2GP if all wagons arrive intact"
`

const questListHelpText = `LootSheet CLI

Usage:
  lootsheet quest list

Shows quest status, promised reward, and title for the quest register.
`

const questAcceptHelpText = `LootSheet CLI

Usage:
  lootsheet quest accept --id ID --date YYYY-MM-DD

Example:
  lootsheet quest accept --id quest-123 --date 2026-03-08
`

const questCompleteHelpText = `LootSheet CLI

Usage:
  lootsheet quest complete --id ID --date YYYY-MM-DD

Example:
  lootsheet quest complete --id quest-123 --date 2026-03-12
`

const questCollectHelpText = `LootSheet CLI

Usage:
  lootsheet quest collect --id ID --amount AMOUNT --date YYYY-MM-DD [--description TEXT]

Examples:
  lootsheet quest collect --id quest-123 --amount 10GP --date 2026-03-15
  lootsheet quest collect --id quest-123 --amount 5GP --date 2026-03-16 --description "Second pouch from the mayor"
`

const questWriteoffHelpText = `LootSheet CLI

Usage:
  lootsheet quest writeoff --id ID --date YYYY-MM-DD [--description TEXT]

Example:
  lootsheet quest writeoff --id quest-123 --date 2026-04-20 --description "Patron vanished into the Feywild"
`

const lootHelpText = `LootSheet CLI

Usage:
  lootsheet loot create --name TEXT [--source TEXT] [--quantity N] [--holder TEXT] [--notes TEXT]
  lootsheet loot list
  lootsheet loot appraise --id ID --value AMOUNT --date YYYY-MM-DD [--appraiser TEXT] [--notes TEXT]
  lootsheet loot recognize --appraisal-id ID --date YYYY-MM-DD [--description TEXT]
  lootsheet loot sell --id ID --amount AMOUNT --date YYYY-MM-DD [--description TEXT]

Subcommands:
  create     register found loot off-ledger
  list       show tracked loot items
  appraise   record an appraisal without recognizing it on-ledger
  recognize  bring an appraisal onto the ledger
  sell       record a sale and any gain or loss
`

const lootCreateHelpText = `LootSheet CLI

Usage:
  lootsheet loot create --name TEXT [--source TEXT] [--quantity N] [--holder TEXT] [--notes TEXT]

Examples:
  lootsheet loot create --name "Moonstone" --source "Goblin cave" --quantity 3
  lootsheet loot create --name "Silver chalice" --holder "Brom" --notes "Wrapped in old velvet"
`

const lootListHelpText = `LootSheet CLI

Usage:
  lootsheet loot list

Shows tracked loot status, quantity, and name.
`

const lootAppraiseHelpText = `LootSheet CLI

Usage:
  lootsheet loot appraise --id ID --value AMOUNT --date YYYY-MM-DD [--appraiser TEXT] [--notes TEXT]

Examples:
  lootsheet loot appraise --id loot-123 --value 75GP --date 2026-03-08 --appraiser "Guild jeweler"
  lootsheet loot appraise --id loot-123 --value 80GP --date 2026-03-09 --notes "Second opinion"
`

const lootRecognizeHelpText = `LootSheet CLI

Usage:
  lootsheet loot recognize --appraisal-id ID --date YYYY-MM-DD [--description TEXT]

Examples:
  lootsheet loot recognize --appraisal-id appraisal-123 --date 2026-03-10
  lootsheet loot recognize --appraisal-id appraisal-123 --date 2026-03-10 --description "Recognize moonstone inventory"
`

const lootSellHelpText = `LootSheet CLI

Usage:
  lootsheet loot sell --id ID --amount AMOUNT --date YYYY-MM-DD [--description TEXT]

Examples:
  lootsheet loot sell --id loot-123 --amount 55GP --date 2026-03-12
  lootsheet loot sell --id loot-123 --amount 50GP --date 2026-03-12 --description "Sold to dockside broker"
`

const reportHelpText = `LootSheet CLI

Usage:
  lootsheet report trial-balance
  lootsheet report quest-receivables
  lootsheet report promised-quests
  lootsheet report loot-summary
  lootsheet report writeoff-candidates [--as-of YYYY-MM-DD] [--min-age-days N]

Subcommands:
  trial-balance         show debits, credits, balances, and overall balancing status
  quest-receivables     show earned quest rewards that are not fully collected
  promised-quests       show offered or accepted quests that remain off-ledger
  loot-summary          show held or recognized loot with appraisal visibility
  writeoff-candidates   show older completed quests with remaining uncollected balances
`

const reportTrialBalanceHelpText = `LootSheet CLI

Usage:
  lootsheet report trial-balance

Prints the trial balance for the current ledger, including total debits and credits and whether the books are balanced.
`

const reportQuestReceivablesHelpText = `LootSheet CLI

Usage:
  lootsheet report quest-receivables

Shows promised, paid, and outstanding amounts for earned quest rewards that still have an unpaid balance.
`

const reportPromisedQuestsHelpText = `LootSheet CLI

Usage:
  lootsheet report promised-quests

Shows offered and accepted quests with promised reward, advance, and bonus terms before recognition.
`

const reportLootSummaryHelpText = `LootSheet CLI

Usage:
  lootsheet report loot-summary

Shows held and recognized loot items with quantity and latest appraisal value where available.
`

const reportWriteoffCandidatesHelpText = `LootSheet CLI

Usage:
  lootsheet report writeoff-candidates [--as-of YYYY-MM-DD] [--min-age-days N]

Defaults:
  --as-of         today
  --min-age-days  30

Examples:
  lootsheet report writeoff-candidates
  lootsheet report writeoff-candidates --as-of 2026-04-01 --min-age-days 45
`
