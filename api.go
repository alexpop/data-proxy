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

func serveRestEndpoints(port int, config *AzureConfig) {
	router := httprouter.New()
	router.GET("/version", returnVersion)
	router.POST("/azure/workspace/:workspace/log/:log_name", config.createRun)

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

	log.Printf("Listening on 0.0.0.0:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}

// GET /version
func returnVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	commonHTTP(w, r)
	ver := types.VersionJson{
		Version: VERSION,
		Sha256:  BINARY_SHA256,
	}
	jsonOutPoint := utils.PrettyPrintJson(types.JsonResponse{Data: ver})
	w.Write(jsonOutPoint)
}

// POST /azure/workspace/:id/log/:name
func (config AzureConfig) createRun(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	commonHTTP(w, r)
	logName := param.ByName("log_name")
	workspace := param.ByName("workspace")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, utils.JsonErrorIt(fmt.Sprintf("Body read error: %s", err.Error())), http.StatusBadRequest)
		return
	}
	// 01234567-3ca5-4b65-8383-c12a5cda28b3
	if regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$").MatchString(workspace) {
		if conf, ok := config.WksIdMap[workspace]; ok {
			// workspace has the pattern of a UUID defined in the config file
			err = azure.PostData(workspace, logName, conf.Secret, string(bodyBytes))
		} else if conf, ok = config.WksNameMap[workspace]; ok {
			// workspace has the pattern of a UUID, but is actually a name, uncommon but possible
			err = azure.PostData(conf.Id, logName, conf.Secret, string(bodyBytes))
		} else {
			// workspace not a name either, 404-ing
			http.Error(w, utils.JsonErrorIt(fmt.Sprintf("Workspace %s not found in the proxy config.", err.Error())), http.StatusNotFound)
			return
		}
	} else if conf, ok := config.WksNameMap[workspace]; ok {
		// workspace doesn't have the pattern of a UUID, must be a name to continue
		err = azure.PostData(conf.Id, logName, conf.Secret, string(bodyBytes))
	} else {
		// workspace not a name either, 404-ing
		http.Error(w, utils.JsonErrorIt(fmt.Sprintf("Workspace %s not found in the proxy config.", err.Error())), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, utils.JsonErrorIt(fmt.Sprintf("Azure API error: %s", err.Error())), http.StatusInternalServerError)
		return
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
