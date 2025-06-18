package cli_toml

import (
	"github.com/paularlott/cli"

	"github.com/BurntSushi/toml"
)

type tomlConfiguration struct {
	cli.ConfigFileBase
}

func NewConfigFile(fileName *string, searchPathFunc cli.SearchPathFunc) cli.ConfigFileSource {
	cfg := &tomlConfiguration{}

	cfg.InitConfigFile()

	cfg.FileName = fileName
	cfg.SearchPath = searchPathFunc
	cfg.Unmarshal = toml.Unmarshal
	cfg.Marshal = toml.Marshal

	return cfg
}
