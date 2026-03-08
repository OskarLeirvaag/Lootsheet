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
			sim.SetSize(96, 28)
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
		CommandHandler: func(_ context.Context, command Command) (ShellData, StatusMessage, error) {
			called = true
			got = command
			return updated, StatusMessage{
				Level: StatusSuccess,
				Text:  "Account 1000 deactivated.",
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
		CommandHandler: func(context.Context, Command) (ShellData, StatusMessage, error) {
			return ShellData{}, StatusMessage{}, errors.New("account 1000 cannot be deactivated right now")
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

func testAccountsShellData(active bool) ShellData {
	status := "inactive"
	summary := []string{"Accounts: 1 total", "Active: 0  Inactive: 1"}
	action := &ItemActionData{
		ID:           "account.activate",
		Label:        "t activate",
		ConfirmTitle: "Activate account 1000?",
		ConfirmLines: []string{"Party Cash"},
	}
	if active {
		status = "active"
		summary = []string{"Accounts: 1 total", "Active: 1  Inactive: 0"}
		action = &ItemActionData{
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
					PrimaryAction: action,
				},
			},
		},
	}
}

func simulationPlainText(screen tcell.SimulationScreen) string {
	cells, width, height := screen.GetContents()
	if width == 0 || height == 0 {
		return ""
	}

	lines := make([]string, 0, height)
	for y := 0; y < height; y++ {
		runes := make([]rune, width)
		for x := 0; x < width; x++ {
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
