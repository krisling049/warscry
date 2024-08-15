package main

import (
	"fmt"
	wcd "github.com/krisling049/warcry_go"
	"log"
	"net/http"
	"os"
)

var (
	AllFighters  = wcd.Fighters{}
	AllAbilities = wcd.Abilities{}
	AllWarbands  = wcd.Warbands{}
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
	mux.Handle("/abilities", &wcd.AbilityHandler{Abilities: AllAbilities})

	// Run the server
	serveErr := http.ListenAndServe(GetPort(), mux)
	if serveErr != nil {
		log.Fatalln(serveErr)
	}

	fmt.Println("done")
}
