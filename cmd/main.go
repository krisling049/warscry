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
	var rootContent string

	AllFighters.FromGit()
	AllAbilities.FromGit()

	rootContent = "Welcome to Warcry!\n" +
		fmt.Sprintf("Use /fighters to view %v fighters.\n", len(AllFighters)) +
		"Fighter characteristics can be queried using ?characteristic=value, e.g. /fighters?attacks=4\n" +
		"Append operators (__gt, __gte, __lt, __lte) to the characteristic for greater than, greater than or equal to etc\n" +
		"e.g. /fighters?attacks__gte=4 for all fighters with 4 or more attacks\n\n" +
		fmt.Sprintf("Use /abilities to view %v abilities.\n", len(AllAbilities)) +
		"Query ability characteristics in the same way (no operators supported currently).\n" +
		"description=word will find any ability with that word/substring in its description."

	AllWarbands = *warscry.LoadWarbands(&AllFighters, &AllAbilities)
	if AllWarbands != nil {
		log.Println("data loaded")
	}

	mux := http.NewServeMux()

	// Register the routes and handlers
	mux.Handle("/", &warscry.RootHandler{
		Content: rootContent,
	},
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
