package warscry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	FightersURL  = "https://krisling049.github.io/warcry_data/fighters.json"
	AbilitiesURL = "https://krisling049.github.io/warcry_data/abilities_battletraits.json"
)

// RefreshConfig controls data refresh behavior
type RefreshConfig struct {
	PollInterval time.Duration
	DataStore    *DataStore
	StopChan     chan struct{}
}

// NewRefreshConfig creates default configuration
func NewRefreshConfig(dataStore *DataStore) *RefreshConfig {
	return &RefreshConfig{
		PollInterval: 30 * time.Minute,
		DataStore:    dataStore,
		StopChan:     make(chan struct{}),
	}
}

// RefreshState tracks ETags for conditional requests
type RefreshState struct {
	fightersETag  string
	abilitiesETag string
}

// StartRefreshLoop begins background polling (non-blocking)
// Returns immediately after starting goroutine
func (cfg *RefreshConfig) StartRefreshLoop() {
	go cfg.refreshLoop()
}

// StopRefreshLoop signals the background goroutine to stop
func (cfg *RefreshConfig) StopRefreshLoop() {
	close(cfg.StopChan)
}

// refreshLoop runs periodic ETag checks and data reloads
func (cfg *RefreshConfig) refreshLoop() {
	state := &RefreshState{}
	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	log.Printf("refresh loop started (interval: %v)", cfg.PollInterval)

	for {
		select {
		case <-ticker.C:
			cfg.checkAndRefresh(state)
		case <-cfg.StopChan:
			log.Println("refresh loop stopped")
			return
		}
	}
}

// checkAndRefresh performs ETag check and reloads if data changed
func (cfg *RefreshConfig) checkAndRefresh(state *RefreshState) {
	log.Println("checking for data updates...")

	// Check both sources
	fightersChanged, newFightersETag, fightersErr := cfg.hasChanged(FightersURL, state.fightersETag)
	abilitiesChanged, newAbilitiesETag, abilitiesErr := cfg.hasChanged(AbilitiesURL, state.abilitiesETag)

	// Handle errors (non-fatal, keep old data)
	if fightersErr != nil {
		log.Printf("ERROR: failed to check fighters ETag: %v", fightersErr)
		return
	}
	if abilitiesErr != nil {
		log.Printf("ERROR: failed to check abilities ETag: %v", abilitiesErr)
		return
	}

	// No changes detected
	if !fightersChanged && !abilitiesChanged {
		log.Println("no data changes detected")
		return
	}

	// At least one source changed - reload both
	log.Printf("data changed (fighters: %v, abilities: %v) - refreshing...",
		fightersChanged, abilitiesChanged)

	fighters, abilities, loadErr := cfg.loadDataSources()
	if loadErr != nil {
		log.Printf("ERROR: refresh failed: %v - keeping old data", loadErr)
		return
	}

	// Generate warbands from new data
	warbands := LoadWarbands(&fighters, &abilities)

	// Atomic update
	cfg.DataStore.LoadData(&fighters, &abilities, warbands)

	// Update state on success
	state.fightersETag = newFightersETag
	state.abilitiesETag = newAbilitiesETag

	fCount, aCount := cfg.DataStore.GetCounts()
	log.Printf("data refresh complete (%d fighters, %d abilities)", fCount, aCount)
}

// hasChanged performs HEAD request and compares ETag
// Returns: (changed bool, newETag string, error)
func (cfg *RefreshConfig) hasChanged(url string, currentETag string) (bool, string, error) {
	req, reqErr := http.NewRequest(http.MethodHead, url, nil)
	if reqErr != nil {
		return false, "", fmt.Errorf("failed to create request: %w", reqErr)
	}

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return false, "", fmt.Errorf("HEAD request failed: %w", respErr)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("WARNING: failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	newETag := resp.Header.Get("ETag")
	if newETag == "" {
		// No ETag header - assume changed to be safe
		return true, "", nil
	}

	// First check or ETag changed
	if currentETag == "" || currentETag != newETag {
		return true, newETag, nil
	}

	return false, newETag, nil
}

// loadDataSources fetches and validates fighters and abilities
func (cfg *RefreshConfig) loadDataSources() (Fighters, Abilities, error) {
	fighters := Fighters{}
	abilities := Abilities{}

	// Load fighters
	fightersData, fErr := gitLoad(FightersURL)
	if fErr != nil {
		return fighters, abilities, fmt.Errorf("load fighters: %w", fErr)
	}

	if err := json.Unmarshal(fightersData, &fighters); err != nil {
		return fighters, abilities, fmt.Errorf("unmarshal fighters: %w", err)
	}

	// Validate fighters
	for i, fighter := range fighters {
		if err := fighter.Validate(); err != nil {
			return fighters, abilities, fmt.Errorf("invalid fighter at index %d: %w", i, err)
		}
	}

	// Load abilities
	abilitiesData, aErr := gitLoad(AbilitiesURL)
	if aErr != nil {
		return fighters, abilities, fmt.Errorf("load abilities: %w", aErr)
	}

	if err := json.Unmarshal(abilitiesData, &abilities); err != nil {
		return fighters, abilities, fmt.Errorf("unmarshal abilities: %w", err)
	}

	// Validate abilities
	for i, ability := range abilities {
		if err := ability.Validate(); err != nil {
			return fighters, abilities, fmt.Errorf("invalid ability at index %d: %w", i, err)
		}
	}

	return fighters, abilities, nil
}

// gitLoad performs GET request (lowercase to avoid export, reuse pattern from models.go)
func gitLoad(url string) ([]byte, error) {
	resp, getErr := http.Get(url)
	if getErr != nil {
		return nil, getErr
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("WARNING: failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode > 299 {
		return nil, errors.New(fmt.Sprintf("response failed with status code: %d", resp.StatusCode))
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	return body, nil
}
