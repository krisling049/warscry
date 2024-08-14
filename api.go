package warcry_go

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type FighterHandler struct {
	Fighters Fighters
}

func All(s []bool) bool {
	var trues int

	for _, b := range s {
		if b {
			trues += 1
		}
	}
	if len(s) == trues {
		return true
	}
	return false
}

func StringInclude(characteristic string, values []string) bool {
	var (
		lowerValues []string
		Include     = false
	)
	if len(values) < 1 {
		Include = true
		return Include
	}

	for _, n := range values {
		lowerValues = append(lowerValues, strings.ToLower(n))
	}

	if slices.Contains(lowerValues, strings.ToLower(characteristic)) {
		Include = true
	}

	return Include
}

func IntInclude(characteristic int, values []string) (bool, error) {
	var (
		Include   = false
		intValues []int
	)
	if len(values) < 1 {
		Include = true
		return Include, nil
	}

	for _, s := range values {
		i, err := strconv.Atoi(s)
		if err != nil {
			return false, err
		}
		intValues = append(intValues, i)
	}

	if slices.Contains(intValues, characteristic) {
		Include = true
	}
	return Include, nil
}

func StringSliceInclude(characteristic []string, values []string) bool {
	var (
		Include     = false
		lowerValues []string
		conditions  []bool
	)
	if len(values) < 1 {
		Include = true
		return Include
	}

	for _, s := range values {
		lowerValues = append(lowerValues, strings.ToLower(s))
	}

	for _, v := range lowerValues {
		if slices.Contains(characteristic, v) {
			conditions = append(conditions, true)
		} else {
			conditions = append(conditions, false)
		}
	}

	if All(conditions) {
		Include = true
	}

	return Include
}

func (h *FighterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		successfulQuery = false
		response        []byte
	)
	perr := r.ParseForm()
	if perr != nil {
		response = []byte(fmt.Sprintf("failed to read request -- %s", perr))
	}

	var toRet Fighters
	var conditions []bool

	if len(r.Form) > 0 {
		for _, f := range h.Fighters {
			conditions = nil
			conditions = append(conditions, StringInclude(f.Name, r.Form["name"]))
			conditions = append(conditions, StringInclude(f.Id, r.Form["_id"]))
			conditions = append(conditions, StringInclude(f.BladebornRunemark, r.Form["bladeborn"]))
			conditions = append(conditions, StringInclude(f.GrandAlliance, r.Form["grand_alliance"]))
			conditions = append(conditions, StringInclude(f.FactionRunemark, r.Form["warband"]))
			mv, mErr := IntInclude(f.Movement, r.Form["movement"])
			if mErr != nil {
				response = []byte(fmt.Sprintf("an error occurred -- %s", mErr))
				break
			} else {
				conditions = append(conditions, mv)
			}
			wo, wErr := IntInclude(f.Wounds, r.Form["wounds"])
			if wErr != nil {
				response = []byte(fmt.Sprintf("an error occurred -- %s", wErr))
				break
			} else {
				conditions = append(conditions, wo)
			}
			pt, pErr := IntInclude(f.Points, r.Form["points"])
			if pErr != nil {
				response = []byte(fmt.Sprintf("an error occurred -- %s", pErr))
				break
			} else {
				conditions = append(conditions, pt)
			}
			to, tErr := IntInclude(f.Toughness, r.Form["toughness"])
			if tErr != nil {
				response = []byte(fmt.Sprintf("an error occurred -- %s", tErr))
				break
			} else {
				conditions = append(conditions, to)
			}
			conditions = append(conditions, StringSliceInclude(f.Runemarks, r.Form["runemarks"]))

			if All(conditions) {
				toRet = append(toRet, f)
			}
		}
	} else {
		toRet = append(toRet, h.Fighters...)
	}
	if len(response) == 0 {
		successfulQuery = true
	}
	//        "weapons": [
	//            {
	//                "attacks": 3,
	//                "dmg_crit": 3,
	//                "dmg_hit": 1,
	//                "max_range": 1,
	//                "min_range": 0,
	//                "runemark": "axe",
	//                "strength": 3
	//            }
	//        ],

	if successfulQuery {
		marshalledResponse, err := json.Marshal(toRet)
		if err != nil {
			response = []byte(fmt.Sprintf("an error occurred while getting the requested data -- %s", err))
		}
		response = marshalledResponse
	}
	w.Write(response)
}
