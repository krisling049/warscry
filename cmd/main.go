package main

import (
	"fmt"
	wcd "github.com/krisling049/warcry_go"
	"log"
	"net/http"
)

var (
	AllFighters  = wcd.Fighters{}
	AllAbilities = wcd.Abilities{}
	AllWarbands  = wcd.Warbands{}
)

func main() {
	AllFighters.FromGit()
	AllAbilities.FromGit()

	AllWarbands = *wcd.LoadWarbands(&AllFighters, &AllAbilities)
	if AllWarbands != nil {
		log.Println("data loaded")
	}

	mux := http.NewServeMux()

	// Register the routes and handlers
	mux.Handle("/", &wcd.RootHandler{
		Content: fmt.Sprintf("Welcome to Warscry!\nWarbands: %v\nFighters: %v\nAbilities: %v",
			len(AllWarbands), len(AllFighters), len(AllAbilities),
		)},
	)
	mux.Handle("/fighters", &wcd.FighterHandler{Fighters: AllFighters})

	// Run the server
	serveErr := http.ListenAndServe(":4424", mux)
	if serveErr != nil {
		log.Fatalln(serveErr)
	}

	fmt.Println("done")
}
