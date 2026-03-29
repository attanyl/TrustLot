package main

import (
	"fmt"
	"os"
)

var custodians = map[string]string{
	"northriver": "NorthRiver — CSV/SFTP delivery, CUSIP identifiers",
	"atlas":      "Atlas Trust — JSON API delivery, ISIN identifiers",
	"pioneer":    "Pioneer Clearing — Fixed-width delivery, rounding quirks",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("trustlot synthetic data generator")
		fmt.Println()
		fmt.Println("Usage: synth <custodian>")
		fmt.Println()
		fmt.Println("Available custodians:")
		for name, desc := range custodians {
			fmt.Printf("  %-12s %s\n", name, desc)
		}
		os.Exit(1)
	}

	name := os.Args[1]
	desc, ok := custodians[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown custodian: %s\n", name)
		os.Exit(1)
	}

	fmt.Printf("generating synthetic data for %s\n", desc)
	fmt.Println("(not yet implemented — scaffold only)")
}
