package render

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestShellRenderShowsTabsAndFooterHelp(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(100, 28, theme.Base)

	data := DefaultShellData()
	NewShell(&data).Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	for _, token := range []string{
		"LootSheet TUI",
		"Section: Dashboard",
		"Sections: [Dashboard]   Journal      Quests       Loot",
		"1-7 jump",
		"e/i/a entry",
		"? terms",
		"q quit",
		"Ctrl+L refresh",
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("shell output missing %q:\n%s", token, output)
		}
	}
}

func TestShellTabsLineKeepsStableWidthAcrossSections(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)

	dashboardTabs := shell.tabsLine()
	shell.Section = SectionJournal
	journalTabs := shell.tabsLine()
	shell.Section = SectionLoot
	lootTabs := shell.tabsLine()

	if len(dashboardTabs) != len(journalTabs) || len(journalTabs) != len(lootTabs) {
		t.Fatalf("tab widths changed across sections: dashboard=%d journal=%d loot=%d", len(dashboardTabs), len(journalTabs), len(lootTabs))
	}

	for _, token := range []string{"[Dashboard]", "[Journal]", "[Loot]"} {
		if !strings.Contains(dashboardTabs+journalTabs+lootTabs, token) {
			t.Fatalf("tabs output missing active marker %q", token)
		}
	}
}

func TestShellRenderShowsScrollableSettingsScreen(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(96, 28, theme.Base)

	items := make([]ListItemData, 0, 12)
	for index := range 12 {
		row := fmt.Sprintf("%04d asset active Account %02d", 1000+index, index)
		items = append(items, ListItemData{
			Key:         fmt.Sprintf("%04d", 1000+index),
			Row:         row,
			DetailTitle: fmt.Sprintf("Account %02d", index),
			DetailLines: []string{fmt.Sprintf("Detail for account %02d", index)},
		})
	}

	data := ShellData{
		Dashboard: DefaultDashboardData(),
		SettingsAccounts: ListScreenData{
			HeaderLines:  []string{"Accounts from smoke.db.", "Chart of accounts."},
			SummaryLines: []string{"Accounts: 12 total", "Active: 12  Inactive: 0"},
			Items:        items,
		},
	}
	shell := NewShell(&data)

	shell.HandleAction(ActionShowSettings)
	for range 6 {
		shell.HandleAction(ActionMoveDown)
	}
	shell.Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	if !strings.Contains(output, "Section: Settings") {
		t.Fatalf("settings screen missing section header:\n%s", output)
	}
	if !strings.Contains(output, "Accounts ") || !strings.Contains(output, "/12") {
		t.Fatalf("settings screen missing scroll title:\n%s", output)
	}
	if strings.Contains(output, "Account 00") {
		t.Fatalf("settings screen did not scroll:\n%s", output)
	}
	if !strings.Contains(output, "Account 06") {
		t.Fatalf("settings screen missing expected visible row:\n%s", output)
	}
	if !strings.Contains(output, "↑↓ select") {
		t.Fatalf("settings screen missing selection help:\n%s", output)
	}
	if !strings.Contains(output, "Detail for account 06") {
		t.Fatalf("settings screen missing detail pane content:\n%s", output)
	}
}

func TestShellRenderKeepsDetailVisibleOnStandardTerminal(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(80, 24, theme.Base)

	data := ShellData{
		Dashboard: DefaultDashboardData(),
		SettingsAccounts: ListScreenData{
			HeaderLines:  []string{"Accounts from smoke.db.", "Chart of accounts."},
			SummaryLines: []string{"Accounts: 2 total", "Active: 2  Inactive: 0"},
			Items: []ListItemData{
				{
					Key:         "1000",
					Row:         "1000 asset active Party Cash",
					DetailTitle: "Account 1000",
					DetailLines: []string{"Name: Party Cash", "Status: active"},
				},
				{
					Key:         "1100",
					Row:         "1100 asset active Quest Receivable",
					DetailTitle: "Account 1100",
					DetailLines: []string{"Name: Quest Receivable", "Status: active"},
				},
			},
		},
	}

	shell := NewShell(&data)
	shell.HandleAction(ActionShowSettings)
	shell.Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	for _, token := range []string{"Account 1000", "Name: Party Cash"} {
		if !strings.Contains(output, token) {
			t.Fatalf("standard terminal output missing %q:\n%s", token, output)
		}
	}
}

func TestShellActionOpensConfirmAndEmitsCommand(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		SettingsAccounts: ListScreenData{
			Items: []ListItemData{
				{
					Key:         "1000",
					Row:         "1000 asset active Party Cash",
					DetailTitle: "Account 1000",
					DetailLines: []string{"Name: Party Cash"},
					Actions: []ItemActionData{{
						Trigger:      ActionToggle,
						ID:           "account.deactivate",
						Label:        "t deactivate",
						ConfirmTitle: "Deactivate account 1000?",
						ConfirmLines: []string{"Party Cash"},
					}},
				},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionShowSettings)

	result := shell.HandleAction(ActionToggle)
	if !result.Redraw {
		t.Fatal("expected primary action to trigger redraw")
	}
	if shell.confirm == nil {
		t.Fatal("expected confirm modal to open")
	}

	cancel := shell.HandleAction(ActionQuit)
	if !cancel.Redraw {
		t.Fatal("expected quit inside confirm modal to cancel it")
	}
	if shell.confirm != nil {
		t.Fatal("expected confirm modal to close on quit")
	}

	shell.HandleAction(ActionToggle)
	result = shell.HandleAction(ActionConfirm)
	if result.Command == nil {
		t.Fatal("expected confirm action to emit command")
	}
	if result.Command.ID != "account.deactivate" {
		t.Fatalf("command id = %q, want account.deactivate", result.Command.ID)
	}
	if result.Command.ItemKey != "1000" {
		t.Fatalf("command item key = %q, want 1000", result.Command.ItemKey)
	}
}

func TestShellJournalReverseUsesOnlyReverseAction(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Journal: ListScreenData{
			Items: []ListItemData{
				{
					Key:         "entry-1",
					Row:         "#1    2026-03-08 posted   Restock arrows",
					DetailTitle: "Entry #1",
					DetailLines: []string{"Date: 2026-03-08", "Lines:", "5100 Adventuring Supplies DR 25 CP"},
					Actions: []ItemActionData{{
						Trigger:      ActionReverse,
						ID:           "journal.reverse",
						Label:        "r reverse",
						ConfirmTitle: "Reverse entry #1?",
						ConfirmLines: []string{"Restock arrows"},
					}},
				},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionShowJournal)

	if result := shell.HandleAction(ActionToggle); result.Redraw || shell.confirm != nil {
		t.Fatalf("toggle action should not open journal reverse modal: %#v", result)
	}

	result := shell.HandleAction(ActionReverse)
	if !result.Redraw {
		t.Fatal("expected reverse action to trigger redraw")
	}
	if shell.confirm == nil {
		t.Fatal("expected confirm modal to open for journal reverse")
	}

	result = shell.HandleAction(ActionConfirm)
	if result.Command == nil {
		t.Fatal("expected journal confirm to emit command")
	}
	if result.Command.ID != "journal.reverse" {
		t.Fatalf("command id = %q, want journal.reverse", result.Command.ID)
	}
	if result.Command.ItemKey != "entry-1" {
		t.Fatalf("command item key = %q, want entry-1", result.Command.ItemKey)
	}
}

func TestShellQuestActionsShowCombinedFooterAndMatchTriggers(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Quests: ListScreenData{
			Items: []ListItemData{
				{
					Key:         "quest-1",
					Row:         "25 GP collectible 25 GP due Goblin Bounty (Mayor Rowan)",
					DetailTitle: "Goblin Bounty",
					DetailLines: []string{"Outstanding: 25 GP", "Collected so far: 0 CP"},
					Actions: []ItemActionData{
						{
							Trigger:      ActionEdit,
							ID:           "quest.update",
							Label:        "u edit",
							Mode:         ItemActionModeCompose,
							ComposeMode:  "quest",
							ComposeTitle: "Edit Quest",
							ComposeFields: map[string]string{
								"title":       "Goblin Bounty",
								"patron":      "Mayor Rowan",
								"description": "",
								"reward":      "25 GP",
								"advance":     "0 CP",
								"bonus":       "",
								"notes":       "",
								"status":      "collectible",
								"accepted_on": "2026-03-08",
							},
						},
						{
							Trigger:      ActionCollect,
							ID:           "quest.collect_full",
							Label:        "c collect",
							ConfirmTitle: "Collect full payment?",
							ConfirmLines: []string{"Outstanding: 25 GP"},
						},
						{
							Trigger:      ActionWriteOff,
							ID:           "quest.writeoff_full",
							Label:        "w write off",
							ConfirmTitle: "Write off quest?",
							ConfirmLines: []string{"Outstanding: 25 GP"},
						},
					},
				},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionShowQuests)

	if help := shell.footerHelpText(DefaultKeyMap()); !strings.Contains(help, "u edit") || !strings.Contains(help, "c collect") || !strings.Contains(help, "w write off") {
		t.Fatalf("quest footer help = %q, want edit, collect, and write off labels", help)
	}

	if result := shell.HandleAction(ActionToggle); result.Redraw || shell.confirm != nil {
		t.Fatalf("toggle action should not open quest modal: %#v", result)
	}

	if result := shell.HandleAction(ActionCollect); !result.Redraw || shell.confirm == nil || shell.confirm.Action.ID != "quest.collect_full" {
		t.Fatalf("collect action did not open collect modal: %#v %#v", result, shell.confirm)
	}
	shell.HandleAction(ActionQuit)

	if result := shell.HandleAction(ActionWriteOff); !result.Redraw || shell.confirm == nil || shell.confirm.Action.ID != "quest.writeoff_full" {
		t.Fatalf("write-off action did not open write-off modal: %#v %#v", result, shell.confirm)
	}
}

func TestShellEditActionOpensPrefilledCompose(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Quests: ListScreenData{
			Items: []ListItemData{
				{
					Key:         "quest-1",
					Row:         "25 GP collectible 25 GP due Goblin Bounty (Mayor Rowan)",
					DetailTitle: "Goblin Bounty",
					DetailLines: []string{"Outstanding: 25 GP"},
					Actions: []ItemActionData{{
						Trigger:      ActionEdit,
						ID:           "quest.update",
						Label:        "u edit",
						Mode:         ItemActionModeCompose,
						ComposeMode:  "quest",
						ComposeTitle: "Edit Quest",
						ComposeFields: map[string]string{
							"title":       "Goblin Bounty",
							"patron":      "Mayor Rowan",
							"description": "Clear the cave",
							"reward":      "25 GP",
							"advance":     "0 CP",
							"bonus":       "",
							"notes":       "Bring proof",
							"status":      "collectible",
							"accepted_on": "2026-03-08",
						},
					}},
				},
			},
		},
	}

	shell := NewShell(&data)
	shell.HandleAction(ActionShowQuests)

	result := shell.HandleAction(ActionEdit)
	if !result.Redraw || shell.compose == nil {
		t.Fatalf("edit action did not open compose: %#v %#v", result, shell.compose)
	}
	if shell.compose.CommandID != "quest.update" {
		t.Fatalf("compose command id = %q, want quest.update", shell.compose.CommandID)
	}
	if shell.compose.Fields["title"] != "Goblin Bounty" {
		t.Fatalf("compose title field = %q, want Goblin Bounty", shell.compose.Fields["title"])
	}
	if shell.compose.Fields["status"] != "collectible" {
		t.Fatalf("compose status field = %q, want collectible", shell.compose.Fields["status"])
	}
}

func TestShellLootRecognizeUsesOnlyRecognizeAction(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Loot: ListScreenData{
			Items: []ListItemData{
				{
					Key:         "loot-1",
					Row:         "7 GP 5 SP    qty:1   held        Gold Necklace (Merchant)",
					DetailTitle: "Gold Necklace",
					DetailLines: []string{"Latest appraisal: 7 GP 5 SP", "Accounting state: appraised but off-ledger"},
					Actions: []ItemActionData{{
						Trigger:      ActionRecognize,
						ID:           "loot.recognize_latest",
						Label:        "n recognize",
						ConfirmTitle: "Recognize \"Gold Necklace\"?",
						ConfirmLines: []string{"Latest appraisal: 7 GP 5 SP"},
					}},
				},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionShowLoot)

	if help := shell.footerHelpText(DefaultKeyMap()); !strings.Contains(help, "n recognize") {
		t.Fatalf("loot footer help = %q, want n recognize", help)
	}

	if result := shell.HandleAction(ActionCollect); result.Redraw || shell.confirm != nil {
		t.Fatalf("collect action should not open loot recognize modal: %#v", result)
	}

	result := shell.HandleAction(ActionRecognize)
	if !result.Redraw {
		t.Fatal("expected recognize action to trigger redraw")
	}
	if shell.confirm == nil {
		t.Fatal("expected confirm modal to open for loot recognition")
	}
	if shell.confirm.Action.ID != "loot.recognize_latest" {
		t.Fatalf("confirm action id = %q, want loot.recognize_latest", shell.confirm.Action.ID)
	}
}

func TestShellLootSellUsesInputModalAndCarriesAmount(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Loot: ListScreenData{
			Items: []ListItemData{
				{
					Key:         "loot-1",
					Row:         "7 GP 5 SP    qty:1   recognized  Gold Necklace (Merchant)",
					DetailTitle: "Gold Necklace",
					DetailLines: []string{"Status: recognized", "Recognized value: 7 GP 5 SP"},
					Actions: []ItemActionData{{
						Trigger:     ActionSell,
						ID:          "loot.sell",
						Label:       "s sell",
						Mode:        ItemActionModeInput,
						InputTitle:  "Sell \"Gold Necklace\"?",
						InputPrompt: "Sale amount",
						InputHelp:   []string{"Sale date: 2026-03-10", "Recognized value: 7 GP 5 SP"},
						Placeholder: "7 GP 5 SP",
					}},
				},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionShowLoot)

	if help := shell.footerHelpText(DefaultKeyMap()); !strings.Contains(help, "s sell") {
		t.Fatalf("loot footer help = %q, want s sell", help)
	}

	if result := shell.HandleAction(ActionRecognize); result.Redraw || shell.input != nil {
		t.Fatalf("recognize action should not open loot sell modal: %#v", result)
	}

	result := shell.HandleAction(ActionSell)
	if !result.Redraw {
		t.Fatal("expected sell action to trigger redraw")
	}
	if shell.input == nil {
		t.Fatal("expected input modal to open for loot sale")
	}

	if result, handled := shell.handleInputKeyEvent(tcell.NewEventKey(tcell.KeyRune, '7', tcell.ModNone), ActionNone); !handled || !result.Redraw {
		t.Fatalf("typing first rune did not update input: %#v handled=%v", result, handled)
	}
	shell.handleInputKeyEvent(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone), ActionNone)
	shell.handleInputKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone), ActionNone)
	shell.handleInputKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone), ActionNone)

	result, handled := shell.handleInputKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionConfirm)
	if !handled || result.Command == nil {
		t.Fatalf("enter did not emit sale command: %#v handled=%v", result, handled)
	}
	if result.Command.ID != "loot.sell" {
		t.Fatalf("command id = %q, want loot.sell", result.Command.ID)
	}
	if result.Command.Fields["amount"] != "7 gp" {
		t.Fatalf("command amount = %q, want 7 gp", result.Command.Fields["amount"])
	}
}

func TestShellInputModalShowsBlankSubmitErrorAndClearHelp(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Loot: ListScreenData{
			Items: []ListItemData{
				{
					Key:         "loot-1",
					Row:         "7 GP 5 SP    qty:1   recognized  Gold Necklace (Merchant)",
					DetailTitle: "Gold Necklace",
					DetailLines: []string{"Status: recognized"},
					Actions: []ItemActionData{{
						Trigger:     ActionSell,
						ID:          "loot.sell",
						Label:       "s sell",
						Mode:        ItemActionModeInput,
						InputTitle:  "Sell \"Gold Necklace\"?",
						InputPrompt: "Sale amount",
						Placeholder: "7 GP 5 SP",
					}},
				},
			},
		},
	}

	shell := NewShell(&data)
	shell.HandleAction(ActionShowLoot)
	shell.HandleAction(ActionSell)

	result, handled := shell.handleInputKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionConfirm)
	if !handled || !result.Redraw {
		t.Fatalf("blank submit should redraw with an error: %#v handled=%v", result, handled)
	}
	if shell.input == nil || shell.input.ErrorText == "" {
		t.Fatal("expected blank submit to keep input modal open with error text")
	}

	if help := shell.footerHelpText(DefaultKeyMap()); !strings.Contains(help, "Ctrl+U clear") {
		t.Fatalf("input footer help = %q, want clear help", help)
	}
}

func TestShellReloadPreservesSelectionByKey(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		SettingsAccounts: ListScreenData{
			Items: []ListItemData{
				{Key: "1000", Row: "1000 asset active Party Cash"},
				{Key: "1100", Row: "1100 asset active Quest Receivable"},
				{Key: "1200", Row: "1200 asset active Loot Inventory"},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionShowSettings)
	shell.HandleAction(ActionMoveDown)

	if item := shell.currentSelectedItem(settingsTabAccounts); item == nil || item.Key != "1100" {
		t.Fatalf("selected item before reload = %#v, want key 1100", item)
	}

	reloaded := ShellData{
		Dashboard: DefaultDashboardData(),
		SettingsAccounts: ListScreenData{
			Items: []ListItemData{
				{Key: "0900", Row: "0900 asset active Treasury Box"},
				{Key: "1100", Row: "1100 asset active Quest Receivable"},
				{Key: "1200", Row: "1200 asset active Loot Inventory"},
			},
		},
	}
	shell.Reload(&reloaded)

	if item := shell.currentSelectedItem(settingsTabAccounts); item == nil || item.Key != "1100" {
		t.Fatalf("selected item after reload = %#v, want key 1100", item)
	}
}

func TestShellExpenseComposeOpensAndEmitsCommand(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		EntryCatalog: EntryCatalog{
			DefaultDate: "2026-03-10",
			ExpenseAccounts: []AccountOption{
				{Code: "5100", Name: "Arrows & Ammunition", Type: "expense"},
			},
			FundingAccounts: []AccountOption{
				{Code: "1000", Name: "Party Cash", Type: "asset"},
			},
		},
	}

	shell := NewShell(&data)
	if result := shell.HandleAction(ActionNewExpense); !result.Redraw || shell.compose == nil {
		t.Fatalf("expected expense compose to open: %#v %#v", result, shell.compose)
	}

	for _, event := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
	} {
		shell.handleComposeKeyEvent(event, DefaultKeyMap().Resolve(event))
	}

	result, handled := shell.handleComposeKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionConfirm)
	if !handled || result.Command == nil {
		t.Fatalf("submit did not emit expense command: %#v handled=%v", result, handled)
	}
	if result.Command.ID != "entry.expense.create" {
		t.Fatalf("command id = %q, want entry.expense.create", result.Command.ID)
	}
	if result.Command.Fields["date"] != "2026-03-10" {
		t.Fatalf("command date = %q, want 2026-03-10", result.Command.Fields["date"])
	}
	if result.Command.Fields["description"] != "Arrows" {
		t.Fatalf("command description = %q, want Arrows", result.Command.Fields["description"])
	}
	if result.Command.Fields["amount"] != "25" {
		t.Fatalf("command amount = %q, want 25", result.Command.Fields["amount"])
	}
	if result.Command.Fields["account_code"] != "5100" {
		t.Fatalf("command account code = %q, want typed account code", result.Command.Fields["account_code"])
	}
}

func TestShellExpenseComposeAcceptsArrowKeyNavigation(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		EntryCatalog: EntryCatalog{
			DefaultDate: "2026-03-10",
			ExpenseAccounts: []AccountOption{
				{Code: "5100", Name: "Arrows & Ammunition", Type: "expense"},
			},
			FundingAccounts: []AccountOption{
				{Code: "1000", Name: "Party Cash", Type: "asset"},
			},
		},
	}

	shell := NewShell(&data)
	if result := shell.HandleAction(ActionNewExpense); !result.Redraw || shell.compose == nil {
		t.Fatalf("expected expense compose to open: %#v %#v", result, shell.compose)
	}

	for _, event := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
	} {
		shell.handleComposeKeyEvent(event, DefaultKeyMap().Resolve(event))
	}

	result, handled := shell.handleComposeKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionConfirm)
	if !handled || result.Command == nil {
		t.Fatalf("submit did not emit expense command: %#v handled=%v", result, handled)
	}
	if result.Command.Fields["description"] != "Arrows" {
		t.Fatalf("command description = %q, want Arrows", result.Command.Fields["description"])
	}
	if result.Command.Fields["amount"] != "25" {
		t.Fatalf("command amount = %q, want 25", result.Command.Fields["amount"])
	}
	if result.Command.Fields["account_code"] != "5100" {
		t.Fatalf("command account code = %q, want 5100", result.Command.Fields["account_code"])
	}
}

func TestShellSectionLaunchersFollowCurrentScreen(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		SettingsAccounts: ListScreenData{
			Items: []ListItemData{{
				Key: "1000",
				Row: "1000 asset active Party Cash",
				Actions: []ItemActionData{
					{Trigger: ActionDelete, ID: "account.delete", Label: "d remove"},
					{Trigger: ActionToggle, ID: "account.deactivate", Label: "t deactivate"},
				},
			}},
		},
		Journal: ListScreenData{
			Items: []ListItemData{{
				Key:     "entry-1",
				Row:     "#1 2026-03-10 posted Restock",
				Actions: []ItemActionData{{Trigger: ActionReverse, ID: "journal.reverse", Label: "r reverse"}},
			}},
		},
		Quests: ListScreenData{
			Items: []ListItemData{{
				Key:     "quest-1",
				Row:     "25 GP collectible Goblin Bounty",
				Actions: []ItemActionData{{Trigger: ActionCollect, ID: "quest.collect_full", Label: "c collect"}},
			}},
		},
		Loot: ListScreenData{
			Items: []ListItemData{{
				Key:     "loot-1",
				Row:     "7 GP held Gold Necklace",
				Actions: []ItemActionData{{Trigger: ActionRecognize, ID: "loot.recognize_latest", Label: "n recognize"}},
			}},
		},
	}

	shell := NewShell(&data)
	if help := shell.footerHelpText(DefaultKeyMap()); !strings.Contains(help, "e/i/a entry") {
		t.Fatalf("dashboard help = %q", help)
	}

	shell.HandleAction(ActionShowSettings)
	if help := shell.footerHelpText(DefaultKeyMap()); !strings.Contains(help, "a add account") || !strings.Contains(help, "d remove") || !strings.Contains(help, "t deactivate") {
		t.Fatalf("settings help = %q", help)
	}

	shell.HandleAction(ActionShowJournal)
	if help := shell.footerHelpText(DefaultKeyMap()); !strings.Contains(help, "e/i entry") || !strings.Contains(help, "r reverse") || strings.Contains(help, "e/i/a entry") {
		t.Fatalf("journal help = %q", help)
	}

	shell.HandleAction(ActionShowQuests)
	if help := shell.footerHelpText(DefaultKeyMap()); !strings.Contains(help, "a add") || !strings.Contains(help, "u edit") || !strings.Contains(help, "c collect") {
		t.Fatalf("quests help = %q", help)
	}

	shell.HandleAction(ActionShowLoot)
	if help := shell.footerHelpText(DefaultKeyMap()); !strings.Contains(help, "a add") || !strings.Contains(help, "u edit") || !strings.Contains(help, "n recognize") {
		t.Fatalf("loot help = %q", help)
	}
}

func TestShellSectionSpecificComposeLaunches(t *testing.T) {
	data := DefaultShellData()
	data.EntryCatalog.DefaultDate = "2026-03-10"

	shell := NewShell(&data)
	if result := shell.HandleAction(ActionNewCustom); !result.Redraw || shell.compose == nil || shell.compose.Mode != composeModeCustom {
		t.Fatalf("dashboard a should open custom compose: %#v %#v", result, shell.compose)
	}

	shell.compose = nil
	shell.HandleAction(ActionShowSettings)
	if result := shell.HandleAction(ActionNewCustom); !result.Redraw || shell.compose == nil || shell.compose.Mode != composeModeAccount {
		t.Fatalf("settings a should open account compose: %#v %#v", result, shell.compose)
	}

	shell.compose = nil
	shell.HandleAction(ActionShowJournal)
	if result := shell.HandleAction(ActionNewExpense); !result.Redraw || shell.compose == nil || shell.compose.Mode != composeModeExpense {
		t.Fatalf("journal e should open expense compose: %#v %#v", result, shell.compose)
	}

	shell.compose = nil
	if result := shell.HandleAction(ActionNewCustom); result.Redraw || shell.compose != nil {
		t.Fatalf("journal a should not open compose: %#v %#v", result, shell.compose)
	}

	shell.HandleAction(ActionShowQuests)
	if result := shell.HandleAction(ActionNewCustom); !result.Redraw || shell.compose == nil || shell.compose.Mode != composeModeQuest {
		t.Fatalf("quests a should open quest compose: %#v %#v", result, shell.compose)
	}

	shell.compose = nil
	shell.HandleAction(ActionShowLoot)
	if result := shell.HandleAction(ActionNewCustom); !result.Redraw || shell.compose == nil || shell.compose.Mode != composeModeLoot {
		t.Fatalf("loot a should open loot compose: %#v %#v", result, shell.compose)
	}
}

func TestShellGlossaryModalOpensAndCloses(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(100, 28, theme.Base)

	data := DefaultShellData()
	shell := NewShell(&data)
	result := shell.HandleAction(ActionHelp)
	if !result.Redraw || shell.glossary == nil {
		t.Fatalf("expected glossary to open: %#v %#v", result, shell.glossary)
	}

	shell.Render(buffer, &theme, keymap)
	output := buffer.PlainText()
	for _, token := range []string{"Dashboard Terms", "To share now:", "? close"} {
		if !strings.Contains(output, token) {
			t.Fatalf("glossary output missing %q:\n%s", token, output)
		}
	}

	result = shell.HandleAction(ActionHelp)
	if !result.Redraw || shell.glossary != nil {
		t.Fatalf("expected glossary to close on ?: %#v %#v", result, shell.glossary)
	}
}

func TestSearchModalOpensAndCloses(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)

	result := shell.HandleAction(ActionSearch)
	if !result.Redraw || shell.search == nil {
		t.Fatalf("expected search to open: %#v", result)
	}

	result, handled := shell.handleSearchKeyEvent(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), ActionQuit)
	if !handled || !result.Redraw || shell.search != nil {
		t.Fatalf("expected search to close on Esc: %#v handled=%v search=%v", result, handled, shell.search)
	}
}

func TestSearchFiltersResults(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Loot: ListScreenData{
			Items: []ListItemData{
				{Key: "loot-1", Row: "Gold Necklace", DetailTitle: "Gold Necklace"},
				{Key: "loot-2", Row: "Silver Ring", DetailTitle: "Silver Ring"},
			},
		},
		Quests: ListScreenData{
			Items: []ListItemData{
				{Key: "quest-1", Row: "Goblin Bounty", DetailTitle: "Goblin Bounty"},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionSearch)

	// Type "gold".
	for _, r := range "gold" {
		shell.handleSearchKeyEvent(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone), ActionNone)
	}

	if len(shell.search.Results) != 1 {
		t.Fatalf("expected 1 result for 'gold', got %d", len(shell.search.Results))
	}
	if shell.search.Results[0].ItemKey != "loot-1" {
		t.Fatalf("expected loot-1, got %s", shell.search.Results[0].ItemKey)
	}
}

func TestSearchSectionFilter(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Loot: ListScreenData{
			Items: []ListItemData{
				{Key: "loot-1", Row: "Gold Necklace"},
			},
		},
		Quests: ListScreenData{
			Items: []ListItemData{
				{Key: "quest-1", Row: "Goblin Bounty"},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionSearch)

	// All sections — should see both items.
	if len(shell.search.Results) < 2 {
		t.Fatalf("expected at least 2 results in All filter, got %d", len(shell.search.Results))
	}

	// Cycle Right to first section (Journal) — no items there.
	shell.handleSearchKeyEvent(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone), ActionNone)
	// Keep cycling to Quests (index 2 in searchableSections: Journal=1, Quests=2).
	shell.handleSearchKeyEvent(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone), ActionNone)
	if shell.search.FilterIndex != 2 {
		t.Fatalf("expected filter index 2 (Quests), got %d", shell.search.FilterIndex)
	}
	if len(shell.search.Results) != 1 || shell.search.Results[0].ItemKey != "quest-1" {
		t.Fatalf("expected quest-1 in Quests filter, got %v", shell.search.Results)
	}
}

func TestSearchEnterNavigates(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Loot: ListScreenData{
			Items: []ListItemData{
				{Key: "loot-1", Row: "Gold Necklace"},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionSearch)

	// Cycle filter to Loot (index 3 in searchableSections: Journal=1, Quests=2, Loot=3).
	for range 3 {
		shell.handleSearchKeyEvent(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone), ActionNone)
	}
	if len(shell.search.Results) != 1 {
		t.Fatalf("expected 1 result in Loot filter, got %d", len(shell.search.Results))
	}

	result, handled := shell.handleSearchKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionConfirm)
	if !handled || !result.Redraw {
		t.Fatalf("enter did not navigate: %#v handled=%v", result, handled)
	}
	if shell.search != nil {
		t.Fatal("expected search modal to close after Enter")
	}
	if shell.Section != SectionLoot {
		t.Fatalf("expected section Loot, got %v", shell.Section)
	}
	if shell.selectedKeys[SectionLoot] != "loot-1" {
		t.Fatalf("expected selected key loot-1, got %s", shell.selectedKeys[SectionLoot])
	}
}

func TestSearchFooterHelp(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	keymap := DefaultKeyMap()

	help := shell.footerHelpText(keymap)
	if !strings.Contains(help, "/ search") {
		t.Fatalf("normal footer missing '/ search': %s", help)
	}

	shell.HandleAction(ActionSearch)
	help = shell.footerHelpText(keymap)
	if !strings.Contains(help, "Enter select") || !strings.Contains(help, "Esc close") {
		t.Fatalf("search footer missing expected help: %s", help)
	}
	if strings.Contains(help, "/ search") {
		t.Fatalf("search footer should not show '/ search': %s", help)
	}
}
