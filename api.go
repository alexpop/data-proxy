package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"./types"
	"./utils"
	"github.com/julienschmidt/httprouter"
)

var VERSION string = "0.0.0"
var BINARY_SHA256 = "unknown"

func serveRestEndpoints(port int) {
	router := httprouter.New()
	router.GET("/v1/version", returnVersion)
	router.POST("/v1/azure/workspace/:workspace/log/:log_name", createRun)

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

	log.Printf("   Listening on 0.0.0.0:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}

// GET /v1/version
func returnVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	commonHTTP(w, r)
	ver := types.VersionJson{
		Version: VERSION,
		Sha256:  BINARY_SHA256,
	}
	jsonOutPoint := utils.PrettyPrintJson(types.JsonResponse{Data: ver})
	w.Write(jsonOutPoint)
}

// POST /v1/azure/workspace/:id/log/:name
func createRun(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	commonHTTP(w, r)
	logName := param.ByName("log_name")
	workspace := param.ByName("workspace")

	jsonOutPoint := utils.PrettyPrintJson(types.JsonResponse{Data: fmt.Sprintf("Successfully posted data to workspace '%s', log '%s'", workspace, logName)})
	w.Write(jsonOutPoint)
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
	http.Error(w, utils.JsonErrorIt("Internal server error", ""), http.StatusNotFound)
}

func dieNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.Error(w, utils.JsonErrorIt("URI path not defined", ""), http.StatusNotFound)
}
