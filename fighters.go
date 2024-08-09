package main

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
