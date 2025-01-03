package warscry

import (
	"fmt"
	"log"
	"net/http"
)

var operatorKeys = []string{"__gt", "__gte", "__lt", "__lte", ""}

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
	var (
		conditions []bool
		toCheck    string
	)

	// weapon []string characteristics
	conditions = append(conditions, StringInclude(weapon.Runemark, r.Form["weapon_runemark"]))

	// weapon int characteristics
	for _, key := range operatorKeys {
		toCheck = fmt.Sprintf("attacks%s", key)
		if r.Form[toCheck] != nil {
			a, aErr := IntInclude(weapon.Attacks, r.Form[toCheck], GetOperator(key))
			if aErr != nil {
				return false, aErr
			} else {
				conditions = append(conditions, a)
			}
		}
	}

	for _, key := range operatorKeys {
		toCheck = fmt.Sprintf("strength%s", key)
		if r.Form[toCheck] != nil {
			s, sErr := IntInclude(weapon.Strength, r.Form[toCheck], GetOperator(key))
			if sErr != nil {
				return false, sErr
			} else {
				conditions = append(conditions, s)
			}
		}
	}

	for _, key := range operatorKeys {
		toCheck = fmt.Sprintf("dmg_hit%s", key)
		if r.Form[toCheck] != nil {
			dh, dhErr := IntInclude(weapon.DamageHit, r.Form[toCheck], GetOperator(key))
			if dhErr != nil {
				return false, dhErr
			} else {
				conditions = append(conditions, dh)
			}
		}
	}

	for _, key := range operatorKeys {
		toCheck = fmt.Sprintf("dmg_crit%s", key)
		if r.Form[toCheck] != nil {
			dc, dcErr := IntInclude(weapon.DamageCrit, r.Form[toCheck], GetOperator(key))
			if dcErr != nil {
				return false, dcErr
			} else {
				conditions = append(conditions, dc)
			}
		}
	}

	for _, key := range operatorKeys {
		toCheck = fmt.Sprintf("max_range%s", key)
		if r.Form[toCheck] != nil {
			mar, marErr := IntInclude(weapon.MaximumRange, r.Form[toCheck], GetOperator(key))
			if marErr != nil {
				return false, marErr
			} else {
				conditions = append(conditions, mar)
			}
		}
	}

	for _, key := range operatorKeys {
		toCheck = fmt.Sprintf("min_range%s", key)
		if r.Form[toCheck] != nil {
			mir, mirErr := IntInclude(weapon.MinimumRange, r.Form[toCheck], GetOperator(key))
			if mirErr != nil {
				return false, mirErr
			} else {
				conditions = append(conditions, mir)
			}
		}
	}

	if All(conditions) {
		return true, nil
	}
	return false, nil
}
