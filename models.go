package warcry_go

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

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
	Id              string   `json:"_id"`
	Name            string   `json:"name"`
	FactionRunemark string   `json:"warband"`
	Runemarks       []string `json:"runemarks"`
	Subfaction      string   `json:"subfaction"`
	GrandAlliance   string   `json:"grand_alliance"`
	Movement        int      `json:"movement"`
	Toughness       int      `json:"toughness"`
	Wounds          int      `json:"wounds"`
	Points          int      `json:"points,omitempty"`
	Weapons         []Weapon `json:"weapons"`
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
		log.Fatalf("error marshalling fighter data -- %s", jsonErr)
	}
}

func (A *Abilities) FromGit() {
	url := "https://krisling049.github.io/warcry_data/abilities_battletraits.json"
	data, dataErr := GitLoad(url)
	if dataErr != nil {
		log.Fatalf("error loading ability data from %s -- %s", url, dataErr)
	}
	jsonErr := json.Unmarshal(data, &A)
	if jsonErr != nil {
		log.Fatalf("error marshalling ability data -- %s", jsonErr)
	}
}
