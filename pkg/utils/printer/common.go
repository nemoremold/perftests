package printer

import (
	"fmt"
	"strings"
)

// TODO: use logger to provide a better way to also export.
// PrintEmptyLine prints an empty line.
func PrintEmptyLine() {
	fmt.Println()
}

// printInBox prints content in surrounded by a symbol .
func printInBox(symbol rune, print func()) {
	fmt.Printf("%c", symbol)
	print()
	fmt.Printf("%c\n", symbol)
}

// printDivider prints a divider with provided rune.
func printDivider(width int) {
	printInBox('+', func() {
		fmt.Print(strings.Repeat("-", width))
	})
}
