package main

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
