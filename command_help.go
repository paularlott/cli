package cli

import (
	"fmt"
	"strings"
)

func (c *Command) ShowHelp() {
	// Make the command name from the chain of commands
	chain := []string{}
	for _, cmd := range c.commandChain {
		chain = append(chain, cmd.Name)
	}
	cmdName := strings.Join(chain, " ")

	// Display name and version
	fmt.Printf("Name:\n   %s", cmdName)
	if c.Usage != "" {
		fmt.Printf(" - %s", c.Usage)
	}
	fmt.Println()
	fmt.Println()

	// Display usage
	fmt.Println("Usage:")
	usageString := fmt.Sprintf("   %s", cmdName)

	// Add flags indicator if we have flags
	if len(c.Flags) > 0 {
		usageString += " [flags]"
	}

	// Add arguments
	for _, arg := range c.Arguments {
		if arg.isRequired() {
			usageString += fmt.Sprintf(" <%s>", arg.name())
		} else {
			usageString += fmt.Sprintf(" [%s]", arg.name())
		}
	}

	// Add command placeholder if subcommands exist
	if len(c.Commands) > 0 {
		usageString += " [command]"
	}

	if c.MaxArgs > 0 || c.MaxArgs == UnlimitedArgs {
		if c.MinArgs > 0 {
			usageString += " <args...>"
		} else {
			usageString += " [args...]"
		}
	}

	fmt.Println(usageString)
	fmt.Println()

	// Display version if available
	if c.Version != "" {
		fmt.Printf("Version:\n   %s\n\n", c.Version)
	}

	// Display detailed description if available
	if c.Description != "" {
		fmt.Println("Description:")
		paragraphs := strings.Split(c.Description, "\n\n")
		for _, para := range paragraphs {
			fmt.Printf("   ")
			c.printWrappedText(strings.TrimSpace(para), 3, 80)
			fmt.Println("\n")
		}
	}

	// Display subcommands if any
	if len(c.Commands) > 0 {
		fmt.Println("Available Commands:")
		for _, cmd := range c.Commands {
			fmt.Printf("   %-15s %s\n", cmd.Name, cmd.Usage)
		}
		fmt.Println()
	}

	// Group flags into local and global
	localFlags := []Flag{}
	globalFlags := []Flag{}

	for _, flag := range c.Flags {
		if flag.isHidden() {
			continue
		}

		if flag.isGlobal() {
			globalFlags = append(globalFlags, flag)
		} else {
			localFlags = append(localFlags, flag)
		}
	}

	// Add found global flags from parent commands
	for _, flag := range c.globalFlags {
		if flag.isHidden() {
			continue
		}

		globalFlags = append(globalFlags, flag)
	}

	// Display local flags if any
	if len(localFlags) > 0 {
		fmt.Println("Flags:")
		c.displayFormattedFlags(localFlags)
		fmt.Println()
	}

	// Display global flags if any
	if len(globalFlags) > 0 {
		fmt.Println("Global Flags:")
		c.displayFormattedFlags(globalFlags)
		fmt.Println()
	}

	// Display arguments if any
	if len(c.Arguments) > 0 {
		fmt.Println("Arguments:")

		// Find maximum width for argument names to align descriptions
		maxArgWidth := 0
		for _, arg := range c.Arguments {
			argNameWithType := arg.name() + " " + arg.typeText()
			if len(argNameWithType) > maxArgWidth {
				maxArgWidth = len(argNameWithType)
			}
		}

		// Add some padding (at least 2 spaces)
		maxArgWidth += 2

		// Ensure we don't go beyond a reasonable width
		if maxArgWidth > 40 {
			maxArgWidth = 40
		}

		for _, arg := range c.Arguments {
			prefix := "   "
			if arg.isRequired() {
				prefix = " \033[1;31m*\033[0m "
			}

			argNameWithType := arg.name() + " " + arg.typeText()

			// Truncate if too long (similar to flag handling)
			if len(argNameWithType) > maxArgWidth-2 && maxArgWidth == 40 {
				argNameWithType = argNameWithType[:maxArgWidth-5] + "..."
			}

			// Print argument name and type with padding
			fmt.Printf("%s%-*s", prefix, maxArgWidth, argNameWithType)

			// Print the description with proper wrapping
			c.printWrappedText(arg.usage(), maxArgWidth+3, 80)
			fmt.Println()
		}
		fmt.Println()
	}
}

func (c *Command) displayFormattedFlags(flags []Flag) {
	// Find maximum width for flag definitions to align descriptions
	maxDefWidth := 0
	for _, flag := range flags {
		defWidth := len(flag.flagDefinition())
		if defWidth > maxDefWidth {
			maxDefWidth = defWidth
		}
	}

	// Add some padding (at least 2 spaces)
	maxDefWidth += 2

	// Ensure we don't go beyond a reasonable width
	if maxDefWidth > 40 {
		maxDefWidth = 40
	}

	for _, flag := range flags {
		def := flag.flagDefinition()
		desc := flag.getUsage()
		defaultValue := flag.defaultValueText()

		// Truncate definition if it's too long
		if len(def) > maxDefWidth-2 && maxDefWidth == 40 {
			def = def[:maxDefWidth-5] + "..."
		}

		prefix := "   "
		if flag.isRequired() {
			prefix = " \033[1;31m*\033[0m "
		}

		// Print flag definition with padding
		fmt.Printf("%s%-*s", prefix, maxDefWidth, def)

		// Add default value if available
		if defaultValue != "" {
			desc += fmt.Sprintf(" (default: %s)", defaultValue)
		}

		// Print the description with proper wrapping
		c.printWrappedText(desc, maxDefWidth+3, 80)
		fmt.Println()
	}
}

// Helper function to print wrapped text with proper indentation
func (c *Command) printWrappedText(text string, indent, width int) {
	// Calculate available width for text
	availWidth := width - indent

	// If text fits on one line, just print it
	if len(text) <= availWidth {
		fmt.Print(text)
		return
	}

	// Special handling for text with default values
	// Extract default value part if present
	var defaultPart string
	mainText := text

	// Check if text contains default value pattern
	if idx := strings.LastIndex(text, " (default:"); idx != -1 {
		mainText = strings.Trim(text[:idx], " ")
		defaultPart = text[idx:]
	}

	// Process main text
	words := strings.Split(mainText, " ")
	line := ""
	firstLine := true

	for _, word := range words {
		// Check if adding this word would exceed available width
		if len(line)+len(word)+1 > availWidth {
			// Print current line
			if firstLine {
				fmt.Print(line)
				firstLine = false
			} else {
				fmt.Printf("\n%s%s", strings.Repeat(" ", indent), line)
			}
			line = word
		} else {
			if line == "" {
				line = word
			} else {
				line += " " + word
			}
		}
	}

	// Print the last line of main text
	if line != "" {
		if firstLine {
			fmt.Print(line)
			firstLine = false
		} else {
			fmt.Printf("\n%s%s", strings.Repeat(" ", indent), line)
		}
	}

	// Handle default value part if present
	if defaultPart != "" {
		// If default part fits on current line, append it
		if !firstLine && len(line)+len(defaultPart) <= availWidth {
			fmt.Print(defaultPart)
		} else {
			// Otherwise, put default part on its own line
			fmt.Printf("\n%s%s", strings.Repeat(" ", indent), strings.Trim(defaultPart, " "))
		}
	}
}
