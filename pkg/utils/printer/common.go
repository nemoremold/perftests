package printer

import (
	"fmt"
	"strings"
)

// PrintEmptyLine prints an empty line.
func PrintEmptyLine() {
	fmt.Println()
}

// printInBox prints content in surrounded by a symbol .
func printInBox(symbol rune, print func()) {
	fmt.Print(fmt.Sprint(symbol))
	print()
	fmt.Print(fmt.Sprint(symbol) + "\n")
}

// printDivider prints a divider with provided rune.
func printDivider(width int, symbol rune) {
	printInBox('+', func() {
		fmt.Printf("%*v", width, strings.Repeat(fmt.Sprint(symbol), width))
	})
}
