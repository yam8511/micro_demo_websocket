package main

// Conf 全域設定變數
var Conf *Config

// App 應用環境
type App struct {
	Env  string `toml:"env"`
	IP   string `toml:"ip"`
	Host string `toml:"host"`
	Port string `toml:"port"`
	Site string `toml:"site"`
}

// Pusher 型態
type Pusher struct {
	ID       uint64
	Name     string `toml:"name"`
	IP       string `toml:"ip"`
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	APIKey   string `toml:"api_key"`
	MaxLink  int64  `toml:"max_link"`
	ChatRoom string `toml:"chat_room"`
	RPC      RPC    `toml:"rpc"`
}

// RPCInfo RPC資訊
type RPCInfo struct {
	Service string `toml:"service"`
	Method  string `toml:"method"`
}

// RPC 微服務資訊
type RPC struct {
	Register   RPCInfo `toml:"register"`
	Deregister RPCInfo `toml:"deregister"`
}

// Config 型態
type Config struct {
	App    App    `toml:"app"`
	Pusher Pusher `toml:"pusher"`
}

// PushData 推送資訊
type PushData struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// UserSocket 型態
type UserSocket struct {
	UserID       int64 `json:"user_id"`
	HandicapType int64 `json:"handicap_type"`
	LineType     int64 `json:"line_type"`
	Channel      chan []byte
}
