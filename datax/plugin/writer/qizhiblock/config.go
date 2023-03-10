package qizhiblock

import (
	"encoding/json"
	"github.com/Breeze0806/go-etl/config"
)

type Config struct {
	RestApiEndpoint  string `json:"restApiEndpoint"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Channel          string `json:"channel"`
	ChainCode        string `json:"chainCode"`
	ChainCodeFunction string `json:"chainCodeFunction"`
	KeyColumn        int    `json:"keyColumn"`
}

//NewConfig 通过json配置conf获取csv输入配置
func NewConfig(conf *config.JSON) (*Config, error) {
	c := &Config{}
	if err := json.Unmarshal([]byte(conf.String()), c); err != nil {
		return nil, err
	}
	return c, nil
}
