package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"runtime/debug"

	"./azure"

	"./types"
	"./utils"
	"github.com/julienschmidt/httprouter"
)

var VERSION string = "0.0.0"
var BINARY_SHA256 = "unknown"

type ApiData struct {
	AzureMaps *types.AzureMaps
	Stats     *types.JsonStats
}

func serveRestEndpoints(hostPort string, apiData *ApiData) {
	router := httprouter.New()
	router.GET("/version", returnVersion)
	router.GET("/stats", apiData.returnStats)
	router.POST("/azure/workspace/:workspace/log/:log_name", apiData.postWorkspaceLog)

	router.PanicHandler = dieHard
	// router.NotFound = ... later if we want to return a different message for not found
	// router.MethodNotAllowed = ... later if we want to return a different message for method
	router.NotFound = http.HandlerFunc(dieNotFound)

	// Reply to browser OPTIONS calls (e.g. CORS)
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			// Set CORS headers
			header := w.Header()
			header.Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers")
			header.Set("Access-Control-Allow-Methods", header.Get("Allow"))
			header.Set("Access-Control-Allow-Origin", "*")
		}

		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})

	apiData.Stats.ResponseCodes = make(map[string]uint32, 0)

	log.Printf("Listening on %s", hostPort)
	log.Fatal(http.ListenAndServe(hostPort, router))
}

// GET /version
func returnVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	commonHTTP(w, r)
	ver := types.JsonVersion{
		Version: VERSION,
		Sha256:  BINARY_SHA256,
	}
	jsonOutPoint := utils.PrettyPrintJson(types.JsonResponse{Data: ver})
	w.Write(jsonOutPoint)
}

// GET /stats
func (apiData *ApiData) returnStats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	commonHTTP(w, r)
	jsonOutPoint := utils.PrettyPrintJson(types.JsonResponse{Data: apiData.Stats})
	w.Write(jsonOutPoint)
}

// POST /azure/workspace/:id/log/:name
func (apiData *ApiData) postWorkspaceLog(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	commonHTTP(w, r)
	logName := param.ByName("log_name")
	workspace := param.ByName("workspace")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, utils.JsonErrorIt(fmt.Sprintf("Body read error: %s", err.Error())), http.StatusBadRequest)
		apiData.updateStats(http.StatusBadRequest)
		return
	}
	statusCode := 0
	if regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$").MatchString(workspace) {
		if conf, ok := apiData.AzureMaps.WksIdMap[workspace]; ok {
			// workspace has the pattern of a UUID defined in the config file
			err, statusCode = azure.PostData(workspace, logName, conf.Secret, string(bodyBytes))
		} else if conf, ok = apiData.AzureMaps.WksNameMap[workspace]; ok {
			// workspace has the pattern of a UUID, but is actually a name, uncommon but possible
			err, statusCode = azure.PostData(conf.Id, logName, conf.Secret, string(bodyBytes))
		} else {
			// workspace not a name either, 404-ing
			http.Error(w, utils.JsonErrorIt(fmt.Sprintf("Workspace %s not found in the proxy config", err.Error())), http.StatusNotFound)
			apiData.updateStats(http.StatusNotFound)
			return
		}
	} else if conf, ok := apiData.AzureMaps.WksNameMap[workspace]; ok {
		// workspace doesn't have the pattern of a UUID, must be a name to continue
		err, statusCode = azure.PostData(conf.Id, logName, conf.Secret, string(bodyBytes))
	} else {
		// workspace not a name either, 404-ing
		http.Error(w, utils.JsonErrorIt(fmt.Sprintf("Workspace %s not found in the proxy config", workspace)), http.StatusNotFound)
		apiData.updateStats(http.StatusNotFound)
		return
	}
	statusCodeString := ""
	if statusCode > 0 {
		statusCodeString = fmt.Sprintf("%d ", statusCode)
	}

	if err != nil {
		http.Error(w, utils.JsonErrorIt(fmt.Sprintf("Azure %sAPI error: %s", statusCodeString, err.Error())), statusCode)
		apiData.updateStats(http.StatusInternalServerError)
		return
	}
	apiData.updateStats(statusCode)
}

func (apiData *ApiData) updateStats(code int) {
	codeStr := fmt.Sprintf("%d", code)
	if _, ok := apiData.Stats.ResponseCodes[codeStr]; ok {
		apiData.Stats.ResponseCodes[codeStr] += 1
	} else {
		apiData.Stats.ResponseCodes[codeStr] = 1
	}
}

func commonHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ip, _ := utils.RightSplit(r.RemoteAddr, ":")
	log.Printf(" > %s %s from %s", r.Method, r.URL, ip)
}

func dieHard(w http.ResponseWriter, r *http.Request, err interface{}) {
	log.Println(r.URL.Path, string(debug.Stack())) // Collecting panic trace
	//debug.PrintStack()                             // or we can use PrintStack
	http.Error(w, utils.JsonErrorIt("Internal server error"), http.StatusNotFound)
}

func dieNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.Error(w, utils.JsonErrorIt("URI path not defined"), http.StatusNotFound)
}
