package system

import (
	"bufio"
	"fmt"
	"os"

	"github.com/inconshreveable/mousetrap"
)

// PauseBeforeExit keeps the console open for double-click launches on Windows
// so configuration and startup errors stay visible to the user.
func PauseBeforeExit() {
	if !mousetrap.StartedByExplorer() {
		return
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprint(os.Stderr, "Press Enter to exit...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
