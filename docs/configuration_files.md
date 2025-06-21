# Configuration Files

Flags can be set using a configuration file. The configuration file format is typically TOML, YAML, or JSON and is supplied to the root command as a file reader.

The configuration file reader takes a pointer to a variable which will hold the name of the configuration file to load, this allows flags to set the configuration file. The reader can accept a function that returns a list of paths to search for the configuration file. The search path is tried if the configuration file can't be opened on the first attempt.

The following example provides a TOML configuration file reader, that searches for the config file in multiple locations.

```go
var configFile = "example.toml"

cmd := &cli.Command{
  ConfigFile: cli_toml.NewConfigFile(&configFile, func() []string {
    // Look for the config file in:
    //   - The current directory
    //   - The user's home directory
    //   - The user's .config directory

    paths := []string{"."}

    home, err := os.UserHomeDir()
    if err == nil {
      paths = append(paths, home)
    }

    paths = append(paths, filepath.Join(home, ".config"))
    paths = append(paths, filepath.Join(home, ".config", "example"))

    return paths
  }),
  Flags: []cli.Flag{
    &cli.StringFlag{
      Name:     "config",
      AssignTo: &configFile,
      Global:   true,
    },
    &cli.StringFlag{
      Name:     "listen",
      Usage:    "The address and port to listen on",
      ConfigPath: []string{"server.listen"},
    },
  },
}
```

The `ConfigPath` field on the `Flag` struct can be used to specify the path to data within the configuration file. Where multiple paths are specified, the first one found will be used. In the above example `server.listen` will use the `listen` value from the `server` section of the configuration file.

```toml
[server]
listen = ":8080"
```

## Watching for Changes

If the library is built with the tag `cli_watch` then it's possible to watch the configuration file for changes and act upon those changes.

The `OnChange` function is called against the ConfigFile object to register a handler and enable watching. The registered handler is called when the configuration file has been written and reloaded.

```go
cmd.ConfigFile.OnChange(func() {
  fmt.Println("Config file changed:", cmd.ConfigFile.FileUsed())
  cmd.ReloadFlags()
})
```

The handler can optionally call `ReloadFlags` on the root command to refresh the flag values, when the flags are reloaded any variables that flags are assigned to are updated.

## Accessing Data

While the configuration file is designed to be used for supplying data to flags, it's also possible to read and write data to the configuration file directly through the `GetValue` and `SetValue` functions.

A value can be updated as follows:

```go
cfg := cmd.ConfigFile

cfg.SetValue("server.listen", ":8080")
cfg.Save()
```

The `GetKeys` function can be called to get a list of keys within a configuration path:

```go
keys := cfg.GetKeys("server")
```

This would return the key `listen` along with any other keys in the `server` section.

Keys can also be deleted with the `DeleteKey` function, once the key has been deleted `Save` must be called to updated the configuration file.

## Adding File Readers

File readers are designed to be simple to allow additional file formats to be supported with minimal effort.

The following reader supports configuration files in JSON format:

```go
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
```
