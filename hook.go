package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexpop/data-proxy/types"
	"github.com/alexpop/data-proxy/utils"
	"github.com/julienschmidt/httprouter"
)

var apiData *ApiData
var logData *types.JsonProxyLog

func init() {
	configString64 := os.Getenv("CONFIG_YAML_CONTENT")
	if configString64 == "" {
		log.Fatal("Backend ENV variable CONFIG_YAML_CONTENT not set, aborting!")
	}
	configString, err := base64.StdEncoding.DecodeString(configString64)
	if err != nil {
		log.Fatal("Backend ENV variable CONFIG_YAML_CONTENT cannot be base64 decoded")
	}
	err, _, azureMaps := LoadAndValidateYamlConfig(configString)
	if err != nil {
		log.Fatalf("Error loading config from CONFIG_YAML_CONTENT: %s", err.Error())
		return
	}
	apiData = &ApiData{
		AzureMaps: azureMaps,
		Stats: &types.JsonStats{
			StartTime:     time.Now().UTC().Format(time.RFC3339),
			ResponseCodes: make(map[string]uint32, 0),
		},
	}
	fmt.Println(" * Initialized apiData!")
}

// HttpHook is used when deploying the package to PaaS, for example as a Google Cloud function
// The config file is provided as a base64 encoded value for the CONFIG_YAML_CONTENT ENV variable (`cat dp-config-proxy.yaml | base64 -w 0`)
func HttpHook(w http.ResponseWriter, r *http.Request) {
	logData = commonHTTP(w, r)
	router := httprouter.New()
	router.GET("/version", apiData.returnVersion)
	router.GET("/stats", apiData.returnStats)
	router.POST("/azure/workspace/:workspace/log/:log_name", apiData.postWorkspaceLog)
	targetHandler, params, _ := router.Lookup(r.Method, r.RequestURI)
	fmt.Printf("GOT: %s %+v %v\n", r.Method, r.RequestURI, params)
	if targetHandler != nil {
		targetHandler(w, r, params)
	} else {
		http.Error(w, utils.JsonErrorIt(logData, fmt.Sprintf("Unsupported Method (%s) & URI (%s) combination", r.Method, r.RequestURI), http.StatusNotFound), http.StatusNotFound)
		return
	}
}
