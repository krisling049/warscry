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
			op, opErr := GetOperator(toCheck)
			if opErr != nil {
				log.Printf("%s - invalid operator for movement: %v", f.Name, opErr)
				continue
			}
			mv, mErr := IntInclude(f.Movement.Int(), r.Form[toCheck], op)
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
			op, opErr := GetOperator(toCheck)
			if opErr != nil {
				log.Printf("%s - invalid operator for wounds: %v", f.Name, opErr)
				continue
			}
			wo, wErr := IntInclude(f.Wounds.Int(), r.Form[toCheck], op)
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
			op, opErr := GetOperator(toCheck)
			if opErr != nil {
				log.Printf("%s - invalid operator for points: %v", f.Name, opErr)
				continue
			}
			pt, pErr := IntInclude(f.Points.Int(), r.Form[toCheck], op)
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
			op, opErr := GetOperator(toCheck)
			if opErr != nil {
				log.Printf("%s - invalid operator for toughness: %v", f.Name, opErr)
				continue
			}
			to, tErr := IntInclude(f.Toughness.Int(), r.Form[toCheck], op)
			if tErr != nil {
				log.Printf("%s - error while querying fighter\n%e", f.Name, tErr)
			} else {
				conditions = append(conditions, to)
			}
		}
	}

	// weapon characteristics
	var weaponConditions []bool
	for _, weapon := range f.Weapons {
		weaponInclude, weaponErr := weapon.MatchesRequest(r)
		if weaponErr != nil {
			log.Printf("%s - error while querying weapon: %v", f.Name, weaponErr)
			continue
		}
		weaponConditions = append(weaponConditions, weaponInclude)
	}

	// If no weapons, or at least one weapon matches the query, include this fighter
	if len(weaponConditions) == 0 {
		conditions = append(conditions, true)
	} else {
		conditions = append(conditions, Any(weaponConditions))
	}

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
			op, opErr := GetOperator(toCheck)
			if opErr != nil {
				return false, opErr
			}
			a, aErr := IntInclude(weapon.Attacks.Int(), r.Form[toCheck], op)
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
			op, opErr := GetOperator(toCheck)
			if opErr != nil {
				return false, opErr
			}
			s, sErr := IntInclude(weapon.Strength.Int(), r.Form[toCheck], op)
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
			op, opErr := GetOperator(toCheck)
			if opErr != nil {
				return false, opErr
			}
			dh, dhErr := IntInclude(weapon.DamageHit.Int(), r.Form[toCheck], op)
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
			op, opErr := GetOperator(toCheck)
			if opErr != nil {
				return false, opErr
			}
			dc, dcErr := IntInclude(weapon.DamageCrit.Int(), r.Form[toCheck], op)
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
			op, opErr := GetOperator(toCheck)
			if opErr != nil {
				return false, opErr
			}
			mar, marErr := IntInclude(weapon.MaximumRange.Int(), r.Form[toCheck], op)
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
			op, opErr := GetOperator(toCheck)
			if opErr != nil {
				return false, opErr
			}
			mir, mirErr := IntInclude(weapon.MinimumRange.Int(), r.Form[toCheck], op)
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
