package models

// 配置信息
type EtcdConfig struct {
	// auth
	Enableauth     bool   `json:"enableauth,omitempty"`
	Username     string   `json:"username,omitempty"`
	Password     string   `json:"password,omitempty"`

	// db
	Userdsn     string   `json:"userdsn,omitempty"`
	Syncdsn     string   `json:"syncdsn,omitempty"`

	// node
	Nodeurl     string   `json:"nodeurl,omitempty"`
	Walleturl     string   `json:"walleturl,omitempty"`
	Rpcuser     string   `json:"rpcuser,omitempty"`
	Rpcpass     string   `json:"rpcpass,omitempty"`

	// agent
	Agenturl     string   `json:"agenturl,omitempty"`
	Agentuser     string   `json:"agentuser,omitempty"`
	Agentpass     string   `json:"agentpass,omitempty"`
}
