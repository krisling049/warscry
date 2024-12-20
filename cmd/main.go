package main

import (
	"fmt"
	"github.com/krisling049/warcry_go/warscry"
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
		Content: fmt.Sprintf("Welcome to Warscry!\nWarbands: %v\nFighters: %v\nAbilities: %v",
			len(AllWarbands), len(AllFighters), len(AllAbilities),
		)},
	)
	mux.Handle("/fighters", &warscry.FighterHandler{Fighters: AllFighters})
	mux.Handle("/abilities", &warscry.AbilityHandler{Abilities: AllAbilities})

	// Run the server
	serveErr := http.ListenAndServe(GetPort(), mux)
	if serveErr != nil {
		log.Fatalln(serveErr)
	}

	fmt.Println("done")
}
