package warscry

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
)

type FighterHandler struct {
	Fighters Fighters
}

type AbilityHandler struct {
	Abilities Abilities
}

type RootHandler struct {
	Content string
}

func (R *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("serving root response to %s", r.RemoteAddr)
	response := []byte(R.Content)
	_, err := w.Write(response)
	if err != nil {
		log.Printf("WARNING: failed to write response -- %s", err)
	}
}

func All(s []bool) bool {
	var trues int

	for _, b := range s {
		if b {
			trues += 1
		}
	}
	if len(s) == trues {
		return true
	}
	return false
}

func Any(s []bool) bool {
	if slices.Contains(s, true) {
		return true
	}
	return false
}

type Operator func(int, int) bool

func Equals(characteristic int, requested int) bool {
	if requested == characteristic {
		return true
	}
	return false
}

func GreaterThan(characteristic int, minimum int) bool {
	if characteristic > minimum {
		return true
	}
	return false
}

func GreaterThanOrEqualTo(characteristic int, minimum int) bool {
	if characteristic >= minimum {
		return true
	}
	return false
}

func LessThan(characteristic int, maximum int) bool {
	if characteristic < maximum {
		return true
	}
	return false
}

func LessThanOrEqualTo(characteristic int, maximum int) bool {
	if characteristic <= maximum {
		return true
	}
	return false
}

func GetOperator(queryKey string) Operator {
	OperatorMap := map[string]Operator{
		"gt":  GreaterThan,
		"gte": GreaterThanOrEqualTo,
		"lt":  LessThan,
		"lte": LessThanOrEqualTo,
	}
	if strings.Contains(queryKey, "__") {
		opString := strings.Split(queryKey, "_")
		return OperatorMap[opString[len(opString)-1]]
	}
	return Equals
}

func StringInclude(characteristic string, values []string) bool {
	var (
		lowerValues []string
		Include     = false
	)
	if len(values) < 1 {
		Include = true
		return Include
	}

	for _, n := range values {
		lowerValues = append(lowerValues, strings.ToLower(n))
	}

	if slices.Contains(lowerValues, strings.ToLower(characteristic)) {
		Include = true
	}

	return Include
}

func SubStringInclude(characteristic string, values []string) bool {
	var (
		lowerValues []string
		Include     = false
	)
	if len(values) < 1 {
		Include = true
		return Include
	}

	for _, n := range values {
		lowerValues = append(lowerValues, strings.ToLower(n))
	}

	for _, s := range lowerValues {
		if strings.Contains(strings.ToLower(characteristic), s) {
			Include = true
		}
		if Include {
			break
		}
	}

	return Include
}

func IntInclude(characteristic int, values []string, o Operator) (bool, error) {
	var (
		Include   = false
		intValues []int
	)
	if len(values) < 1 {
		Include = true
		return Include, nil
	}

	for _, s := range values {
		i, err := strconv.Atoi(s)
		if err != nil {
			return false, err
		}
		intValues = append(intValues, i)
	}

	for _, v := range intValues {
		if o(characteristic, v) {
			Include = true
		}
	}

	//if slices.Contains(intValues, characteristic) {
	//	Include = true
	//}
	return Include, nil
}

func StringSliceInclude(characteristic []string, values []string) bool {
	var (
		Include     = false
		lowerValues []string
		conditions  []bool
	)
	if len(values) < 1 {
		Include = true
		return Include
	}

	for _, s := range values {
		lowerValues = append(lowerValues, strings.ToLower(s))
	}

	for _, v := range lowerValues {
		if slices.Contains(characteristic, v) {
			conditions = append(conditions, true)
		} else {
			conditions = append(conditions, false)
		}
	}

	if All(conditions) {
		Include = true
	}

	return Include
}

func (h *FighterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		successfulQuery = false
		response        []byte
		toRet           Fighters
		wg              sync.WaitGroup
	)
	perr := r.ParseForm()
	if perr != nil {
		response = []byte(fmt.Sprintf("failed to read request -- %s", perr))
	}

	// Create a channel to recieve Fighters
	// Fighter.MatchesRequest() will send a Fighter into this channel if it meets the requirements of the request
	fChan := make(chan Fighter, len(h.Fighters))

	if len(r.Form) > 0 {
		// If criteria are specified in the form, we need to check which Fighters meet that criteria
		for i := range h.Fighters {
			// Add 1 to the WaitGroup for each fighter we check
			wg.Add(1)
			index := i
			go func() {
				// defer Done until MatchesRequest has completed
				defer wg.Done()
				h.Fighters[index].MatchesRequest(r, fChan)
			}()
		}
		// Wait until all Fighter.MatchesRequest() calls have completed
		wg.Wait()
		// Close the channel to indicate all Fighters have been checked for this request

	} else {
		// If not criteria have been specified, return all Fighters
		toRet = append(toRet, h.Fighters...)
	}

	close(fChan)
	// Loop over any fighters found in the fChan channel
	for includedFighter := range fChan {
		toRet = append(toRet, includedFighter)
	}
	// If len(response) is 0 that means we didn't write an error to it
	if len(response) == 0 {
		successfulQuery = true
	}

	if successfulQuery {
		log.Printf("returning %d fighters to %s", len(toRet), r.RemoteAddr)
		// Convert the Fighters to valid json
		marshalledResponse, err := json.Marshal(toRet)
		if err != nil {
			response = []byte(fmt.Sprintf("an error occurred while getting the requested data -- %s", err))
		}
		// Set the json of fighters as our response
		response = marshalledResponse
	}
	// Write our response ,either list of fighters or an error message, to be return to the requester
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(response)
	if writeErr != nil {
		log.Printf("WARNING: failed to write response -- %s", writeErr)
	}
}

func (h *AbilityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		successfulQuery = false
		response        []byte
	)
	perr := r.ParseForm()
	if perr != nil {
		response = []byte(fmt.Sprintf("failed to read request -- %s", perr))
	}

	var toRet Abilities

	if len(r.Form) > 0 {
		for _, a := range h.Abilities {
			include, err := a.MatchesRequest(r)
			if err != nil {
				response = []byte(fmt.Sprintf("an error occurred -- %s", err))
				break
			}
			if include {
				toRet = append(toRet, a)
			}
		}
	} else {
		toRet = append(toRet, h.Abilities...)
	}
	if len(response) == 0 {
		successfulQuery = true
	}

	if successfulQuery {
		log.Printf("returning %d abilities to %s", len(toRet), r.RemoteAddr)
		marshalledResponse, err := json.Marshal(toRet)
		if err != nil {
			response = []byte(fmt.Sprintf("an error occurred while getting the requested data -- %s", err))
		}
		response = marshalledResponse
	}
	_, writeErr := w.Write(response)
	if writeErr != nil {
		log.Printf("WARNING: failed to write response -- %s", writeErr)
	}
}
