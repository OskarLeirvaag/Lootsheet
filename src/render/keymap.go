package render

import "github.com/OskarLeirvaag/Lootsheet/src/render/model"

// Type aliases re-export model keymap types.
type Action = model.Action
type KeyStroke = model.KeyStroke
type Binding = model.Binding
type KeyMap = model.KeyMap

const (
	ActionNone            = model.ActionNone
	ActionQuit            = model.ActionQuit
	ActionRedraw          = model.ActionRedraw
	ActionNextSection     = model.ActionNextSection
	ActionPrevSection     = model.ActionPrevSection
	ActionShowDashboard   = model.ActionShowDashboard
	ActionShowSettings    = model.ActionShowSettings
	ActionShowLedger      = model.ActionShowLedger
	ActionShowJournal     = model.ActionShowJournal
	ActionShowQuests      = model.ActionShowQuests
	ActionShowLoot        = model.ActionShowLoot
	ActionShowAssets      = model.ActionShowAssets
	ActionTransfer        = model.ActionTransfer
	ActionMoveUp          = model.ActionMoveUp
	ActionMoveDown        = model.ActionMoveDown
	ActionPageUp          = model.ActionPageUp
	ActionPageDown        = model.ActionPageDown
	ActionMoveTop         = model.ActionMoveTop
	ActionMoveBottom      = model.ActionMoveBottom
	ActionEdit            = model.ActionEdit
	ActionDelete          = model.ActionDelete
	ActionToggle          = model.ActionToggle
	ActionReverse         = model.ActionReverse
	ActionCollect         = model.ActionCollect
	ActionWriteOff        = model.ActionWriteOff
	ActionAppraise        = model.ActionAppraise
	ActionRecognize       = model.ActionRecognize
	ActionSell            = model.ActionSell
	ActionNewExpense      = model.ActionNewExpense
	ActionNewIncome       = model.ActionNewIncome
	ActionNewCustom       = model.ActionNewCustom
	ActionEditTemplate    = model.ActionEditTemplate
	ActionExecuteTemplate = model.ActionExecuteTemplate
	ActionSubmitCompose   = model.ActionSubmitCompose
	ActionConfirm         = model.ActionConfirm
	ActionShowCodex       = model.ActionShowCodex
	ActionShowNotes       = model.ActionShowNotes
	ActionShowCompendium = model.ActionShowCompendium
	ActionHelp            = model.ActionHelp
	ActionSearch          = model.ActionSearch
	ActionSwitchCampaign  = model.ActionSwitchCampaign
	ActionExportCSV       = model.ActionExportCSV
	ActionExportExcel     = model.ActionExportExcel
	ActionExportPDF       = model.ActionExportPDF
)

// DefaultKeyMap returns the interactive keyboard controls.
var DefaultKeyMap = model.DefaultKeyMap
