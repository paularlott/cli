//go:build cli_watch

package cli

import "github.com/fsnotify/fsnotify"

func (c *ConfigFileBase) OnChange(handler ConfigFileChangeHandler) error {
	// Ensure the config is loaded
	if err := c.LoadData(); err != nil {
		return err
	}

	// Remember the change handler
	c.changeHandler = handler

	// If no watcher then set it up
	if c.watcher == nil {
		c.watcher, _ = fsnotify.NewWatcher()

		go func() {
			for {
				select {
				case event, ok := <-c.watcher.Events:
					if !ok {
						return
					}

					if event.Op&fsnotify.Write == fsnotify.Write {
						// Reload the config file & call the handler
						c.isLoaded = false
						c.LoadData()
						c.changeHandler()
					}

				case _, ok := <-c.watcher.Errors:
					if !ok {
						return
					}
				}
			}
		}()

		c.watcher.Add(c.fileUsed)
	}

	return nil
}
