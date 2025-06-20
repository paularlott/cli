# Shell Completion

Shell completion is not enabled by default. To enable it, add `cli.GenerateCompletionCommand(),` to the list of subcommands in your root command:

```go
Commands: []*cli.Command{
	cli.GenerateCompletionCommand(),
},
```

Shell completion is available for Bash, Zsh, Fish and Powershell.

### Bash

```shell
# Generate the completion script
myapp completion bash > ~/.bash_completion.d/myapp
source ~/.bash_completion.d/myapp
```

### Zsh

```shell
# Generate the completion script
myapp completion zsh > "${fpath[1]}/_myapp"
```

On macOS, you may need to add these lines to your `~/.zshrc`:

```shell
autoload -U compinit
compinit
```

### Fish

```shell
myapp completion fish > ~/.config/fish/completions/myapp.fish
source ~/.config/fish/completions/myapp.fish
```

### Powershell

```shell
myapp completion powershell > ~/myapp.ps1
. ~/myapp.ps1
```