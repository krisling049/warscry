package warscry

import (
	"fmt"
	"log"
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

func (f *Fighter) MatchesRequest(r *http.Request, c chan<- Fighter) {
	var conditions []bool
	var toCheck string
	operatorKeys := []string{"__gt", "__gte", "__lt", "__lte", ""}

	// fighter string characteristics
	conditions = append(conditions, StringInclude(f.Name, r.Form["name"]))
	conditions = append(conditions, StringInclude(f.Id, r.Form["_id"]))
	conditions = append(conditions, StringInclude(f.Subfaction, r.Form["subfaction"]))
	conditions = append(conditions, StringInclude(f.GrandAlliance, r.Form["grand_alliance"]))
	conditions = append(conditions, StringInclude(f.FactionRunemark, r.Form["warband"]))

	// fighter []string characteristics
	conditions = append(conditions, StringSliceInclude(f.Runemarks, r.Form["runemarks"]))

	// fighter int characteristics
	for _, key := range operatorKeys {
		toCheck = fmt.Sprintf("movement%s", key)
		if r.Form[toCheck] != nil {
			mv, mErr := IntInclude(f.Movement, r.Form[toCheck], GetOperator(toCheck))
			if mErr != nil {
				log.Printf("%s - error while querying fighter\n%e", f.Name, mErr)
			} else {
				conditions = append(conditions, mv)
			}
		}
	}

	for _, key := range operatorKeys {
		toCheck = fmt.Sprintf("wounds%s", key)
		if r.Form[toCheck] != nil {
			wo, wErr := IntInclude(f.Wounds, r.Form[toCheck], GetOperator(key))
			if wErr != nil {
				log.Printf("%s - error while querying fighter\n%e", f.Name, wErr)
			} else {
				conditions = append(conditions, wo)
			}
		}
	}

	for _, key := range operatorKeys {
		toCheck = fmt.Sprintf("points%s", key)
		if r.Form[toCheck] != nil {
			pt, pErr := IntInclude(f.Points, r.Form[toCheck], GetOperator(key))
			if pErr != nil {
				log.Printf("%s - error while querying fighter\n%e", f.Name, pErr)
			} else {
				conditions = append(conditions, pt)
			}
		}
	}

	for _, key := range operatorKeys {
		toCheck = fmt.Sprintf("toughness%s", key)
		if r.Form[toCheck] != nil {
			to, tErr := IntInclude(f.Toughness, r.Form[toCheck], GetOperator(key))
			if tErr != nil {
				log.Printf("%s - error while querying fighter\n%e", f.Name, tErr)
			} else {
				conditions = append(conditions, to)
			}
		}
	}

	// weapon characteristics
	weapon1Include, weapon1Err := f.Weapons[0].MatchesRequest(r)
	weaponConditions := []bool{weapon1Include}
	if weapon1Err != nil {
		log.Printf("%s - error while querying fighter\n%e", f.Name, weapon1Err)
	}
	if len(f.Weapons) == 2 {
		weapon2Include, weapon2Err := f.Weapons[1].MatchesRequest(r)
		if weapon2Err != nil {
			log.Printf("%s - error while querying fighter\n%e", f.Name, weapon2Err)
		}
		weaponConditions = append(weaponConditions, weapon2Include)
	}

	conditions = append(conditions, Any(weaponConditions))

	if All(conditions) {
		c <- *f
	}
}

func (weapon *Weapon) MatchesRequest(r *http.Request) (bool, error) {
	var conditions []bool

	// weapon []string characteristics
	conditions = append(conditions, StringInclude(weapon.Runemark, r.Form["weapon_runemark"]))

	// weapon int characteristics
	for _, key := range []string{"attacks", "attacks__gt", "attacks__lt"} {
		a, aErr := IntInclude(weapon.Attacks, r.Form[key], GetOperator(key))
		if aErr != nil {
			return false, aErr
		} else {
			conditions = append(conditions, a)
		}
	}

	for _, key := range []string{"strength", "strength__gt", "strength__lt"} {
		s, sErr := IntInclude(weapon.Strength, r.Form[key], GetOperator(key))
		if sErr != nil {
			return false, sErr
		} else {
			conditions = append(conditions, s)
		}
	}

	for _, key := range []string{"dmg_hit", "dmg_hit__gt", "dmg_hit__lt"} {
		d, dErr := IntInclude(weapon.DamageHit, r.Form[key], GetOperator(key))
		if dErr != nil {
			return false, dErr
		} else {
			conditions = append(conditions, d)
		}
	}

	for _, key := range []string{"dmg_crit", "dmg_crit__gt", "dmg_crit__lt"} {
		dc, dcErr := IntInclude(weapon.DamageHit, r.Form[key], GetOperator(key))
		if dcErr != nil {
			return false, dcErr
		} else {
			conditions = append(conditions, dc)
		}
	}

	for _, key := range []string{"max_range", "max_range__gt", "max_range__lt"} {
		maxRange, maxrErr := IntInclude(weapon.DamageHit, r.Form[key], GetOperator(key))
		if maxrErr != nil {
			return false, maxrErr
		} else {
			conditions = append(conditions, maxRange)
		}
	}

	for _, key := range []string{"min_range", "min_range__gt", "min_range__lt"} {
		minRange, minErr := IntInclude(weapon.DamageHit, r.Form[key], GetOperator(key))
		if minErr != nil {
			return false, minErr
		} else {
			conditions = append(conditions, minRange)
		}
	}

	if All(conditions) {
		return true, nil
	}
	return false, nil
}
