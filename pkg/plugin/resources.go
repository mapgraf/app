package plugin

import (
	"encoding/json"
	"fmt"
	"mapgl-app/pkg/database"
	"mapgl-app/pkg/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	_ "modernc.org/sqlite"
)

//var (
//	dbMu sync.RWMutex
//	DB   *redka.DB
//)

func init() {
	// Initialize the DB instance
	//DB = database.DB
}

//func GetDB() *redka.DB {
//	dbMu.RLock()
//	defer dbMu.RUnlock()
//	return DB
//}

type SetItem struct {
	id    int
	Elem  []byte
	Score float64
}

type MySetItem struct {
	id    int
	Elem  string
	Score float64
}

type RedisIdItem struct {
	ID    string  `json:"tsId"`
	Name  string  `json:"name"`
	Score float64 `json:"updatedAt"`
}

type NewIdDoc struct {
	Name      string `json:"name"`
	TsId      string `json:"tsId"`
	UpdatedAt int64  `json:"updatedAt"`
	Deleted   bool   `json:"_deleted"`
}

type NewEdgeDoc struct {
	Id          string        `json:"id"`
	ParPath     []interface{} `json:"parPath"`
	UpdatedAt   float64       `json:"updatedAt"`
	IsEphemeral *bool         `json:"isEph,omitempty"`
	Deleted     bool          `json:"_deleted"`
}

type Response struct {
	Status    string `json:"status"`
	OrgName   string `json:"org"`
	Host      string `json:"host"`
	IsPower   bool   `json:"isPower"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

// ConvertSetItems converts []SetItem to []MySetItem
func ConvertSetItems(items []SetItem) []MySetItem {
	var convertedItems []MySetItem
	for _, item := range items {
		// Convert []byte to string
		elemString := string(item.Elem)

		// Append to convertedItems slice
		convertedItems = append(convertedItems, MySetItem{
			id:    item.id,
			Elem:  elemString,
			Score: item.Score,
		})
	}
	return convertedItems
}

func (a *App) pullIds(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		MinTimestamp int64  `json:"minTimestamp"`
		FileName     string `json:"fileName"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert MinTimestamp from int64 to float64
	minTimestampFloat := float64(body.MinTimestamp)
	fileName := body.FileName

	//   maxScore := "+inf"

	// Retrieve members of the sorted set within the specified score range

	//DB := GetDB()
	DB, err0 := database.GetDB("./public/seed/" + fileName + ".db")
	if err0 != nil {
		log.DefaultLogger.Error("Failed to get database connection:", "filename", fileName, "error", err0)
		http.Error(w, "failed to get database connection", http.StatusInternalServerError)
		return
	}

	maxEl, err1 := DB.ZSet().RangeWith("lastIds").ByRank(0, 0).Desc().Run()
	if err1 != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Failed !!!!!!!!!!!!!! lastIds fial!! : %v", err1))
	}

	var maxScore float64
	if len(maxEl) > 0 {
		maxScore = maxEl[0].Score
	}

	log.DefaultLogger.Info(fmt.Sprintf("minTimestampFloat: %+v", minTimestampFloat))
	log.DefaultLogger.Info(fmt.Sprintf("MaxScore: %+v", maxScore))

	setItems, err2 := DB.ZSet().RangeWith("lastIds").ByScore(minTimestampFloat, maxScore).Run()
	if err2 != nil {
		panic(err2)
	}

	var customSetItems []SetItem
	for _, item := range setItems {
		customSetItems = append(customSetItems, SetItem{
			Elem:  item.Elem,
			Score: item.Score,
		})
	}

	// Convert custom SetItems to MySetItems
	mySetItems := ConvertSetItems(customSetItems)

	var redisItems []RedisIdItem

	for _, item := range mySetItems {
		// Retrieve the string value from Redis using the key from Elem property
		value, err := DB.Str().Get(item.Elem)
		if err != nil {
			// Handle error
			log.DefaultLogger.Error(fmt.Sprintf("Failed to retrieve value for key %s: %v", item.Elem, err))
			continue // Continue to the next item
		}

		// Extract the value from the result tuple
		stringValue := string(value)

		// Create a new RedisIdItem struct and append it to the redisItems slice
		redisItems = append(redisItems, RedisIdItem{
			ID:    item.Elem, // Assuming item.Elem is already a string
			Name:  stringValue,
			Score: item.Score,
		})
	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(redisItems); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (a *App) pullEdges(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		MinTimestamp int64  `json:"minTimestamp"`
		FileName     string `json:"fileName"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert MinTimestamp from int64 to float64
	minTimestampFloat := float64(body.MinTimestamp)

	//   maxScore := "+inf"

	// Retrieve members of the sorted set within the specified score range

	//DB := GetDB()
	fileName := body.FileName
	DB, err0 := database.GetDB("./public/seed/" + fileName + ".db")
	if err0 != nil {
		log.DefaultLogger.Error("Failed to get database connection:", "filename", fileName, "error", err0)
		http.Error(w, "failed to get database connection", http.StatusInternalServerError)
		return
	}

	maxEl, err := DB.ZSet().RangeWith("lastEdges").ByRank(0, 0).Desc().Run()
	var maxScore float64
	if len(maxEl) > 0 {
		maxScore = maxEl[0].Score
	}

	setItems, err := DB.ZSet().RangeWith("lastEdges").ByScore(minTimestampFloat, maxScore).Run()
	if err != nil {
		panic(err)
	}

	var customSetItems []SetItem
	for _, item := range setItems {
		customSetItems = append(customSetItems, SetItem{
			Elem:  item.Elem,
			Score: item.Score,
		})
	}

	// Convert custom SetItems to MySetItems
	mySetItems := ConvertSetItems(customSetItems)

	var redisItems []map[string]interface{}

	for _, item := range mySetItems {
		// Retrieve the map[string]core.Value value from Redis using the key from Elem property
		valueMap, err := DB.Hash().Items(item.Elem)
		if err != nil {
			// Handle error
			log.DefaultLogger.Error(fmt.Sprintf("Failed to retrieve value for key %s: %v", item.Elem, err))
			continue // Continue to the next item
		}

		// Extract individual values from the valueMap and unmarshal them
		parPathJSON := valueMap["parPath"].String()

		deletedStr := valueMap["deleted"].String()
		deletedValue, err := strconv.ParseBool(deletedStr)
		var parPathValue []interface{} // Assuming parPathValue is an array of any type

		// Unmarshal the JSON strings into appropriate Go data types
		if err := json.Unmarshal([]byte(parPathJSON), &parPathValue); err != nil {
			// Handle error
		}

		var isEphValue *bool // Change to pointer type to make it optional

		isEphStr, isEphExists := valueMap["isEph"]
		if isEphExists {
			isEphVal, err := strconv.ParseBool(isEphStr.String())
			if err != nil {
				// Handle error
			}
			isEphValue = &isEphVal
		}

		// Construct RedisEdgeItem instance
		redisItem := NewEdgeDoc{
			Id:          item.Elem,
			ParPath:     parPathValue,
			UpdatedAt:   item.Score,
			Deleted:     deletedValue,
			IsEphemeral: isEphValue, // Assign the pointer
		}

		// Convert RedisEdgeItem to map[string]interface{}
		redisItemMap := map[string]interface{}{
			"id":        redisItem.Id,
			"parPath":   redisItem.ParPath,
			"updatedAt": redisItem.UpdatedAt,
			"_deleted":  redisItem.Deleted,
		}

		// Add IsEphemeral to map only if it's not nil
		if redisItem.IsEphemeral != nil {
			redisItemMap["isEph"] = *redisItem.IsEphemeral
		}

		// Append the map to redisItems
		redisItems = append(redisItems, redisItemMap)
	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(redisItems); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (a *App) pushIds(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		NewDocs  []NewIdDoc `json:"newDocs"`
		FileName string     `json:"fileName"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	NewDocs := body.NewDocs

	for _, item := range NewDocs {
		// Retrieve the string value from Redis using the key from Elem property

		//DB := GetDB()
		fileName := body.FileName
		DB, err0 := database.GetDB("./public/seed/" + fileName + ".db")
		if err0 != nil {
			log.DefaultLogger.Error("Failed to get database connection:", "filename", fileName, "error", err0)
			http.Error(w, "failed to get database connection", http.StatusInternalServerError)
			return
		}

		err := DB.Str().Set(item.TsId, item.Name)
		if err != nil {
			// Handle error
			log.DefaultLogger.Error(fmt.Sprintf("Ids: Failed to set value for key %s: %v", item.TsId, err))
			continue // Continue to the next item
		}
		_, err2 := DB.ZSet().AddMany("lastIds", map[any]float64{
			item.TsId: float64(item.UpdatedAt),
		})

		if err2 != nil {
			// Handle error
			log.DefaultLogger.Error(fmt.Sprintf("Ids: Failed to zadd value for key %s: %v", item.TsId, err))
			continue // Continue to the next item
		}
	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(NewDocs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (a *App) pushEdges(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		NewDocs  []NewEdgeDoc `json:"newDocs"`
		FileName string       `json:"fileName"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	NewDocs := body.NewDocs

	for _, item := range NewDocs {
		// Retrieve the string value from Redis using the key from Elem property

		//updatedAtJSON := strconv.FormatInt(item.UpdatedAt, 10)
		parPathJSON, _ := json.Marshal(item.ParPath)

		// Construct the map for setting hash values
		hashValues := map[string]interface{}{
			"deleted": item.Deleted,
			"parPath": string(parPathJSON),
		}

		// Add isEph to the hashValues only if it's not nil
		if item.IsEphemeral != nil {
			hashValues["isEph"] = *item.IsEphemeral
		}

		//DB := GetDB()
		fileName := body.FileName
		DB, err0 := database.GetDB("./public/seed/" + fileName + ".db")
		if err0 != nil {
			log.DefaultLogger.Error("Failed to get database connection:", "filename", fileName, "error", err0)
			http.Error(w, "failed to get database connection", http.StatusInternalServerError)
			return
		}

		_, err := DB.Hash().SetMany(item.Id, hashValues)

		if err != nil {
			// Handle error
			log.DefaultLogger.Error(fmt.Sprintf("Edges: Failed to zadd value for key %s: %v", item.Id, item.ParPath, err))
			continue // Continue to the next item
		}

		_, err2 := DB.ZSet().AddMany("lastEdges", map[any]float64{
			item.Id: float64(item.UpdatedAt),
		})

		if err2 != nil {
			// Handle error
			log.DefaultLogger.Error(fmt.Sprintf("Edges: Failed to zset value for key %s: %v", item.Id, item.ParPath, err))
			continue // Continue to the next item
		}

	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(NewDocs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (a *App) handlePing(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	tokenString := a.MapglSettings.ApiToken
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Host string `json:"host"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	publicKey := JWT_PUBLIC_KEY
	host := body.Host

	claims, err := util.DecodeToken(tokenString, publicKey)
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Invalid license token: %v %v", err, tokenString))
		return
	}

	response := createResponseFromClaims(host, claims, tokenString)
	if response.Status == "token expired: "+tokenString {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	writeJSONResponse(w, response)
}

func writeErrorResponse(w http.ResponseWriter, errorMsg string) {

	response := Response{
		Status:    errorMsg,
		OrgName:   "invalid token",
		Host:      "invalid token",
		ExpiresAt: time.Now().Unix(),
		IsPower:   false,
	}
	writeJSONResponse(w, response)
}

func createResponseFromClaims(host string, claims *util.Claims, tokenString string) Response {
	response := Response{
		OrgName:   claims.OrgName,
		Host:      host,
		Token:     tokenString,
		ExpiresAt: claims.ExpiresAt.Time.Unix(), // global token expire
		IsPower:   false,
	}

	hostPerm, err := util.CheckHost(host, claims)

	if err == nil {
		// valid
		response.Status = "token valid"

	} else {
		response.Status = fmt.Sprintf(
			"token invalid: %s (exp: %v, err: %v)",
			claims.ExpiresAt.Time.Unix(),
			err,
		)
	}

	if hostPerm.Domain != "" { // we found a host entry
		// hostPerm is valid; override expiration
		exp := hostPerm.ExpiresAt            //.Time.Unix()
		response.ExpiresAt = exp.Time.Unix() // always overwrite
		if exp.Time.Before(time.Now()) {
			response.Status += ". Host expired: " + fmt.Sprintf("%d", response.ExpiresAt)
		} else {
			response.IsPower = true
		}
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		response.ExpiresAt = claims.ExpiresAt.Time.Unix()
		response.Status += ". Global token expired: " + fmt.Sprintf("%d", claims.ExpiresAt.Time.Unix())
	}

	return response
}

func writeJSONResponse(w http.ResponseWriter, response Response) {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(responseJSON); err != nil {
		http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (a *App) handleEcho(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a *App) registerRoutes(r *mux.Router) {

	r.HandleFunc("/ping", a.handlePing)
	r.HandleFunc("/echo", a.handleEcho)

	r.HandleFunc("/pullIds", a.pullIds)
	r.HandleFunc("/pullEdges", a.pullEdges)

	publicKey := JWT_PUBLIC_KEY
	if ok, _ := util.HasSomePowerHost(a.MapglSettings.ApiToken, publicKey); ok {
		r.HandleFunc("/pushIds", a.pushIds)
		r.HandleFunc("/pushEdges", a.pushEdges)
	} else {
		log.DefaultLogger.Info("Not a power host")
	}

	//log.DefaultLogger.Info(fmt.Sprintf("Test. Gen token expires at: %s", time.Unix(expirationTime, 0)))

}
