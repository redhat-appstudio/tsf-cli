package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// printDisclaimer emits a support-level disclaimer to stderr on interactive
// sessions. Suppressed when TSSC_NO_DISCLAIMER is set or stderr is not a
// terminal.
func printDisclaimer() {
	if os.Getenv("TSF_NO_DISCLAIMER") != "" {
		return
	}
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return
	}
	fmt.Fprintf(
		os.Stderr,
		"NOTE: The TSF installation program generates your first deployment "+
			"of RHADS-TSF but does not support upgrades. "+
			"Each products must be manually reconfigured for production workloads.\n\n",
	)
}
