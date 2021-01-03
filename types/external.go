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

type JsonVersion struct {
	Version string `json:"binary_version"`
	Sha256  string `json:"binary_sha256"`
}

type JsonStats struct {
	StartTime     string            `json:"start_time"`
	ResponseCodes map[string]uint32 `json:"response_codes"`
}

type JsonRootLog struct {
	Info     string        `json:"info,omitempty"`
	Debug    string        `json:"debug,omitempty"`
	Error    string        `json:"error,omitempty"`
	ProxyLog *JsonProxyLog `json:"proxy,omitempty"`
	Time     string        `json:"time"`
}

type JsonProxyLog struct {
	Status uint16 `json:"status"`
	Body   string `json:"body,omitempty"`
	Method string `json:"method"`
	URI    string `json:"uri"`
	IP     string `json:"ip,omitempty"`
}
