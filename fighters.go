package warcry_go

import (
	"net/http"
)

func (F *Fighters) GetWarband(factionRunemark string) *Fighters {
	warband := Fighters{}
	for _, f := range *F {
		if f.FactionRunemark == factionRunemark {
			warband = append(warband, f)
		}
	}
	return &warband
}

func (F *Fighters) GetIds() []string {
	var Ids []string
	for _, f := range *F {
		Ids = append(Ids, f.Id)
	}
	return Ids
}

func (f *Fighter) MatchesRequest(r *http.Request) (bool, error) {
	var conditions []bool
	// fighter characteristics
	conditions = append(conditions, StringInclude(f.Name, r.Form["name"]))
	conditions = append(conditions, StringInclude(f.Id, r.Form["_id"]))
	conditions = append(conditions, StringInclude(f.BladebornRunemark, r.Form["bladeborn"]))
	conditions = append(conditions, StringInclude(f.GrandAlliance, r.Form["grand_alliance"]))
	conditions = append(conditions, StringInclude(f.FactionRunemark, r.Form["warband"]))
	mv, mErr := IntInclude(f.Movement, r.Form["movement"])
	if mErr != nil {
		return false, mErr
	} else {
		conditions = append(conditions, mv)
	}
	wo, wErr := IntInclude(f.Wounds, r.Form["wounds"])
	if wErr != nil {
		return false, wErr
	} else {
		conditions = append(conditions, wo)
	}
	pt, pErr := IntInclude(f.Points, r.Form["points"])
	if pErr != nil {
		return false, pErr
	} else {
		conditions = append(conditions, pt)
	}
	to, tErr := IntInclude(f.Toughness, r.Form["toughness"])
	if tErr != nil {
		return false, tErr
	} else {
		conditions = append(conditions, to)
	}
	conditions = append(conditions, StringSliceInclude(f.Runemarks, r.Form["runemarks"]))

	// weapon characteristics
	weapon1Include, weapon1Err := f.Weapons[0].MatchesRequest(r)
	weaponConditions := []bool{weapon1Include}
	if weapon1Err != nil {
		return false, weapon1Err
	}
	if len(f.Weapons) == 2 {
		weapon2Include, weapon2Err := f.Weapons[1].MatchesRequest(r)
		if weapon2Err != nil {
			return false, weapon2Err
		}
		weaponConditions = append(weaponConditions, weapon2Include)
	}

	conditions = append(conditions, Any(weaponConditions))

	if All(conditions) {
		return true, nil
	}
	return false, nil
}

func (weapon *Weapon) MatchesRequest(r *http.Request) (bool, error) {
	var conditions []bool

	conditions = append(conditions, StringInclude(weapon.Runemark, r.Form["weapon_runemark"]))

	a, aErr := IntInclude(weapon.Attacks, r.Form["attacks"])
	if aErr != nil {
		return false, aErr
	}
	conditions = append(conditions, a)

	s, sErr := IntInclude(weapon.Strength, r.Form["strength"])
	if sErr != nil {
		return false, sErr
	}
	conditions = append(conditions, s)

	d, dErr := IntInclude(weapon.DamageHit, r.Form["dmg_hit"])
	if dErr != nil {
		return false, dErr
	}
	conditions = append(conditions, d)

	dc, dcErr := IntInclude(weapon.DamageCrit, r.Form["dmg_crit"])
	if dcErr != nil {
		return false, dcErr
	}
	conditions = append(conditions, dc)

	maxRange, maxrErr := IntInclude(weapon.MaximumRange, r.Form["max_range"])
	if maxrErr != nil {
		return false, maxrErr
	}
	conditions = append(conditions, maxRange)

	minRange, minErr := IntInclude(weapon.MinimumRange, r.Form["min_range"])
	if minErr != nil {
		return false, minErr
	}
	conditions = append(conditions, minRange)

	if All(conditions) {
		return true, nil
	}
	return false, nil
}
