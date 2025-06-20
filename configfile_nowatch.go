//go:build !cli_watch

package cli

import "fmt"

func (c *ConfigFileBase) OnChange(handler ConfigFileChangeHandler) error {
	return fmt.Errorf("OnChange not supported: build with '-tags cli_watch' tag to enable")
}
