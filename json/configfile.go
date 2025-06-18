package cli_json

import (
	"encoding/json"

	"github.com/paularlott/cli"
)

type jsonConfiguration struct {
	cli.ConfigFileBase
}

func NewConfigFile(fileName *string, searchPathFunc cli.SearchPathFunc) cli.ConfigFileSource {
	cfg := &jsonConfiguration{}

	cfg.InitConfigFile()

	cfg.FileName = fileName
	cfg.SearchPath = searchPathFunc
	cfg.Unmarshal = json.Unmarshal
	cfg.Marshal = json.Marshal

	return cfg
}
