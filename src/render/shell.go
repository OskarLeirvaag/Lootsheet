package render

import (
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/render/goldrain"
)

type confirmState struct {
	Section Section
	ItemKey string
	Action  ItemActionData
}

type inputState struct {
	Section         Section
	ItemKey         string
	Action          ItemActionData
	Title           string
	Prompt          string
	Value           string
	Placeholder     string
	ErrorText       string
	RequiredMessage string
	HelpLines       []string
}

type glossaryState struct {
	Title string
	Lines []string
}

type codexPickerState struct {
	Section   Section
	Name      string
	TypeIndex int
	Focus     int // 0 = name, 1 = type list
	ErrorText string
}

type handleResult struct {
	Command *Command
	Quit    bool
	Redraw  bool
	Reload  bool
}

// Shell renders the interactive multi-screen TUI.
type Shell struct {
	Data            ShellData
	Section         Section
	settingsTab     int
	scrolls         map[Section]int
	selectedKeys    map[Section]string
	selectedIndexes map[Section]int
	viewHeights     map[Section]int
	status          StatusMessage
	confirm         *confirmState
	input           *inputState
	compose         *composeState
	glossary        *glossaryState
	editor          *editorState
	codexPicker     *codexPickerState
	search          *searchState
	searchHandler   SearchHandler
	rain            *goldrain.GoldRain

	editorSaveInFlight  bool
	editorQuitAfterSave bool
	quitConfirm         bool
	disconnected        bool
}

// SetSearchHandler installs a server-side search callback. Sections for which
// the handler returns nil fall back to client-side filtering.
func (s *Shell) SetSearchHandler(h SearchHandler) {
	if s != nil {
		s.searchHandler = h
	}
}

// NewShell constructs the interactive TUI shell state.
func NewShell(data *ShellData) *Shell {
	shell := &Shell{
		Data:            resolveShellData(data),
		Section:         SectionDashboard,
		scrolls:         make(map[Section]int),
		selectedKeys:    make(map[Section]string),
		selectedIndexes: make(map[Section]int),
		viewHeights:     make(map[Section]int),
		rain:            goldrain.NewGoldRain(),
	}
	shell.reconcileSelections()

	return shell
}

// TickRain advances the gold rain animation by one frame.
func (s *Shell) TickRain() {
	if s == nil || s.rain == nil {
		return
	}
	s.rain.Update()
}

func (s *Shell) activeSettingsSection() Section {
	if s.settingsTab >= 0 && s.settingsTab < len(settingsTabs) {
		return settingsTabs[s.settingsTab]
	}
	return settingsTabs[0]
}

// listSection returns the effective section for list data operations.
// For Settings, this returns the active tab's virtual section.
func (s *Shell) listSection() Section {
	if s.Section == SectionSettings {
		return s.activeSettingsSection()
	}
	return s.Section
}

// Reload swaps the shell snapshot while keeping navigation state intact.
func (s *Shell) Reload(data *ShellData) {
	if s == nil {
		return
	}

	s.Data = resolveShellData(data)
	s.confirm = nil
	s.input = nil
	s.compose = nil
	s.glossary = nil
	s.codexPicker = nil
	s.search = nil
	s.quitConfirm = false

	if s.editorSaveInFlight {
		s.editorSaveInFlight = false
		if s.editor != nil {
			s.editor.Dirty = false
			s.editor.StatusText = "Saved."
		}
		if s.editorQuitAfterSave {
			s.editorQuitAfterSave = false
			s.editor = nil
		}
	} else {
		s.editor = nil
	}

	s.reconcileSelections()
}

// SetDisconnected marks the shell as disconnected from the server.
// All open modals are cleared so the disconnect overlay renders on top.
func (s *Shell) SetDisconnected() {
	if s == nil {
		return
	}
	s.disconnected = true
	s.CloseModal()
}

// SetStatus updates the transient status line.
func (s *Shell) SetStatus(status StatusMessage) {
	if s == nil {
		return
	}

	s.status = status
}

// CloseModal closes any currently open modal.
func (s *Shell) CloseModal() {
	if s == nil {
		return
	}
	s.confirm = nil
	s.input = nil
	s.compose = nil
	s.glossary = nil
	s.codexPicker = nil
	s.search = nil
	s.editor = nil
	s.quitConfirm = false
}

// Navigate switches to the requested section and optionally selects the given item key.
func (s *Shell) Navigate(section Section, selectedKey string) {
	if s == nil {
		return
	}

	if section.Scrollable() {
		s.Section = section
		if strings.TrimSpace(selectedKey) != "" {
			s.selectedKeys[section] = strings.TrimSpace(selectedKey)
		}
		s.reconcileSelection(section)
		return
	}

	s.Section = section
}

func (s *Shell) listDataForSection(section Section) *ListScreenData {
	switch section {
	case SectionAccounts:
		return &s.Data.Accounts
	case SectionJournal:
		return &s.Data.Journal
	case SectionQuests:
		return &s.Data.Quests
	case SectionLoot:
		return &s.Data.Loot
	case SectionAssets:
		return &s.Data.Assets
	case SectionCodex:
		return &s.Data.Codex
	case SectionNotes:
		return &s.Data.Notes
	case settingsTabAccounts:
		return &s.Data.SettingsAccounts
	case settingsTabCodexTypes:
		return &s.Data.SettingsCodexTypes
	case settingsTabCampaigns:
		return &s.Data.SettingsCampaigns
	default:
		return nil
	}
}
