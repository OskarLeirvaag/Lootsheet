package render

import (
	"context"
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
