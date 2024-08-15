package warcry_go

import "net/http"

func (A *Abilities) GetWarband(factionRunemark string) *Abilities {
	warband := Abilities{}
	for _, a := range *A {
		if a.FactionRunemark == factionRunemark {
			warband = append(warband, a)
		}
	}
	return &warband
}

func (A *Abilities) GetIds() []string {
	var Ids []string
	for _, a := range *A {
		Ids = append(Ids, a.Id)
	}
	return Ids
}

func (a *Ability) MatchesRequest(r *http.Request) (bool, error) {
	var conditions []bool

	// ability characteristics
	conditions = append(conditions, StringInclude(a.Id, r.Form["_id"]))
	conditions = append(conditions, StringInclude(a.Name, r.Form["name"]))
	conditions = append(conditions, StringInclude(a.FactionRunemark, r.Form["warband"]))
	conditions = append(conditions, StringInclude(a.Type, r.Form["cost"]))
	conditions = append(conditions, SubStringInclude(a.Description, r.Form["description"]))
	conditions = append(conditions, StringSliceInclude(a.Runemarks, r.Form["runemarks"]))

	if All(conditions) {
		return true, nil
	}
	return false, nil
}
