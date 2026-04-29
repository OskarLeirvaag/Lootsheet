// Package ddb provides an HTTP client for the D&D Beyond API.
// It calls DDB services directly (no proxy) and returns pure Go types.
package ddb

import "encoding/json"

// --- Config types (from /api/config/json, no auth needed) ---

// Config holds the parsed game configuration from D&D Beyond.
type Config struct {
	Sources          []ConfigSource          `json:"sources"`
	Conditions       []ConfigConditionEntry  `json:"conditions"`
	BasicActions     []ConfigBasicAction     `json:"basicActions"`
	Rules            []ConfigRule            `json:"rules"`
	WeaponProperties []ConfigWeaponProperty  `json:"weaponProperties"`
	ChallengeRatings []ConfigChallengeRating `json:"challengeRatings"`
	CreatureSizes    []ConfigCreatureSize    `json:"creatureSizes"`
	MonsterTypes     []ConfigMonsterType     `json:"monsterTypes"`
}

type ConfigSource struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	IsReleased       bool   `json:"isReleased"`
	SourceCategoryID int    `json:"sourceCategoryId"`
}

// BadSourceIDs are DDB source IDs we always exclude. Ported from
// ddb-adventure-muncher (munch/data/ddb.js BAD_IDS): broken/empty entries that
// otherwise pollute the picker.
var BadSourceIDs = map[int]struct{}{
	31: {}, // CR data, file not found
	53: {}, // SAC, file not found
	42: {}, // TMR, file not found
	29: {}, // UA, no content
	4:  {}, // EE players
	26: {}, // CoS players
	30: {}, // ddb
}

type ConfigConditionEntry struct {
	Definition ConfigConditionDef `json:"definition"`
}

type ConfigConditionDef struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
}

type ConfigBasicAction struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ConfigRule struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ConfigWeaponProperty struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ConfigChallengeRating struct {
	ID    int     `json:"id"`
	Value float64 `json:"value"`
}

type ConfigCreatureSize struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ConfigMonsterType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// --- Monster types (from monster-service, auth required) ---

// MonsterResponse wraps the paginated monster API response.
type MonsterResponse struct {
	Data []RawMonster `json:"data"`
}

// RawMonster is a monster entry from the DDB API.
type RawMonster struct {
	ID                        int             `json:"id"`
	Name                      string          `json:"name"`
	ArmorClass                int             `json:"armorClass"`
	ArmorClassDescription     string          `json:"armorClassDescription"`
	AverageHitPoints          int             `json:"averageHitPoints"`
	HitPointDice              HitPointDice    `json:"hitPointDice"`
	ChallengeRatingID         int             `json:"challengeRatingId"`
	SizeID                    int             `json:"sizeId"`
	TypeID                    int             `json:"typeId"`
	Tags                      []string        `json:"tags"`
	Stats                     []MonsterStat   `json:"stats"`
	Movements                 []MonsterMove   `json:"movements"`
	SensesHTML                string          `json:"sensesHtml"`
	SkillsHTML                string          `json:"skillsHtml"`
	LanguageDescription       string          `json:"languageDescription"`
	PassivePerception         int             `json:"passivePerception"`
	SpecialTraitsDescription  string          `json:"specialTraitsDescription"`
	ActionsDescription        string          `json:"actionsDescription"`
	BonusActionsDescription   string          `json:"bonusActionsDescription"`
	ReactionsDescription      string          `json:"reactionsDescription"`
	LegendaryActionsDescription string       `json:"legendaryActionsDescription"`
	CharacteristicsDescription string         `json:"characteristicsDescription"`
	Sources                   []SourceRef     `json:"sources"`
	IsHomebrew                bool            `json:"isHomebrew"`
	IsReleased                bool            `json:"isReleased"`
	URL                       string          `json:"url"`
	RawJSON                   json.RawMessage `json:"-"` // populated after unmarshalling
}

type HitPointDice struct {
	DiceString string `json:"diceString"`
}

type MonsterStat struct {
	StatID int `json:"statId"`
	Value  int `json:"value"`
}

type MonsterMove struct {
	MovementID int `json:"movementId"`
	Speed      int `json:"speed"`
}

type SourceRef struct {
	SourceID int `json:"sourceId"`
}

// --- Spell types (from character-service, auth required) ---

// SpellResponse wraps the spell API response.
type SpellResponse struct {
	Data []RawSpellEntry `json:"data"`
}

// RawSpellEntry is a spell entry from the DDB API (definition nested).
type RawSpellEntry struct {
	ID         int          `json:"id"`
	Definition RawSpellDef  `json:"definition"`
}

type RawSpellDef struct {
	ID                    int             `json:"id"`
	Name                  string          `json:"name"`
	Level                 int             `json:"level"`
	School                string          `json:"school"`
	Activation            SpellActivation `json:"activation"`
	Components            []int           `json:"components"`
	ComponentsDescription string          `json:"componentsDescription"`
	Concentration         bool            `json:"concentration"`
	Ritual                bool            `json:"ritual"`
	Duration              SpellDuration   `json:"duration"`
	Range                 SpellRange      `json:"range"`
	Description           string          `json:"description"`
	Sources               []SourceRef     `json:"sources"`
	IsHomebrew            bool            `json:"isHomebrew"`
	Tags                  []string        `json:"tags"`
	RawJSON               json.RawMessage `json:"-"`
}

type SpellActivation struct {
	ActivationTime *int `json:"activationTime"`
	ActivationType int  `json:"activationType"`
}

type SpellDuration struct {
	DurationInterval int    `json:"durationInterval"`
	DurationType     string `json:"durationType"`
	DurationUnit     string `json:"durationUnit"`
}

type SpellRange struct {
	Origin     string `json:"origin"`
	RangeValue int    `json:"rangeValue"`
}

// --- Item types (from character-service, auth required) ---

// ItemResponse wraps the item API response.
type ItemResponse struct {
	Data []RawItem `json:"data"`
}

// RawItem is an item entry from the DDB API.
type RawItem struct {
	ID              int             `json:"id"`
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	FilterType      string          `json:"filterType"`
	Rarity          string          `json:"rarity"`
	CanAttune       bool            `json:"canAttune"`
	Magic           bool            `json:"magic"`
	Description     string          `json:"description"`
	Tags            []string        `json:"tags"`
	Sources         []SourceRef     `json:"sources"`
	IsHomebrew      bool            `json:"isHomebrew"`
	RawJSON         json.RawMessage `json:"-"`
}

// --- Auth types ---

// AuthResponse is the response from the cobalt token exchange.
type AuthResponse struct {
	Token string `json:"token"`
}

// --- User-data and entitlements (mobile API v6, cobalt-authenticated) ---

// UserData is the lightweight identity payload returned by /mobile/api/v6/user-data.
type UserData struct {
	Status          string `json:"status"`
	UserID          int    `json:"userId"`
	UserDisplayName string `json:"userDisplayName"`
}

// LicenseEntity is one product/book entry inside a License block.
type LicenseEntity struct {
	ID      int  `json:"id"`
	IsOwned bool `json:"isOwned"`
}

// LicenseBlock groups entitlements by EntityTypeID. Books are 496802664.
// DDB returns EntityTypeID as a JSON number, not a string, so we keep the
// Go field typed as int to avoid decode errors.
type LicenseBlock struct {
	EntityTypeID int             `json:"EntityTypeID"`
	Entities     []LicenseEntity `json:"Entities"`
}

// AvailableUserContent wraps the response from /mobile/api/v6/available-user-content.
//
// Some /mobile/api/v6 endpoints place their payload at the top level
// (`{status, Licenses}`) and others nest it under `data`
// (`{status, data: {Licenses}}`). We accept both shapes — see Books().
type AvailableUserContent struct {
	Status   string         `json:"status"`
	Licenses []LicenseBlock `json:"Licenses"`
	Data     LicenseEnvelope `json:"data"`
}

// LicenseEnvelope is the inner `data` payload for /mobile/api/v6 endpoints
// that nest their license blocks. Kept as a named type so revive doesn't flag
// it as a nested struct.
type LicenseEnvelope struct {
	Licenses []LicenseBlock `json:"Licenses"`
}

// Books returns the license blocks regardless of whether the response wraps
// them in a `data` envelope.
func (a AvailableUserContent) Books() []LicenseBlock {
	if len(a.Licenses) > 0 {
		return a.Licenses
	}
	return a.Data.Licenses
}

// EntityTypeIDBooks identifies the book license block in AvailableUserContent.
// Other observed blocks: 701257905 (dice sets), 2103445194 (something else
// — needs investigation if we ever care about non-book entitlements).
const EntityTypeIDBooks = 496802664
