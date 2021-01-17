package compose

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type CmdConfig struct {
	Cmd          []string
	Name         string
	ReadyWhenLog string        // when log contain this, switch to ready
	ReadyTimeout time.Duration // after this timeout, switch to ready
}

var DefaultReadyTimeout time.Duration = time.Second * 10

func Parse(file string) ([]*CmdConfig, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	configs := make([]*CmdConfig, 0)
	err = json.Unmarshal(content, &configs)
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.ReadyTimeout != 0 {
			config.ReadyTimeout *= time.Second
		}
	}

	return configs, nil
}
