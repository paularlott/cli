module github.com/paularlott/cli/examples/config_file

go 1.25.3

replace github.com/paularlott/cli => ../../

require github.com/paularlott/cli v0.0.0-00010101000000-000000000000

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
)
