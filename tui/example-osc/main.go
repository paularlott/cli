// OSC sequence test - outputs directly to terminal without TUI.
//
// Run with:
//
//	go run ./tui/example-osc [--osc777] [--persistent]
package main

import (
	"fmt"
	"os"

	"github.com/paularlott/cli/tui"
)

func main() {
	// useOSC777 := flag.Bool("osc777", false, "use OSC 777 (NotifyWithTitle) instead of OSC 9")
	// persistent := flag.Bool("persistent", false, "use persistent notification style (Ghostty on macOS)")
	// flag.Parse()

	// style := tui.NotifyTransient
	// if *persistent {
	// 	style = tui.NotifyPersistent
	// }

	fmt.Println("=== OSC Sequence Test ===")
	fmt.Println()

	// Test 1: Window Title
	fmt.Println("1. Setting window title to 'OSC Test'...")
	os.Stdout.WriteString(tui.SetWindowTitle("OSC Test"))
	fmt.Println("   (Check your terminal tab/window title)")

	// Test 2: Hyperlink
	fmt.Println()
	fmt.Println("2. Testing hyperlink...")
	fmt.Printf("   Click this: %s\n", tui.Hyperlink("https://github.com", "GitHub"))
	fmt.Printf("   And this: %s\n", tui.Hyperlink("https://google.com", "Google"))

	// Test 3: Notification
	fmt.Println()
	fmt.Println("3. Testing notifications...")

	fmt.Println(tui.Notify("Hello from OSC test!"))

	fmt.Println()
	fmt.Print("Press Enter to exit...")
	os.Stdin.Read([]byte{0})
}
