package main

import (
	"errors"
	"log"
	"slices"
)

func (W *Warband) AddFighter(f *Fighter) error {
	if W.Fighters == nil {
		W.Fighters = Fighters{}
	}
	if slices.Contains(W.Fighters.GetIds(), f.Id) {
		err := errors.New("a Fighter with this Id already exists")
		return err
	}
	W.Fighters = append(W.Fighters, *f)
	return nil
}

func (W *Warband) AddAbility(a *Ability) error {
	if W.Abilities == nil {
		W.Abilities = Abilities{}
	}
	if slices.Contains(W.Abilities.GetIds(), a.Id) {
		err := errors.New("an Ability with this Id already exists")
		return err
	}
	W.Abilities = append(W.Abilities, *a)
	return nil
}

func LoadWarbands(F *Fighters, A *Abilities) *Warbands {
	var wbslice []string
	wbs := make(map[string]Warband)
	bladeborn := make(map[string]string)
	newW := Warbands{}

	for _, f := range *F {
		wb, _ := wbs[f.FactionRunemark]
		err := wb.AddFighter(&f)
		if err != nil {
			log.Fatalf("error while adding fighter -- %s", err)
		}
		wbs[f.FactionRunemark] = wb
		if !slices.Contains(wbslice, f.FactionRunemark) {
			wbslice = append(wbslice, f.FactionRunemark)
		}
		if f.BladebornRunemark != "" {
			bladeborn[f.BladebornRunemark] = f.FactionRunemark
		}
	}
	for _, a := range *A {
		faction := a.FactionRunemark
		bbornfaction, _ := bladeborn[a.FactionRunemark]
		if bbornfaction != "" {
			faction = bbornfaction
		}
		wb, _ := wbs[faction]
		if a.Type == "battle_trait" {
			wb.BattleTraits = append(wb.BattleTraits, a)
		} else {
			err := wb.AddAbility(&a)
			if err != nil {
				log.Fatalf("error while adding ability -- %s", err)
			}
		}
		wbs[faction] = wb
		if !slices.Contains(wbslice, faction) {
			wbslice = append(wbslice, faction)
		}
	}
	for _, name := range wbslice {
		toAdd := wbs[name]
		toAdd.Name = name
		newW = append(newW, toAdd)
	}
	return &newW
}
