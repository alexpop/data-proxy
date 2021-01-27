package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"runtime/debug"

	"github.com/alexpop/data-proxy/azure"
	"github.com/alexpop/data-proxy/jlog"
	"github.com/alexpop/data-proxy/types"
	"github.com/alexpop/data-proxy/utils"
	"github.com/julienschmidt/httprouter"
)

var VERSION string = "0.0.0"
var BINARY_SHA256 = "unknown"

type ApiData struct {
	AzureMaps *types.AzureMaps
	Stats     *types.JsonStats
}

func ServeRestEndpoints(hostPort string, apiData *ApiData) {
	router := httprouter.New()
	router.GET("/version", apiData.returnVersion)
	router.GET("/stats", apiData.returnStats)
	router.POST("/azure/workspace/:workspace/log/:log_name", apiData.postWorkspaceLog)

	router.PanicHandler = apiData.dieHard
	// router.NotFound = ... later if we want to return a different message for not found
	// router.MethodNotAllowed = ... later if we want to return a different message for method
	router.NotFound = http.HandlerFunc(apiData.dieNotFound)

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

	jlog.Info(fmt.Sprintf("Listening on %s", hostPort))
	err := http.ListenAndServe(hostPort, router)
	if err != nil {
		jlog.Fatal(err.Error())
	}
}

// GET /version
func (apiData *ApiData) returnVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logData := commonHTTP(w, r)
	ver := types.JsonVersion{
		Version: VERSION,
		Sha256:  BINARY_SHA256,
	}
	jlog.Proxy(logData)
	jsonOutPoint := utils.PrettyPrintJson(types.JsonResponse{Data: ver})
	apiData.updateStats(http.StatusOK)
	fmt.Fprintf(w, jsonOutPoint)
}

// GET /stats
func (apiData *ApiData) returnStats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logData := commonHTTP(w, r)
	jlog.Proxy(logData)
	jsonOutPoint := utils.PrettyPrintJson(types.JsonResponse{Data: apiData.Stats})
	apiData.updateStats(http.StatusOK)
	fmt.Fprintf(w, jsonOutPoint)
}

// POST /azure/workspace/:id/log/:name
func (apiData *ApiData) postWorkspaceLog(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	logData := commonHTTP(w, r)
	logName := param.ByName("log_name")
	workspace := param.ByName("workspace")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiData.updateStats(http.StatusBadRequest)
		http.Error(w, utils.JsonErrorIt(logData, fmt.Sprintf("Body read error: %s", err.Error()), http.StatusNotFound), http.StatusBadRequest)
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
			apiData.updateStats(http.StatusNotFound)
			// workspace not a name either, 404-ing
			http.Error(w, utils.JsonErrorIt(logData, fmt.Sprintf("Workspace %s not found in the proxy config", err.Error()), http.StatusNotFound), http.StatusNotFound)
			return
		}
	} else if conf, ok := apiData.AzureMaps.WksNameMap[workspace]; ok {
		// workspace doesn't have the pattern of a UUID, must be a name to continue
		err, statusCode = azure.PostData(conf.Id, logName, conf.Secret, string(bodyBytes))
	} else {
		apiData.updateStats(http.StatusNotFound)
		// workspace not a name either, 404-ing
		http.Error(w, utils.JsonErrorIt(logData, fmt.Sprintf("Workspace %s not found in the proxy config", workspace), http.StatusNotFound), http.StatusNotFound)
		return
	}
	statusCodeString := ""
	if statusCode > 0 {
		statusCodeString = fmt.Sprintf("%d ", statusCode)
	}

	if err != nil {
		apiData.updateStats(statusCode)
		http.Error(w, utils.JsonErrorIt(logData, fmt.Sprintf("Azure %sAPI error: %s", statusCodeString, err.Error()), statusCode), statusCode)
		return
	}
	logData.Status = uint16(statusCode)
	jlog.Proxy(logData)
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

func getJsonProxyLog(r *http.Request) *types.JsonProxyLog {
	ip, _ := utils.RightSplit(r.RemoteAddr, ":")
	return &types.JsonProxyLog{
		Method: r.Method,
		URI:    r.URL.String(),
		IP:     ip,
	}
}

func commonHTTP(w http.ResponseWriter, r *http.Request) *types.JsonProxyLog {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return getJsonProxyLog(r)
}

func (apiData *ApiData) dieHard(w http.ResponseWriter, r *http.Request, err interface{}) {
	jlog.Error(r.URL.Path + " " + string(debug.Stack()))
	apiData.updateStats(http.StatusInternalServerError)
	http.Error(w, utils.JsonErrorIt(getJsonProxyLog(r), "Internal server error", http.StatusInternalServerError), http.StatusInternalServerError)
}

func (apiData *ApiData) dieNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	apiData.updateStats(http.StatusNotFound)
	http.Error(w, utils.JsonErrorIt(getJsonProxyLog(r), "URI path not defined", http.StatusNotFound), http.StatusNotFound)
}
