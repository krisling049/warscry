package main

type Ability struct {
	Id              string   `json:"_id"`
	Name            string   `json:"name"`
	Type            string   `json:"cost"`
	FactionRunemark string   `json:"warband"`
	Runemarks       []string `json:"runemarks"`
	Description     string   `json:"description"`
}

type Weapon struct {
	Runemark     string `json:"runemark,omitempty"`
	MinimumRange int    `json:"min_range,omitempty"`
	MaximumRange int    `json:"max_range,omitempty"`
	Attacks      int    `json:"attacks,omitempty"`
	Strength     int    `json:"strength,omitempty"`
	DamageHit    int    `json:"dmg_hit,omitempty"`
	DamageCrit   int    `json:"dmg_crit,omitempty"`
}

type Fighter struct {
	Id                string   `json:"_id,omitempty"`
	Name              string   `json:"name,omitempty"`
	FactionRunemark   string   `json:"warband,omitempty"`
	Runemarks         []string `json:"runemarks,omitempty"`
	BladebornRunemark string   `json:"bladeborn,omitempty"`
	GrandAlliance     string   `json:"grand_alliance,omitempty"`
	Movement          int      `json:"movement,omitempty"`
	Toughness         int      `json:"toughness,omitempty"`
	Wounds            int      `json:"wounds,omitempty"`
	Points            int      `json:"points,omitempty"`
	Weapons           []Weapon `json:"weapons"`
}

type Warband struct {
	Name         string    `json:"name"`
	Fighters     Fighters  `json:"fighters,omitempty"`
	Abilities    Abilities `json:"abilities,omitempty"`
	BattleTraits Abilities `json:"battle_traits,omitempty"`
}

type (
	Fighters  []Fighter
	Abilities []Ability
	Warbands  []Warband
)
