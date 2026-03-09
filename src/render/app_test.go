package render

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

type scriptedScreen struct {
	tcell.SimulationScreen
	onInit         func(tcell.SimulationScreen)
	afterFirstShow func(tcell.SimulationScreen)
	showCount      int
	finished       bool
	lastFrame      string
}

func (s *scriptedScreen) Init() error {
	if err := s.SimulationScreen.Init(); err != nil {
		return err
	}
	if s.onInit != nil {
		s.onInit(s.SimulationScreen)
	}
	return nil
}

func (s *scriptedScreen) Show() {
	s.SimulationScreen.Show()
	s.lastFrame = simulationPlainText(s.SimulationScreen)
	s.showCount++
	if s.showCount == 1 && s.afterFirstShow != nil {
		s.afterFirstShow(s.SimulationScreen)
	}
}

func (s *scriptedScreen) Sync() {
	s.SimulationScreen.Sync()
	s.lastFrame = simulationPlainText(s.SimulationScreen)
}

func (s *scriptedScreen) Fini() {
	s.finished = true
	s.SimulationScreen.Fini()
}

func TestRunExitsCleanlyOnQuitKey(t *testing.T) {
	screen := &scriptedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		onInit: func(sim tcell.SimulationScreen) {
			sim.SetSize(80, 24)
		},
		afterFirstShow: func(sim tcell.SimulationScreen) {
			sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		},
	}

	err := Run(context.Background(), &Options{
		ScreenFactory: func() (Screen, error) {
			return screen, nil
		},
	})
	if err != nil {
		t.Fatalf("run render app: %v", err)
	}
	if !screen.finished {
		t.Fatal("expected screen to be finalized")
	}

	cells, width, height := screen.GetContents()
	if width != 0 || height != 0 {
		t.Fatalf("finalized simulation size = %dx%d, want 0x0", width, height)
	}
	if len(cells) != 0 {
		t.Fatalf("finalized simulation still has %d cells", len(cells))
	}
}

func TestRunHandlesResizeAndRedrawBeforeExit(t *testing.T) {
	screen := &scriptedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		onInit: func(sim tcell.SimulationScreen) {
			sim.SetSize(60, 20)
		},
		afterFirstShow: func(sim tcell.SimulationScreen) {
			sim.SetSize(72, 24)
			_ = sim.PostEvent(tcell.NewEventResize(72, 24))
			sim.InjectKey(tcell.KeyEsc, 0, tcell.ModNone)
		},
	}

	if err := Run(context.Background(), &Options{
		ScreenFactory: func() (Screen, error) {
			return screen, nil
		},
	}); err != nil {
		t.Fatalf("run render app: %v", err)
	}

	plain := screen.lastFrame
	for _, token := range []string{"LootSheet TUI", "Section: Dashboard", "q quit", "Ctrl+L refresh"} {
		if !strings.Contains(plain, token) {
			t.Fatalf("simulation output missing %q:\n%s", token, plain)
		}
	}
}

func TestRunSwitchesSectionsBeforeExit(t *testing.T) {
	screen := &scriptedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		onInit: func(sim tcell.SimulationScreen) {
			sim.SetSize(90, 24)
		},
		afterFirstShow: func(sim tcell.SimulationScreen) {
			sim.InjectKey(tcell.KeyRight, 0, tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		},
	}

	if err := Run(context.Background(), &Options{
		ScreenFactory: func() (Screen, error) {
			return screen, nil
		},
	}); err != nil {
		t.Fatalf("run render app: %v", err)
	}

	if !strings.Contains(screen.lastFrame, "Section: Accounts") {
		t.Fatalf("simulation output missing accounts section:\n%s", screen.lastFrame)
	}
}

func TestRunDispatchesCommandAndShowsSuccessStatus(t *testing.T) {
	screen := &scriptedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		onInit: func(sim tcell.SimulationScreen) {
			sim.SetSize(120, 40)
		},
		afterFirstShow: func(sim tcell.SimulationScreen) {
			sim.InjectKey(tcell.KeyRune, '2', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 't', tcell.ModNone)
			sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		},
	}

	initial := testAccountsShellData(true)
	updated := testAccountsShellData(false)
	var got Command
	var called bool

	if err := Run(context.Background(), &Options{
		ScreenFactory: func() (Screen, error) {
			return screen, nil
		},
		ShellLoader: func(context.Context) (ShellData, error) {
			return initial, nil
		},
		CommandHandler: func(_ context.Context, command Command) (CommandResult, error) {
			called = true
			got = command
			return CommandResult{
				Data: updated,
				Status: StatusMessage{
					Level: StatusSuccess,
					Text:  "Account 1000 deactivated.",
				},
			}, nil
		},
	}); err != nil {
		t.Fatalf("run render app: %v", err)
	}

	if !called {
		t.Fatal("expected command handler to be called")
	}
	if got.ID != "account.deactivate" {
		t.Fatalf("command id = %q, want account.deactivate", got.ID)
	}
	if got.Section != SectionAccounts {
		t.Fatalf("command section = %v, want accounts", got.Section)
	}
	if got.ItemKey != "1000" {
		t.Fatalf("command item key = %q, want 1000", got.ItemKey)
	}

	for _, token := range []string{
		"Account 1000 deactivated.",
		"1000 asset inactive Party Cash",
		"Status: inactive",
		"t activate",
	} {
		if !strings.Contains(screen.lastFrame, token) {
			t.Fatalf("simulation output missing %q:\n%s", token, screen.lastFrame)
		}
	}
}

func TestRunKeepsLootSaleInputModalOpenOnInputError(t *testing.T) {
	screen := &scriptedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		onInit: func(sim tcell.SimulationScreen) {
			sim.SetSize(120, 40)
		},
		afterFirstShow: func(sim tcell.SimulationScreen) {
			sim.InjectKey(tcell.KeyRune, '5', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 's', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'b', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'a', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'd', tcell.ModNone)
			sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := Run(ctx, &Options{
		ScreenFactory: func() (Screen, error) {
			return screen, nil
		},
		ShellLoader: func(context.Context) (ShellData, error) {
			return testLootSellShellData(), nil
		},
		CommandHandler: func(_ context.Context, command Command) (CommandResult, error) {
			if command.Fields["amount"] != "bad" {
				t.Fatalf("command amount = %q, want bad", command.Fields["amount"])
			}
			cancel()
			return CommandResult{}, InputError{Message: `Invalid amount "bad".`}
		},
	}); err != nil {
		t.Fatalf("run render app: %v", err)
	}

	for _, token := range []string{
		`Error: Invalid amount "bad".`,
		"Sale amount: bad",
	} {
		if !strings.Contains(screen.lastFrame, token) {
			t.Fatalf("simulation output missing %q:\n%s", token, screen.lastFrame)
		}
	}
}

func TestRunKeepsCurrentDataAndShowsErrorStatusOnCommandFailure(t *testing.T) {
	screen := &scriptedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		onInit: func(sim tcell.SimulationScreen) {
			sim.SetSize(96, 28)
		},
		afterFirstShow: func(sim tcell.SimulationScreen) {
			sim.InjectKey(tcell.KeyRune, '2', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 't', tcell.ModNone)
			sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		},
	}

	if err := Run(context.Background(), &Options{
		ScreenFactory: func() (Screen, error) {
			return screen, nil
		},
		ShellLoader: func(context.Context) (ShellData, error) {
			return testAccountsShellData(true), nil
		},
		CommandHandler: func(context.Context, Command) (CommandResult, error) {
			return CommandResult{}, errors.New("account 1000 cannot be deactivated right now")
		},
	}); err != nil {
		t.Fatalf("run render app: %v", err)
	}

	for _, token := range []string{
		"account 1000 cannot be deactivated right now",
		"1000 asset active Party Cash",
		"Status: active",
		"t deactivate",
	} {
		if !strings.Contains(screen.lastFrame, token) {
			t.Fatalf("simulation output missing %q:\n%s", token, screen.lastFrame)
		}
	}
}

func TestRunDispatchesJournalReverseAndKeepsOriginalSelection(t *testing.T) {
	screen := &scriptedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		onInit: func(sim tcell.SimulationScreen) {
			sim.SetSize(120, 40)
		},
		afterFirstShow: func(sim tcell.SimulationScreen) {
			sim.InjectKey(tcell.KeyRune, '3', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'r', tcell.ModNone)
			sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		},
	}

	initial := testJournalShellData(false)
	updated := testJournalShellData(true)
	var got Command

	if err := Run(context.Background(), &Options{
		ScreenFactory: func() (Screen, error) {
			return screen, nil
		},
		ShellLoader: func(context.Context) (ShellData, error) {
			return initial, nil
		},
		CommandHandler: func(_ context.Context, command Command) (CommandResult, error) {
			got = command
			return CommandResult{
				Data: updated,
				Status: StatusMessage{
					Level: StatusSuccess,
					Text:  "Entry #1 reversed as entry #2.",
				},
			}, nil
		},
	}); err != nil {
		t.Fatalf("run render app: %v", err)
	}

	if got.ID != "journal.reverse" {
		t.Fatalf("command id = %q, want journal.reverse", got.ID)
	}
	if got.Section != SectionJournal {
		t.Fatalf("command section = %v, want journal", got.Section)
	}
	if got.ItemKey != "entry-1" {
		t.Fatalf("command item key = %q, want entry-1", got.ItemKey)
	}

	for _, token := range []string{
		"Entry #1 reversed as entry #2.",
		"#1    2026-03-08 reversed Restock arrows",
		"Status: reversed",
	} {
		if !strings.Contains(screen.lastFrame, token) {
			t.Fatalf("simulation output missing %q:\n%s", token, screen.lastFrame)
		}
	}
}

func TestRunDispatchesLootSellAndRefreshesSelectionFallback(t *testing.T) {
	screen := &scriptedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		onInit: func(sim tcell.SimulationScreen) {
			sim.SetSize(120, 40)
		},
		afterFirstShow: func(sim tcell.SimulationScreen) {
			sim.InjectKey(tcell.KeyRune, '5', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 's', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, '8', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, ' ', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'g', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'p', tcell.ModNone)
			sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		},
	}

	initial := testLootSellShellData()
	updated := testLootAfterSaleShellData()
	var got Command

	if err := Run(context.Background(), &Options{
		ScreenFactory: func() (Screen, error) {
			return screen, nil
		},
		ShellLoader: func(context.Context) (ShellData, error) {
			return initial, nil
		},
		CommandHandler: func(_ context.Context, command Command) (CommandResult, error) {
			got = command
			return CommandResult{
				Data: updated,
				Status: StatusMessage{
					Level: StatusSuccess,
					Text:  `Sold loot item "Gold Necklace" as entry #10.`,
				},
			}, nil
		},
	}); err != nil {
		t.Fatalf("run render app: %v", err)
	}

	if got.ID != "loot.sell" {
		t.Fatalf("command id = %q, want loot.sell", got.ID)
	}
	if got.Section != SectionLoot {
		t.Fatalf("command section = %v, want loot", got.Section)
	}
	if got.ItemKey != "loot-1" {
		t.Fatalf("command item key = %q, want loot-1", got.ItemKey)
	}
	if got.Fields["amount"] != "8 gp" {
		t.Fatalf("command amount = %q, want 8 gp", got.Fields["amount"])
	}

	for _, token := range []string{
		`Sold loot item "Gold Necklace" as entry #10.`,
		"Silver Chalice",
	} {
		if !strings.Contains(screen.lastFrame, token) {
			t.Fatalf("simulation output missing %q:\n%s", token, screen.lastFrame)
		}
	}
	if strings.Contains(screen.lastFrame, "Gold Necklace (Merchant)") || strings.Contains(screen.lastFrame, `Sell "Gold Necklace"?`) {
		t.Fatalf("sold item should not remain in the post-sale loot list:\n%s", screen.lastFrame)
	}
}

func TestRunDispatchesQuestCollectAndRefreshesSelection(t *testing.T) {
	screen := &scriptedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		onInit: func(sim tcell.SimulationScreen) {
			sim.SetSize(120, 40)
		},
		afterFirstShow: func(sim tcell.SimulationScreen) {
			sim.InjectKey(tcell.KeyRune, '4', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'c', tcell.ModNone)
			sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		},
	}

	initial := testQuestShellData(false)
	updated := testQuestShellData(true)
	var got Command

	if err := Run(context.Background(), &Options{
		ScreenFactory: func() (Screen, error) {
			return screen, nil
		},
		ShellLoader: func(context.Context) (ShellData, error) {
			return initial, nil
		},
		CommandHandler: func(_ context.Context, command Command) (CommandResult, error) {
			got = command
			return CommandResult{
				Data: updated,
				Status: StatusMessage{
					Level: StatusSuccess,
					Text:  "Collected 25 GP for quest \"Goblin Bounty\" as entry #7.",
				},
			}, nil
		},
	}); err != nil {
		t.Fatalf("run render app: %v", err)
	}

	if got.ID != "quest.collect_full" {
		t.Fatalf("command id = %q, want quest.collect_full", got.ID)
	}
	if got.Section != SectionQuests {
		t.Fatalf("command section = %v, want quests", got.Section)
	}
	if got.ItemKey != "quest-1" {
		t.Fatalf("command item key = %q, want quest-1", got.ItemKey)
	}

	for _, token := range []string{
		"Collected 25 GP for quest \"Goblin Bounty\" as entry #7.",
		"Status: paid",
		"Outstanding: 0 CP",
	} {
		if !strings.Contains(screen.lastFrame, token) {
			t.Fatalf("simulation output missing %q:\n%s", token, screen.lastFrame)
		}
	}
}

func TestRunDispatchesLootRecognizeAndRefreshesSelection(t *testing.T) {
	screen := &scriptedScreen{
		SimulationScreen: tcell.NewSimulationScreen("UTF-8"),
		onInit: func(sim tcell.SimulationScreen) {
			sim.SetSize(120, 40)
		},
		afterFirstShow: func(sim tcell.SimulationScreen) {
			sim.InjectKey(tcell.KeyRune, '5', tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'n', tcell.ModNone)
			sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
			sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
		},
	}

	initial := testLootShellData(false)
	updated := testLootShellData(true)
	var got Command

	if err := Run(context.Background(), &Options{
		ScreenFactory: func() (Screen, error) {
			return screen, nil
		},
		ShellLoader: func(context.Context) (ShellData, error) {
			return initial, nil
		},
		CommandHandler: func(_ context.Context, command Command) (CommandResult, error) {
			got = command
			return CommandResult{
				Data: updated,
				Status: StatusMessage{
					Level: StatusSuccess,
					Text:  "Recognized loot item \"Gold Necklace\" as entry #9.",
				},
			}, nil
		},
	}); err != nil {
		t.Fatalf("run render app: %v", err)
	}

	if got.ID != "loot.recognize_latest" {
		t.Fatalf("command id = %q, want loot.recognize_latest", got.ID)
	}
	if got.Section != SectionLoot {
		t.Fatalf("command section = %v, want loot", got.Section)
	}
	if got.ItemKey != "loot-1" {
		t.Fatalf("command item key = %q, want loot-1", got.ItemKey)
	}

	for _, token := range []string{
		"Recognized loot item \"Gold Necklace\" as entry #9.",
		"Status: recognized",
		"Accounting state: on-ledger recognized inventory",
	} {
		if !strings.Contains(screen.lastFrame, token) {
			t.Fatalf("simulation output missing %q:\n%s", token, screen.lastFrame)
		}
	}
}

func testAccountsShellData(active bool) ShellData {
	status := "inactive"
	summary := []string{"Accounts: 1 total", "Active: 0  Inactive: 1"}
	action := &ItemActionData{
		Trigger:      ActionToggle,
		ID:           "account.activate",
		Label:        "t activate",
		ConfirmTitle: "Activate account 1000?",
		ConfirmLines: []string{"Party Cash"},
	}
	if active {
		status = "active"
		summary = []string{"Accounts: 1 total", "Active: 1  Inactive: 0"}
		action = &ItemActionData{
			Trigger:      ActionToggle,
			ID:           "account.deactivate",
			Label:        "t deactivate",
			ConfirmTitle: "Deactivate account 1000?",
			ConfirmLines: []string{"Party Cash"},
		}
	}

	return ShellData{
		Dashboard: DefaultDashboardData(),
		Accounts: ListScreenData{
			HeaderLines:  []string{"Chart of accounts from smoke.db.", "Select an account to inspect it."},
			SummaryLines: summary,
			Items: []ListItemData{
				{
					Key:         "1000",
					Row:         "1000 asset " + status + " Party Cash",
					DetailTitle: "Account 1000",
					DetailLines: []string{
						"Name: Party Cash",
						"Status: " + status,
					},
					Actions: []ItemActionData{*action},
				},
			},
		},
	}
}

func testJournalShellData(reversed bool) ShellData {
	status := "posted"
	detailLines := []string{
		"Date: 2026-03-08",
		"Status: posted",
		"Description: Restock arrows",
		"",
		"Lines:",
		"5100 Adventuring Supplies DR 25 CP (Quiver refill)",
		"1000 Party Cash CR 25 CP",
	}
	action := &ItemActionData{
		Trigger:      ActionReverse,
		ID:           "journal.reverse",
		Label:        "r reverse",
		ConfirmTitle: "Reverse entry #1?",
		ConfirmLines: []string{"Restock arrows"},
	}
	if reversed {
		status = "reversed"
		detailLines = []string{
			"Date: 2026-03-08",
			"Status: reversed",
			"Description: Restock arrows",
			"Reversed by: entry #2",
			"",
			"Lines:",
			"5100 Adventuring Supplies DR 25 CP (Quiver refill)",
			"1000 Party Cash CR 25 CP",
		}
		action = nil
	}

	return ShellData{
		Dashboard: DefaultDashboardData(),
		Journal: ListScreenData{
			HeaderLines:  []string{"Posted journal history from smoke.db.", "Select an entry to inspect it."},
			SummaryLines: []string{"Entries: 2 total", "Posted: 1", "Reversal entries: 1"},
			Items: []ListItemData{
				{
					Key:         "entry-1",
					Row:         "#1    2026-03-08 " + status + " Restock arrows",
					DetailTitle: "Entry #1",
					DetailLines: detailLines,
					Actions:     itemActions(action),
				},
			},
		},
	}
}

func testQuestShellData(collected bool) ShellData {
	status := "collectible"
	outstanding := "25 GP due"
	detailLines := []string{
		"Patron: Mayor Rowan",
		"Status: collectible",
		"Promised reward: 25 GP",
		"Outstanding: 25 GP",
		"Collected so far: 0 CP",
		"Accounting state: collectible but unpaid",
	}
	actions := []ItemActionData{
		{
			Trigger:      ActionCollect,
			ID:           "quest.collect_full",
			Label:        "c collect",
			ConfirmTitle: "Collect full payment for \"Goblin Bounty\"?",
			ConfirmLines: []string{"Outstanding: 25 GP"},
		},
		{
			Trigger:      ActionWriteOff,
			ID:           "quest.writeoff_full",
			Label:        "w write off",
			ConfirmTitle: "Write off \"Goblin Bounty\"?",
			ConfirmLines: []string{"Outstanding: 25 GP"},
		},
	}
	if collected {
		status = "paid"
		outstanding = "-"
		detailLines = []string{
			"Patron: Mayor Rowan",
			"Status: paid",
			"Promised reward: 25 GP",
			"Outstanding: 0 CP",
			"Collected so far: 25 GP",
			"Accounting state: fully collected",
		}
		actions = nil
	}

	return ShellData{
		Dashboard: DefaultDashboardData(),
		Quests: ListScreenData{
			HeaderLines:  []string{"Quest register from smoke.db.", "Select a quest to inspect it."},
			SummaryLines: []string{"Promised quests: 0", "Promised value: 0 CP", "Receivables: 1", "Outstanding: 25 GP"},
			Items: []ListItemData{
				{
					Key:         "quest-1",
					Row:         "25 GP        " + status + "    " + outstanding + " Goblin Bounty (Mayor Rowan)",
					DetailTitle: "Goblin Bounty",
					DetailLines: detailLines,
					Actions:     actions,
				},
			},
		},
	}
}

func testLootShellData(recognized bool) ShellData {
	status := "held"
	detailLines := []string{
		"Status: held",
		"Quantity: 1",
		"Accounting state: appraised but off-ledger",
		"Latest appraisal: 7 GP 5 SP",
		"Appraisals tracked: 2",
		"Appraised on: 2026-03-09",
		"Appraiser: Master jeweler",
		"Source: Merchant",
		"Holder: Bard",
	}
	actions := []ItemActionData{{
		Trigger:      ActionRecognize,
		ID:           "loot.recognize_latest",
		Label:        "n recognize",
		ConfirmTitle: "Recognize \"Gold Necklace\"?",
		ConfirmLines: []string{"Latest appraisal: 7 GP 5 SP"},
	}}
	if recognized {
		status = "recognized"
		detailLines = []string{
			"Status: recognized",
			"Quantity: 1",
			"Accounting state: on-ledger recognized inventory",
			"Latest appraisal: 7 GP 5 SP",
			"Appraisals tracked: 2",
			"Appraised on: 2026-03-09",
			"Appraiser: Master jeweler",
			"Source: Merchant",
			"Holder: Bard",
		}
		actions = nil
	}

	return ShellData{
		Dashboard: DefaultDashboardData(),
		Loot: ListScreenData{
			HeaderLines:  []string{"Unrealized loot register from smoke.db.", "Select a loot item to inspect it."},
			SummaryLines: []string{"Tracked items: 1", "Recognized: 0", "Total quantity: 1", "Appraised value: 7 GP 5 SP"},
			Items: []ListItemData{
				{
					Key:         "loot-1",
					Row:         "7 GP 5 SP    qty:1   " + status + " " + "Gold Necklace (Merchant)",
					DetailTitle: "Gold Necklace",
					DetailLines: detailLines,
					Actions:     actions,
				},
			},
		},
	}
}

func testLootSellShellData() ShellData {
	return ShellData{
		Dashboard: DefaultDashboardData(),
		Loot: ListScreenData{
			HeaderLines:  []string{"Unrealized loot register from smoke.db.", "Select a loot item to inspect it."},
			SummaryLines: []string{"Tracked items: 1", "Recognized: 1", "Total quantity: 1", "Appraised value: 7 GP 5 SP"},
			Items: []ListItemData{
				{
					Key:         "loot-1",
					Row:         "7 GP 5 SP    qty:1   recognized  Gold Necklace (Merchant)",
					DetailTitle: "Gold Necklace",
					DetailLines: []string{
						"Status: recognized",
						"Quantity: 1",
						"Accounting state: on-ledger recognized inventory",
						"Latest appraisal: 7 GP 5 SP",
						"Recognized value: 7 GP 5 SP",
						"Sale state: sellable from recognized basis",
					},
					Actions: []ItemActionData{{
						Trigger:     ActionSell,
						ID:          "loot.sell",
						Label:       "s sell",
						Mode:        ItemActionModeInput,
						InputTitle:  "Sell \"Gold Necklace\"?",
						InputPrompt: "Sale amount",
						InputHelp: []string{
							"Sale date: 2026-03-10",
							"Recognized value: 7 GP 5 SP",
							"Enter sale proceeds in GP/SP/CP format.",
						},
						Placeholder: "7 GP 5 SP",
					}},
				},
			},
		},
	}
}

func testLootAfterSaleShellData() ShellData {
	return ShellData{
		Dashboard: DefaultDashboardData(),
		Loot: ListScreenData{
			HeaderLines:  []string{"Unrealized loot register from smoke.db.", "Select a loot item to inspect it."},
			SummaryLines: []string{"Tracked items: 1", "Recognized: 1", "Total quantity: 1", "Appraised value: 2 GP"},
			Items: []ListItemData{
				{
					Key:         "loot-2",
					Row:         "2 GP         qty:1   recognized  Silver Chalice (Goblin den)",
					DetailTitle: "Silver Chalice",
					DetailLines: []string{
						"Status: recognized",
						"Quantity: 1",
						"Accounting state: on-ledger recognized inventory",
						"Recognized value: 2 GP",
					},
					Actions: []ItemActionData{{
						Trigger:     ActionSell,
						ID:          "loot.sell",
						Label:       "s sell",
						Mode:        ItemActionModeInput,
						InputTitle:  "Sell \"Silver Chalice\"?",
						InputPrompt: "Sale amount",
						Placeholder: "2 GP",
					}},
				},
			},
		},
	}
}

func itemActions(action *ItemActionData) []ItemActionData {
	if action == nil {
		return nil
	}

	return []ItemActionData{*action}
}

func simulationPlainText(screen tcell.SimulationScreen) string {
	cells, width, height := screen.GetContents()
	if width == 0 || height == 0 {
		return ""
	}

	lines := make([]string, 0, height)
	for y := range height {
		runes := make([]rune, width)
		for x := range width {
			cell := cells[(y*width)+x]
			if len(cell.Runes) == 0 {
				runes[x] = ' '
				continue
			}
			runes[x] = cell.Runes[0]
		}
		lines = append(lines, strings.TrimRight(string(runes), " "))
	}

	return strings.Join(lines, "\n")
}
