package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"./jlog"
	"./types"
	"./utils"
	"gopkg.in/yaml.v2"
)

const usage = `
  ./data-proxy config.yaml
`

func main() {
	if len(os.Args) < 2 {
		jlog.Fatal(fmt.Sprintf("ERROR: Missing required argument to the yaml config file: %s", usage))
	}

	// TODO: Add options for:
	// -v --version  CLI
	// --no-version  API

	ex, err := os.Executable()
	if err != nil {
		jlog.Fatal(err.Error())
	}
	BINARY_SHA256, err = utils.GetFileSha256(ex)
	if err != nil {
		jlog.Fatal(err.Error())
	}
	configBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		jlog.Fatal(fmt.Sprintf("ERROR: Reading the config file %s returned error: %s", os.Args[1], err.Error()))
	}
	err, yamlConfig, azureMaps := loadAndValidateYamlConfig(configBytes)
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
	apiData := &ApiData{
		AzureMaps: azureMaps,
		Stats: &types.JsonStats{
			StartTime: time.Now().UTC().Format(time.RFC3339),
		},
	}
	serveRestEndpoints(fmt.Sprintf("%s:%d", listenIP, listenPort), apiData)
}

func loadAndValidateYamlConfig(configBytes []byte) (err error, yamlConfig *types.YamlConfig, azureMaps *types.AzureMaps) {
	yamlConfig = &types.YamlConfig{}
	err = yaml.Unmarshal(configBytes, &yamlConfig)
	if err != nil {
		return err, nil, nil
	}

	azureMaps = &types.AzureMaps{
		WksIdMap:   make(map[string]*types.YamlWorkspace, 0),
		WksNameMap: make(map[string]*types.YamlWorkspace, 0),
	}

	if len(yamlConfig.Workspaces) == 0 {
		return errors.New("must have at least one workspace, aborting"), nil, nil
	}
	for _, workspaceYaml := range yamlConfig.Workspaces {
		workspace := workspaceYaml
		if workspace.Id == "" {
			return fmt.Errorf("missing id for workspace %+v", workspace), nil, nil
		} else if !regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$").MatchString(workspace.Id) {
			return fmt.Errorf("%s is not a valid workspace ID", workspace.Id), nil, nil
		}
		if _, ok := azureMaps.WksIdMap[workspace.Id]; ok {
			return fmt.Errorf("found duplicate workspace id (%s)", workspace.Id), nil, nil
		}
		if workspace.Secret == "" {
			return fmt.Errorf("missing secret for workspace %s", workspace.Id), nil, nil
		}
		azureMaps.WksIdMap[workspace.Id] = &workspace
		if workspace.Name != "" {
			if _, ok := azureMaps.WksNameMap[workspace.Name]; ok {
				return fmt.Errorf("found duplicate workspace name (%s)", workspace.Name), nil, nil
			}
			azureMaps.WksNameMap[workspace.Name] = &workspace
		}
	}
	return nil, yamlConfig, azureMaps
}
