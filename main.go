package main

import (
	"io/ioutil"
	"log"
	"os"

	"./types"
	"./utils"
	"gopkg.in/yaml.v2"
)

const usage = `
  ./data-proxy config.yaml
`

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("ERROR: Missing required argument to the yaml config file: %s", usage)
	}

	// TODO: Add options for:
	// -p --port
	// -v --version  CLI
	// --no-version  API

	configYamlPath := os.Args[1]
	configBytes, err := ioutil.ReadFile(configYamlPath)
	if err != nil {
		log.Fatalf("ERROR: Reading the config file %s returned error: %s", configYamlPath, err.Error())
	}

	configYaml := types.YamlConfig{}
	err = yaml.Unmarshal(configBytes, &configYaml)
	if err != nil {
		log.Fatalf("ERROR: Unmarshaling the config file %s returned error: %s", configYamlPath, err.Error())
	}
	log.Printf("Config file %s read successfully and identified %d workspaces!\n", configYamlPath, len(configYaml.Workspaces))

	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	BINARY_SHA256, err = utils.GetFileSha256(ex)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(BINARY_SHA256)

	serveRestEndpoints(4000)
}
