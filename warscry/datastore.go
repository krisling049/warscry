package warscry

import (
	"sync/atomic"
)

// DataStore holds atomic pointers to current data collections
// Thread-safe for concurrent reads and atomic updates
type DataStore struct {
	fighters  atomic.Pointer[Fighters]
	abilities atomic.Pointer[Abilities]
	warbands  atomic.Pointer[Warbands]
}

// NewDataStore creates an empty data store
func NewDataStore() *DataStore {
	return &DataStore{}
}

// LoadData atomically replaces all data collections
// Safe to call while handlers are reading
func (ds *DataStore) LoadData(fighters *Fighters, abilities *Abilities, warbands *Warbands) {
	ds.fighters.Store(fighters)
	ds.abilities.Store(abilities)
	ds.warbands.Store(warbands)
}

// GetFighters returns current fighter collection (never nil after initial load)
func (ds *DataStore) GetFighters() Fighters {
	ptr := ds.fighters.Load()
	if ptr == nil {
		return Fighters{}
	}
	return *ptr
}

// GetAbilities returns current ability collection (never nil after initial load)
func (ds *DataStore) GetAbilities() Abilities {
	ptr := ds.abilities.Load()
	if ptr == nil {
		return Abilities{}
	}
	return *ptr
}

// GetWarbands returns current warband collection (never nil after initial load)
func (ds *DataStore) GetWarbands() Warbands {
	ptr := ds.warbands.Load()
	if ptr == nil {
		return Warbands{}
	}
	return *ptr
}

// GetCounts returns fighter and ability counts
func (ds *DataStore) GetCounts() (fighterCount, abilityCount int) {
	fighters := ds.GetFighters()
	abilities := ds.GetAbilities()
	return len(fighters), len(abilities)
}
