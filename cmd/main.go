package main

import (
	"fmt"
	"github.com/krisling049/warscry/warscry"
	"log"
	"net/http"
	"os"
)

var (
	AllFighters  = warscry.Fighters{}
	AllAbilities = warscry.Abilities{}
	AllWarbands  = warscry.Warbands{}
)

func GetPort() string {
	var port string
	port = os.Getenv("WARSCRY_PORT")
	if port == "" {
		port = "4424"
	}
	return fmt.Sprintf(":%s", port)
}

func main() {
	AllFighters.FromGit()
	AllAbilities.FromGit()

	AllWarbands = *warscry.LoadWarbands(&AllFighters, &AllAbilities)
	if AllWarbands != nil {
		log.Println("data loaded")
	}

	mux := http.NewServeMux()

	// Register the routes and handlers
	mux.Handle("/", &warscry.RootHandler{
		Version:      "v0.1.0",
		FighterCount: len(AllFighters),
		AbilityCount: len(AllAbilities),
		DocsURL:      "https://github.com/krisling049/warscry/blob/main/openapi.yaml",
	})
	mux.Handle("/fighters", &warscry.FighterHandler{Fighters: AllFighters})
	mux.Handle("/abilities", &warscry.AbilityHandler{Abilities: AllAbilities})
	mux.Handle("/health", &warscry.HealthHandler{
		FighterCount: len(AllFighters),
		AbilityCount: len(AllAbilities),
	})

	// Run the server
	serveErr := http.ListenAndServe(GetPort(), mux)
	if serveErr != nil {
		log.Fatalln(serveErr)
	}

	fmt.Println("done")
}
