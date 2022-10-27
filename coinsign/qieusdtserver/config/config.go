package config

type RedisConfig struct {
	Port     int32  `yaml:"port"      json:"port"`
	Host     string `yaml:"host"      json:"host"`
	Password string `yaml:"password"  json:"password"`
	DBIndex  int    `yaml:"dbindex"   json:"dbindex"`
}

type DBConfig struct {
	Host         string `yaml:"host"        json:"host"`
	Name         string `yaml:"name"        json:"name"`
	Adapter      string `yaml:"adapter"     json:"adapter"`
	User         string `yaml:"user"        json:"user"`
	Password     string `yaml:"password"    json:"password"`
	Pool         int    `yaml:"pool"        json:"pool"`
	MaxIdleConns int    `yaml:"idle_conns"  json:"idle_conns"`
	MaxOpenConns int    `yaml:"open_conns"  json:"open_conns"`
}

type MQConfig struct {
	Host     string `yaml:"host"      json:"host"`
	Port     string `yaml:"port"      json:"port"`
	User     string `yaml:"user"      json:"user"`
	Password string `yaml:"password"  json:"password"`
}

type LogConfig struct {
	Mode         string `yaml:"mode"        json:"mode"`
	Level        string `yaml:"level"       json:"level"`
	Formatter    string `yaml:"formatter"   json:"formatter"`
	LogPath      string `yaml:"log_path"    json:"log_path"`
	LogName      string `yaml:"log_name"    json:"log_name"`
	MaxAge       int    `yaml:"max_age"     json:"max_age"`
	RotationTime int    `yaml:"rotation_time"     json:"rotation_time"`
}

type ConsulConfig struct {
	ConsulAddr string `yaml:"consul_addr"     json:"consul_addr"`
	ServerAddr string `yaml:"server_addr"     json:"server_addr"`
	ServerName string `yaml:"server_name"     json:"server_name"`
	Id         string `yaml:"id"              json:"id"`
}

type UsdtRpcConfig struct {
	Host     string `json:"host" yaml:"host"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
}

type GlobalConfig struct {
	HttpPort        string        `yaml:"http_port"         json:"http_port"`
	Env             string        `yaml:"env"               json:"env"`
	WorkerId        int64         `yaml:"worker_id"         json:"worker_id"`
	GenFilePath     string        `yaml:"gen_file_path"     json:"gen_file_path"`
	LogCfg          LogConfig     `yaml:"log"               json:"log"`
	RabbitMQCfg     MQConfig      `yaml:"rabbitmq"          json:"rabbitmq"`
	ConsulCfg       ConsulConfig  `yaml:"consul_cfg"        json:"consul_cfg"`
	UsdtRpcCfg      UsdtRpcConfig `yaml:"usdt" json:"usdt"`
	LoadAddressPath string        `yaml:"load_address_path" json:"load_address_path"`
	OldAddressFile  string        `yaml:"old_address_file" json:"old_address_file"`
}
