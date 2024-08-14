package warcry_go

type Ability struct {
	Id              string   `json:"_id"`
	Name            string   `json:"name"`
	Type            string   `json:"cost"`
	FactionRunemark string   `json:"warband"`
	Runemarks       []string `json:"runemarks"`
	Description     string   `json:"description"`
}

type Weapon struct {
	Runemark     string `json:"runemark"`
	MinimumRange int    `json:"min_range"`
	MaximumRange int    `json:"max_range"`
	Attacks      int    `json:"attacks"`
	Strength     int    `json:"strength"`
	DamageHit    int    `json:"dmg_hit"`
	DamageCrit   int    `json:"dmg_crit"`
}

type Fighter struct {
	Id                string   `json:"_id"`
	Name              string   `json:"name"`
	FactionRunemark   string   `json:"warband"`
	Runemarks         []string `json:"runemarks"`
	BladebornRunemark string   `json:"bladeborn,omitempty"`
	GrandAlliance     string   `json:"grand_alliance"`
	Movement          int      `json:"movement"`
	Toughness         int      `json:"toughness"`
	Wounds            int      `json:"wounds"`
	Points            int      `json:"points,omitempty"`
	Weapons           []Weapon `json:"weapons"`
}

type (
	Fighters  []Fighter
	Abilities []Ability
	Warbands  []Warband
)

type Warband struct {
	Name         string    `json:"name"`
	Fighters     Fighters  `json:"fighters"`
	Abilities    Abilities `json:"abilities"`
	BattleTraits Abilities `json:"battle_traits"`
}
