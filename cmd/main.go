package main

import (
	"fmt"
	"github.com/krisling049/warscry/warscry"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func GetPort() string {
	var port string
	port = os.Getenv("WARSCRY_PORT")
	if port == "" {
		port = "4424"
	}
	return fmt.Sprintf(":%s", port)
}

// GetPollInterval reads WARSCRY_POLL_INTERVAL env var
// Returns 0 to disable refresh, otherwise duration in minutes
func GetPollInterval() time.Duration {
	intervalStr := os.Getenv("WARSCRY_POLL_INTERVAL")
	if intervalStr == "" {
		return 30 * time.Minute // default: 30 minutes
	}
	if intervalStr == "0" || intervalStr == "disabled" {
		return 0 // disable refresh
	}

	minutes, err := strconv.Atoi(intervalStr)
	if err != nil || minutes < 1 {
		log.Printf("WARNING: invalid WARSCRY_POLL_INTERVAL '%s', using default (30 min)", intervalStr)
		return 30 * time.Minute
	}

	return time.Duration(minutes) * time.Minute
}

func main() {
	// Create data store
	dataStore := warscry.NewDataStore()

	// Initial load (fatal on error - cannot start without data)
	fighters := warscry.Fighters{}
	abilities := warscry.Abilities{}

	fighters.FromGit()
	abilities.FromGit()
	warbands := warscry.LoadWarbands(&fighters, &abilities)

	dataStore.LoadData(&fighters, &abilities, warbands)

	if len(fighters) == 0 || len(abilities) == 0 {
		log.Fatalln("initial data load failed")
	}
	log.Println("initial data loaded")

	// Start refresh loop
	refreshConfig := warscry.NewRefreshConfig(dataStore)
	pollInterval := GetPollInterval()
	if pollInterval > 0 {
		refreshConfig.PollInterval = pollInterval
		refreshConfig.StartRefreshLoop()
		defer refreshConfig.StopRefreshLoop()
		log.Printf("data refresh enabled (interval: %v)", pollInterval)
	} else {
		log.Println("data refresh disabled")
	}

	mux := http.NewServeMux()

	// Register the routes and handlers
	mux.Handle("/", &warscry.RootHandler{
		Version:   "v0.2.0",
		DataStore: dataStore,
		DocsURL:   "https://github.com/krisling049/warscry/blob/main/openapi.yaml",
	})
	mux.Handle("/fighters", &warscry.FighterHandler{DataStore: dataStore})
	mux.Handle("/abilities", &warscry.AbilityHandler{DataStore: dataStore})
	mux.Handle("/health", &warscry.HealthHandler{DataStore: dataStore})

	// Run the server
	serveErr := http.ListenAndServe(GetPort(), mux)
	if serveErr != nil {
		log.Fatalln(serveErr)
	}

	fmt.Println("done")
}
