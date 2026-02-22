# Shell Completion

Shell completion is not enabled by default. To enable it, add `cli.GenerateCompletionCommand()` to the list of subcommands in your root command:

```go
Commands: []*cli.Command{
	cli.GenerateCompletionCommand(),
},
```

Shell completion is available for Bash, Zsh, Fish and PowerShell.

### Bash

To load completions for the current session:

```shell
source <(myapp completion bash)
```

To load completions for each session, execute once:

```shell
myapp completion bash > ~/.bash_completion
```

### Zsh

To load completions for the current session:

```shell
source <(myapp completion zsh)
```

To load completions for each session, execute once:

```shell
myapp completion zsh > "${fpath[1]}/_myapp"
```

### Fish

To load completions for the current session:

```shell
myapp completion fish | source
```

To load completions for each session, execute once:

```shell
myapp completion fish > ~/.config/fish/completions/myapp.fish
```

### PowerShell

To load completions for the current session:

```shell
myapp completion powershell | Out-String | Invoke-Expression
```

To load completions for each session, add to your PowerShell profile (`$PROFILE`):

```shell
myapp completion powershell | Out-String | Invoke-Expression
```

Or save to a file and dot-source it from your profile:

```shell
myapp completion powershell > "$HOME/.myapp_completions.ps1"
# Add to $PROFILE: . "$HOME/.myapp_completions.ps1"
```
