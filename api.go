package warcry_go

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

	fChan := make(chan Fighter, len(h.Fighters))

	if len(r.Form) > 0 {
		for i := range h.Fighters {
			wg.Add(1)
			index := i
			go func() {
				defer wg.Done()
				h.Fighters[index].MatchesRequest(r, fChan)
			}()
		}
		wg.Wait()
		close(fChan)
	} else {
		toRet = append(toRet, h.Fighters...)
	}

	// close(fChan)
	for includedFighter := range fChan {
		toRet = append(toRet, includedFighter)
	}
	if len(response) == 0 {
		successfulQuery = true
	}

	if successfulQuery {
		log.Printf("returning %d fighters to %s", len(toRet), r.RemoteAddr)
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
