package main

import (
	"testing"

	"./types"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigUnmarshallError(t *testing.T) {
	err, yamlConf, azureConf := loadAndValidateYamlConfig([]byte(`
bad
`))
	assert.Nil(t, yamlConf)
	assert.Nil(t, azureConf)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!str `bad` into types.YamlConfig")
}

func TestLoadConfigOneWorkspaceError(t *testing.T) {
	err, _, azureConf := loadAndValidateYamlConfig([]byte(`
hello: world
`))
	assert.Nil(t, azureConf)
	assert.EqualError(t, err, "must have at least one workspace, aborting")
}

func TestLoadConfigInvalidIdError(t *testing.T) {
	err, _, azureConf := loadAndValidateYamlConfig([]byte(`
workspaces:
- id: two
`))
	assert.Nil(t, azureConf)
	assert.EqualError(t, err, "two is not a valid workspace ID")
}

func TestLoadConfigMissingIdError(t *testing.T) {
	err, _, azureConf := loadAndValidateYamlConfig([]byte(`
workspaces:
- name: one
`))
	assert.Nil(t, azureConf)
	assert.EqualError(t, err, "missing id for workspace {Id: Name:one Secret:}")
}

func TestLoadConfigMissingSecretError(t *testing.T) {
	err, _, azureConf := loadAndValidateYamlConfig([]byte(`
workspaces:
- id: 01234567-3ca5-4b65-8383-c12a5cda28b3
`))
	assert.Nil(t, azureConf)
	assert.EqualError(t, err, "missing secret for workspace 01234567-3ca5-4b65-8383-c12a5cda28b3")
}

func TestLoadConfigDupeIdError(t *testing.T) {
	err, _, azureConf := loadAndValidateYamlConfig([]byte(`
workspaces:
- id: 01234567-3ca5-4b65-8383-c12a5cda28b3
  name: one
  secret: sssss
- id: 01234567-3ca5-4b65-8383-c12a5cda28b3
  name: two
  secret: sssss
`))
	assert.Nil(t, azureConf)
	assert.EqualError(t, err, "found duplicate workspace id (01234567-3ca5-4b65-8383-c12a5cda28b3)")
}

func TestLoadConfigDupeNameError(t *testing.T) {
	err, _, azureConf := loadAndValidateYamlConfig([]byte(`
workspaces:
- id: 01234567-3ca5-4b65-8383-c12a5cda28b3
  name: one
  secret: sssss
- id: 51234567-3ca5-4b65-8383-c12a5cda28b4
  name: one
  secret: sssss
`))
	assert.Nil(t, azureConf)
	assert.EqualError(t, err, "found duplicate workspace name (one)")
}

func TestLoadConfigSuccess(t *testing.T) {
	err, yamlConf, azureConf := loadAndValidateYamlConfig([]byte(`
listen_ip: 192.168.100.100  # defaults to 127.0.0.1 if not specified
listen_port: 50000  # defaults to 4000 if not specified
workspaces:
- id: 01234567-3ca5-4b65-8383-c12a5cda28b3
  name: one
  secret: "secret one"
# comments are ok in the config file
- id: 51234567-3ca5-4b65-8383-c12a5cda28b4
  secret: "secret two"
`))
	assert.Nil(t, err)
	assert.NotNil(t, azureConf)
	wks1 := &types.YamlWorkspace{
		Id:     "01234567-3ca5-4b65-8383-c12a5cda28b3",
		Name:   "one",
		Secret: "secret one",
	}
	wks2 := &types.YamlWorkspace{
		Id:     "51234567-3ca5-4b65-8383-c12a5cda28b4",
		Secret: "secret two",
	}
	expectedIdMap := make(map[string]*types.YamlWorkspace, 2)
	expectedIdMap[wks1.Id] = wks1
	expectedIdMap[wks2.Id] = wks2
	expectedNameMap := make(map[string]*types.YamlWorkspace, 2)
	expectedNameMap[wks1.Name] = wks1
	expectedAzureConfig := &types.AzureMaps{
		WksIdMap:   expectedIdMap,
		WksNameMap: expectedNameMap,
	}
	assert.Equal(t, expectedAzureConfig, azureConf)
	assert.Equal(t, "192.168.100.100", yamlConf.ListenIP)
	assert.Equal(t, uint16(50000), yamlConf.ListenPort)
}
