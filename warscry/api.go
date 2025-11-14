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
	DataStore *DataStore
}

type AbilityHandler struct {
	DataStore *DataStore
}

type RootHandler struct {
	Version   string
	DataStore *DataStore
	DocsURL   string
}

type APIInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Endpoints    []string `json:"endpoints"`
	FighterCount int      `json:"fighter_count"`
	AbilityCount int      `json:"ability_count"`
	DocsURL      string   `json:"documentation_url"`
}

type HealthHandler struct {
	DataStore *DataStore
}

type HealthResponse struct {
	Status          string `json:"status"`
	FightersLoaded  int    `json:"fighters_loaded"`
	AbilitiesLoaded int    `json:"abilities_loaded"`
}

func (R *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")

	// Determine content type based on Accept header
	wantsJSON := strings.Contains(accept, "application/json")
	wantsHTML := strings.Contains(accept, "text/html")

	// Get current counts from DataStore
	fighterCount, abilityCount := R.DataStore.GetCounts()

	apiInfo := APIInfo{
		Name:         "Warcry API",
		Version:      R.Version,
		Endpoints:    []string{"/", "/fighters", "/abilities", "/health"},
		FighterCount: fighterCount,
		AbilityCount: abilityCount,
		DocsURL:      R.DocsURL,
	}

	// JSON response for API clients
	if wantsJSON && !wantsHTML {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(apiInfo); err != nil {
			log.Printf("WARNING: failed to encode JSON response -- %s", err)
		}
		return
	}

	// HTML response for browsers
	if wantsHTML {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Warcry API</title>
    <style>
        body { font-family: sans-serif; max-width: 800px; margin: 40px auto; padding: 0 20px; line-height: 1.6; }
        h1 { color: #333; }
        code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
        pre { background: #f4f4f4; padding: 10px; border-radius: 5px; overflow-x: auto; }
        .endpoint { margin: 20px 0; }
        .stats { background: #e8f5e9; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <h1>Welcome to Warcry API %s</h1>
    <div class="stats">
        <strong>Data loaded:</strong> %d fighters, %d abilities
    </div>
    <h2>Endpoints</h2>
    <div class="endpoint">
        <h3>GET /fighters</h3>
        <p>Query fighters by characteristics. Supports operators for numeric fields.</p>
        <p><strong>Examples:</strong></p>
        <pre>GET /fighters?attacks__gte=4
GET /fighters?wounds__gt=20&toughness__gte=5
GET /fighters?warband=stormcast-eternals&runemarks=hero</pre>
        <p><strong>Operators:</strong> <code>__gt</code> (greater than), <code>__gte</code> (greater or equal), <code>__lt</code> (less than), <code>__lte</code> (less or equal)</p>
    </div>
    <div class="endpoint">
        <h3>GET /abilities</h3>
        <p>Query abilities by characteristics.</p>
        <p><strong>Examples:</strong></p>
        <pre>GET /abilities?warband=stormcast-eternals
GET /abilities?description=wounds</pre>
    </div>
    <div class="endpoint">
        <h3>GET /health</h3>
        <p>Health check endpoint. Returns API status.</p>
    </div>
    <h2>Documentation</h2>
    <p>OpenAPI specification: <a href="%s">%s</a></p>
</body>
</html>`, R.Version, fighterCount, abilityCount, R.DocsURL, R.DocsURL)
		if _, err := w.Write([]byte(html)); err != nil {
			log.Printf("WARNING: failed to write HTML response -- %s", err)
		}
		return
	}

	// Plain text fallback
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	plainText := fmt.Sprintf(`Welcome to Warcry API %s

Data loaded: %d fighters, %d abilities

Endpoints:
- GET /fighters - Query fighters by characteristics
- GET /abilities - Query abilities
- GET /health - Health check

Fighter characteristics can be queried using ?characteristic=value
Example: /fighters?attacks=4

Append operators (__gt, __gte, __lt, __lte) for comparisons
Example: /fighters?attacks__gte=4

For abilities, use description=word to search descriptions
Example: /abilities?description=wounds

Documentation: %s
`, R.Version, fighterCount, abilityCount, R.DocsURL)
	if _, err := w.Write([]byte(plainText)); err != nil {
		log.Printf("WARNING: failed to write plain text response -- %s", err)
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	// Get current counts from DataStore
	fighterCount, abilityCount := h.DataStore.GetCounts()

	response := HealthResponse{
		Status:          "ok",
		FightersLoaded:  fighterCount,
		AbilitiesLoaded: abilityCount,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("WARNING: failed to encode health response -- %s", err)
	}
}

// ErrorResponse represents a structured API error
type ErrorResponse struct {
	Error string `json:"error"`
}

// QueryError represents a query parameter validation error
type QueryError struct {
	Parameter string
	Value     string
	Reason    string
}

func (e QueryError) Error() string {
	return fmt.Sprintf("invalid query parameter '%s=%s': %s", e.Parameter, e.Value, e.Reason)
}

// writeErrorJSON writes a JSON error response with the specified HTTP status code
func writeErrorJSON(w http.ResponseWriter, statusCode int, errMsg string) {
	SetHeaderDefaults(&w)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(ErrorResponse{Error: errMsg}); err != nil {
		log.Printf("WARNING: failed to encode error response: %v", err)
	}
}

// validFighterParams lists all recognized query parameters for /fighters endpoint
var validFighterParams = map[string]bool{
	// String params
	"name": true, "_id": true, "warband": true, "subfaction": true,
	"grand_alliance": true, "runemarks": true, "weapon_runemark": true,
	// Integer params (base names only, operators checked separately)
	"movement": true, "toughness": true, "wounds": true, "points": true,
	"attacks": true, "strength": true, "dmg_hit": true, "dmg_crit": true,
	"min_range": true, "max_range": true,
}

// validAbilityParams lists all recognized query parameters for /abilities endpoint
var validAbilityParams = map[string]bool{
	"_id": true, "name": true, "warband": true, "cost": true,
	"description": true, "runemarks": true,
}

// validOperators lists all supported operator suffixes
var validOperators = []string{"__gt", "__gte", "__lt", "__lte"}

// validateQueryParams checks if all query parameters are recognized
// Returns error listing all unrecognized parameters
func validateQueryParams(form map[string][]string, validParams map[string]bool, allowOperators bool) error {
	var invalidParams []string

	for param := range form {
		valid := validParams[param]

		// If operators allowed, check if param is base + operator
		if !valid && allowOperators {
			for _, op := range validOperators {
				if strings.HasSuffix(param, op) {
					baseName := strings.TrimSuffix(param, op)
					if validParams[baseName] {
						valid = true
						break
					}
				}
			}
		}

		if !valid {
			invalidParams = append(invalidParams, param)
		}
	}

	if len(invalidParams) > 0 {
		return fmt.Errorf("unrecognized query parameters: %s", strings.Join(invalidParams, ", "))
	}
	return nil
}

// validateIntParam validates a single integer parameter value
func validateIntParam(name string, values []string) error {
	if len(values) == 0 {
		return nil
	}

	// Check each value (allowing multiple for OR logic)
	for _, val := range values {
		_, err := strconv.Atoi(val)
		if err != nil {
			return QueryError{
				Parameter: name,
				Value:     val,
				Reason:    "must be an integer",
			}
		}
	}
	return nil
}

// validateIntParams validates all integer parameters in the request
func validateIntParams(form map[string][]string, paramNames []string) error {
	for _, param := range paramNames {
		// Check base param and all operator variants
		if form[param] != nil {
			if err := validateIntParam(param, form[param]); err != nil {
				return err
			}
		}
		for _, op := range validOperators {
			fullKey := param + op
			if form[fullKey] != nil {
				if err := validateIntParam(fullKey, form[fullKey]); err != nil {
					return err
				}
			}
		}
	}
	return nil
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

func GetOperator(queryKey string) (Operator, error) {
	OperatorMap := map[string]Operator{
		"gt":  GreaterThan,
		"gte": GreaterThanOrEqualTo,
		"lt":  LessThan,
		"lte": LessThanOrEqualTo,
	}
	if strings.Contains(queryKey, "__") {
		opString := strings.Split(queryKey, "__")
		opKey := opString[len(opString)-1]
		op, exists := OperatorMap[opKey]
		if !exists {
			return nil, fmt.Errorf("invalid operator: %s", opKey)
		}
		return op, nil
	}
	return Equals, nil
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

func SetHeaderDefaults(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")

}

func (h *FighterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get current fighters from DataStore (snapshot at request start)
	fighters := h.DataStore.GetFighters()

	var (
		response []byte
		toRet    Fighters
		wg       sync.WaitGroup
	)

	// Step 1: Parse form data
	if err := r.ParseForm(); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, fmt.Sprintf("failed to parse request: %v", err))
		log.Printf("Bad request from %s: %v", r.RemoteAddr, err)
		return
	}

	// Step 2: Validate all query parameters are recognized
	if err := validateQueryParams(r.Form, validFighterParams, true); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		log.Printf("Bad request from %s: %v", r.RemoteAddr, err)
		return
	}

	// Step 3: Validate all integer parameters have valid values
	intParams := []string{"movement", "toughness", "wounds", "points",
		"attacks", "strength", "dmg_hit", "dmg_crit",
		"min_range", "max_range"}
	if err := validateIntParams(r.Form, intParams); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		log.Printf("Bad request from %s: %v", r.RemoteAddr, err)
		return
	}

	// All validation passed - proceed with filtering
	fChan := make(chan Fighter, len(fighters))

	if len(r.Form) > 0 {
		// Filter fighters based on validated criteria
		for i := range fighters {
			wg.Add(1)
			index := i
			go func() {
				defer wg.Done()
				fighters[index].MatchesRequest(r, fChan)
			}()
		}
		wg.Wait()
	} else {
		// No criteria - return all fighters
		toRet = append(toRet, fighters...)
	}

	close(fChan)
	for includedFighter := range fChan {
		toRet = append(toRet, includedFighter)
	}

	// Marshal and return results
	log.Printf("returning %d fighters to %s", len(toRet), r.RemoteAddr)
	marshalledResponse, err := json.Marshal(toRet)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("error marshalling response: %v", err))
		log.Printf("ERROR: failed to marshal fighters: %s", err)
		return
	}

	response = marshalledResponse
	SetHeaderDefaults(&w)
	_, writeErr := w.Write(response)
	if writeErr != nil {
		log.Printf("WARNING: failed to write response -- %s", writeErr)
	}
}

func (h *AbilityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get current abilities from DataStore (snapshot at request start)
	abilities := h.DataStore.GetAbilities()

	var response []byte

	// Step 1: Parse form data
	if err := r.ParseForm(); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, fmt.Sprintf("failed to parse request: %v", err))
		log.Printf("Bad request from %s: %v", r.RemoteAddr, err)
		return
	}

	// Step 2: Validate all query parameters are recognized
	if err := validateQueryParams(r.Form, validAbilityParams, false); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		log.Printf("Bad request from %s: %v", r.RemoteAddr, err)
		return
	}

	// All validation passed - proceed with filtering
	var toRet Abilities

	if len(r.Form) > 0 {
		for _, a := range abilities {
			include, err := a.MatchesRequest(r)
			if err != nil {
				// This should not happen with validated input
				writeErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("error filtering abilities: %v", err))
				log.Printf("ERROR: unexpected error from %s: %v", r.RemoteAddr, err)
				return
			}
			if include {
				toRet = append(toRet, a)
			}
		}
	} else {
		toRet = append(toRet, abilities...)
	}

	// Marshal and return results
	log.Printf("returning %d abilities to %s", len(toRet), r.RemoteAddr)
	marshalledResponse, err := json.Marshal(toRet)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("error marshalling response: %v", err))
		log.Printf("ERROR: failed to marshal abilities: %s", err)
		return
	}

	response = marshalledResponse
	SetHeaderDefaults(&w)
	_, writeErr := w.Write(response)
	if writeErr != nil {
		log.Printf("WARNING: failed to write response -- %s", writeErr)
	}
}
