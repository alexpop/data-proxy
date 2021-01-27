package api

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/alexpop/data-proxy/types"
	"gopkg.in/yaml.v2"
)

func LoadAndValidateYamlConfig(configBytes []byte) (err error, yamlConfig *types.YamlConfig, azureMaps *types.AzureMaps) {
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
