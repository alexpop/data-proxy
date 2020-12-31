package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"./types"
	"./utils"
	"gopkg.in/yaml.v2"
)

const usage = `
  ./data-proxy config.yaml
`

type AzureConfig struct {
	WksNameMap map[string]*types.YamlWorkspace
	WksIdMap   map[string]*types.YamlWorkspace
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("ERROR: Missing required argument to the yaml config file: %s", usage)
	}

	// TODO: Add options for:
	// -p --port
	// -v --version  CLI
	// --no-version  API

	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	BINARY_SHA256, err = utils.GetFileSha256(ex)
	if err != nil {
		log.Fatal(err)
	}
	configBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("ERROR: Reading the config file %s returned error: %s", os.Args[1], err.Error())
	}
	err, config := loadAndValidateYamlConfig(configBytes)
	if err != nil {
		log.Fatalf("Error loading config file %s. %s", os.Args[1], err.Error())
	}
	log.Printf("Successfully loaded config yaml with %d workspaces", len(config.WksIdMap))
	serveRestEndpoints(4000, config)
}

func loadAndValidateYamlConfig(configBytes []byte) (err error, config *AzureConfig) {
	configYaml := types.YamlConfig{}
	err = yaml.Unmarshal(configBytes, &configYaml)
	if err != nil {
		return err, nil
	}

	config = &AzureConfig{
		WksIdMap:   make(map[string]*types.YamlWorkspace, 0),
		WksNameMap: make(map[string]*types.YamlWorkspace, 0),
	}

	if len(configYaml.Workspaces) == 0 {
		return errors.New("must have at least one workspace, aborting"), nil
	}
	for _, workspaceYaml := range configYaml.Workspaces {
		workspace := workspaceYaml
		if workspace.Id == "" {
			return fmt.Errorf("missing id for workspace %+v", workspace), nil
		} else if !regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$").MatchString(workspace.Id) {
			return fmt.Errorf("%s is not a valid workspace ID", workspace.Id), nil
		}
		if _, ok := config.WksIdMap[workspace.Id]; ok {
			return fmt.Errorf("found duplicate workspace id (%s)", workspace.Id), nil
		}
		if workspace.Secret == "" {
			return fmt.Errorf("missing secret for workspace %s", workspace.Id), nil
		}
		config.WksIdMap[workspace.Id] = &workspace
		if workspace.Name != "" {
			if _, ok := config.WksNameMap[workspace.Name]; ok {
				return fmt.Errorf("found duplicate workspace name (%s)", workspace.Name), nil
			}
			config.WksNameMap[workspace.Name] = &workspace
		}
	}
	return nil, config
}
