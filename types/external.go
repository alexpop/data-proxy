package types

// Struct to unmarshal the yaml config file used by the service at start time
type YamlConfig struct {
	ListenIP   string          `yaml:"listen_ip"`
	ListenPort uint16          `yaml:"listen_port"`
	Workspaces []YamlWorkspace `yaml:"workspaces"`
}

// Struct to unmarshal the workspaces of the yaml config file
type YamlWorkspace struct {
	Id     string `yaml:"id"`
	Name   string `yaml:"name"`
	Secret string `yaml:"secret"`
}

type JsonResponse struct {
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

type VersionJson struct {
	Version string `json:"binary_version"`
	Sha256  string `json:"binary_sha256"`
}
