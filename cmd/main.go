package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	api "github.com/alexpop/data-proxy"
	"github.com/alexpop/data-proxy/jlog"
	"github.com/alexpop/data-proxy/types"
	"github.com/alexpop/data-proxy/utils"
)

func main() {
	if len(os.Args) < 2 {
		jlog.Fatal(fmt.Sprintf("ERROR: Missing required argument to the yaml config file, example usage: %s config.yaml", os.Args[0]))
	}

	// TODO: Add options for:
	// -v --version  CLI
	// --no-version  API

	ex, err := os.Executable()
	if err != nil {
		jlog.Fatal(err.Error())
	}
	api.BINARY_SHA256, err = utils.GetFileSha256(ex)
	if err != nil {
		jlog.Fatal(err.Error())
	}
	configBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		jlog.Fatal(fmt.Sprintf("ERROR: Reading the config file %s returned error: %s", os.Args[1], err.Error()))
	}
	err, yamlConfig, azureMaps := api.LoadAndValidateYamlConfig(configBytes)
	if err != nil {
		jlog.Fatal(fmt.Sprintf("Error loading config file %s. %s", os.Args[1], err.Error()))
	}
	listenIP := "0.0.0.0"
	if yamlConfig.ListenIP != "" {
		listenIP = yamlConfig.ListenIP
	}
	listenPort := uint16(4000)
	if yamlConfig.ListenPort != 0 {
		listenPort = yamlConfig.ListenPort
	}
	apiData := &api.ApiData{
		AzureMaps: azureMaps,
		Stats: &types.JsonStats{
			StartTime:     time.Now().UTC().Format(time.RFC3339),
			ResponseCodes: make(map[string]uint32, 0),
		},
	}
	api.ServeRestEndpoints(fmt.Sprintf("%s:%d", listenIP, listenPort), apiData)
}
