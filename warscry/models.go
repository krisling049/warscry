package warscry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Characteristic represents a non-negative game statistic
type Characteristic int

func NewCharacteristic(value int) (Characteristic, error) {
	if value < 0 {
		return 0, fmt.Errorf("characteristic must be >= 0, got %d", value)
	}
	return Characteristic(value), nil
}

func (c Characteristic) Int() int {
	return int(c)
}

func (c Characteristic) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(c))
}

func (c *Characteristic) UnmarshalJSON(data []byte) error {
	var val int
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	if val < 0 {
		return fmt.Errorf("characteristic must be >= 0, got %d", val)
	}
	*c = Characteristic(val)
	return nil
}

// Type aliases for clearer function signatures
type (
	FighterID   = string
	FighterName = string
	Runemark    = string
	AbilityID   = string
	AbilityName = string
	Description = string
)

type Ability struct {
	Id              AbilityID   `json:"_id"`
	Name            AbilityName `json:"name"`
	Type            string      `json:"cost"`
	FactionRunemark Runemark    `json:"warband"`
	Runemarks       []Runemark  `json:"runemarks"`
	Description     Description `json:"description"`
}

type Weapon struct {
	Runemark     Runemark       `json:"runemark"`
	MinimumRange Characteristic `json:"min_range"`
	MaximumRange Characteristic `json:"max_range"`
	Attacks      Characteristic `json:"attacks"`
	Strength     Characteristic `json:"strength"`
	DamageHit    Characteristic `json:"dmg_hit"`
	DamageCrit   Characteristic `json:"dmg_crit"`
}

type Fighter struct {
	Id              FighterID      `json:"_id"`
	Name            FighterName    `json:"name"`
	FactionRunemark Runemark       `json:"warband"`
	Runemarks       []Runemark     `json:"runemarks"`
	Subfaction      string         `json:"subfaction"`
	GrandAlliance   string         `json:"grand_alliance"`
	Movement        Characteristic `json:"movement"`
	Toughness       Characteristic `json:"toughness"`
	Wounds          Characteristic `json:"wounds"`
	Points          Characteristic `json:"points,omitempty"`
	Weapons         []Weapon       `json:"weapons"`
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
	// Subfactions  []Subfaction `json:"subfactions"`
}

type Subfaction struct {
	Runemark  string `json:"runemark"`
	Bladeborn bool   `json:"bladeborn"`
	HeroesAll bool   `json:"heroes_all"`
}

func GitLoad(url string) ([]byte, error) {
	resp, getErr := http.Get(url)
	if getErr != nil {
		return nil, getErr
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	closeErr := resp.Body.Close()
	if closeErr != nil {
		return nil, closeErr
	}
	if resp.StatusCode > 299 {
		return nil, errors.New(
			fmt.Sprintf("Response failed with status code: %d", resp.StatusCode),
		)
	}
	return body, nil
}

func (F *Fighters) FromGit() {
	url := "https://krisling049.github.io/warcry_data/fighters.json"
	data, dataErr := GitLoad(url)
	if dataErr != nil {
		log.Fatalf("error loading fighter data from %s -- %s", url, dataErr)
	}
	jsonErr := json.Unmarshal(data, &F)
	if jsonErr != nil {
		log.Fatalf("error unmarshalling fighter data -- %s", jsonErr)
	}

	// Validate all fighters
	for i, fighter := range *F {
		if err := fighter.Validate(); err != nil {
			log.Fatalf("invalid fighter data at index %d: %v", i, err)
		}
	}
	log.Printf("loaded and validated %d fighters", len(*F))
}

func (A *Abilities) FromGit() {
	url := "https://krisling049.github.io/warcry_data/abilities_battletraits.json"
	data, dataErr := GitLoad(url)
	if dataErr != nil {
		log.Fatalf("error loading ability data from %s -- %s", url, dataErr)
	}
	jsonErr := json.Unmarshal(data, &A)
	if jsonErr != nil {
		log.Fatalf("error unmarshalling ability data -- %s", jsonErr)
	}

	// Validate all abilities
	for i, ability := range *A {
		if err := ability.Validate(); err != nil {
			log.Fatalf("invalid ability data at index %d: %v", i, err)
		}
	}
	log.Printf("loaded and validated %d abilities", len(*A))
}

// Validate checks if a Fighter has valid data
func (f *Fighter) Validate() error {
	if f.Id == "" {
		return errors.New("fighter ID cannot be empty")
	}
	if f.Name == "" {
		return errors.New("fighter name cannot be empty")
	}
	if f.FactionRunemark == "" {
		return errors.New("fighter faction runemark cannot be empty")
	}
	// Characteristic type guarantees non-negative, but we still check wounds > 0
	if f.Wounds.Int() <= 0 {
		return fmt.Errorf("fighter '%s' has invalid wounds: %d", f.Name, f.Wounds.Int())
	}
	if len(f.Weapons) == 0 {
		return fmt.Errorf("fighter '%s' has no weapons", f.Name)
	}

	for i, weapon := range f.Weapons {
		if err := weapon.Validate(); err != nil {
			return fmt.Errorf("fighter '%s' weapon %d: %w", f.Name, i, err)
		}
	}
	return nil
}

// Validate checks if a Weapon has valid data
func (w *Weapon) Validate() error {
	if w.Runemark == "" {
		return errors.New("weapon runemark cannot be empty")
	}
	// Characteristic type guarantees non-negative
	if w.MinimumRange.Int() > w.MaximumRange.Int() {
		return fmt.Errorf("weapon min range (%d) exceeds max range (%d)",
			w.MinimumRange.Int(), w.MaximumRange.Int())
	}
	if w.Attacks.Int() <= 0 {
		return fmt.Errorf("weapon attacks must be positive: %d", w.Attacks.Int())
	}
	return nil
}

// Validate checks if an Ability has valid data
func (a *Ability) Validate() error {
	if a.Id == "" {
		return errors.New("ability ID cannot be empty")
	}
	if a.Name == "" {
		return errors.New("ability name cannot be empty")
	}
	if a.FactionRunemark == "" {
		return errors.New("ability faction runemark cannot be empty")
	}
	if a.Type == "" {
		return errors.New("ability type cannot be empty")
	}
	return nil
}
