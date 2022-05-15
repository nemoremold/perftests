package printer

import (
	"fmt"
	"strings"
)

// Alignment indicates where the line should be aligned.
type Alignment string

const (
	// LEFT indicates the line should be aligned to the left side.
	LEFT Alignment = "left"
	// RIGHT indicates the line should be aligned to the right side.
	RIGHT Alignment = "right"
	// CENTER indicates the line should be aligned in the center.
	CENTER Alignment = "center"
)

// TableRow comprises a number of lines that each is an entry of a table.
type TableRow []Line

// Valid checks whether a table row fits into a table (both must have same number of columns).
func (r *TableRow) Valid(t *Table) bool {
	return r.ColumnsCount() == t.columnsCount
}

// Print prints a table row inside a table (with the table's column widths).
func (r *TableRow) Print(t *Table) {
	if r.Valid(t) {
		printInBox('|', func() {
			for index, entry := range *r {
				if index != 0 {
					fmt.Print("|")
				}
				entry.Print(t.columnWidths[index])
			}
		})
	}
}

// AddEntry adds an entry to the table row.
func (r *TableRow) AddEntry(entry Line) {
	*r = append(*r, entry)
}

// ColumnsCount return the number of columns in a table row.
func (r *TableRow) ColumnsCount() int {
	return len(*r)
}

// Line is a single line of string that has the alignment attribute.
type Line struct {
	// Content is the content of the line.
	Content string
	// Align is the alignment of the line.
	Align Alignment
}

// Len returns the length of the line.
func (l *Line) Len() int {
	return len(l.Content)
}

// Print prints the line inside a restricted area (width), its alignment will
// be considered when printing in that area.
func (l *Line) Print(width int) {
	if width < len(l.Content) {
		fmt.Print(l.Content)
	} else {
		switch l.Align {
		case LEFT:
			fmt.Printf("%-*v", width, l.Content)
		case RIGHT:
			fmt.Printf("%*v", width, l.Content)
		case CENTER:
			blanks := width - len(l.Content)
			left, right := blanks>>1, blanks>>1
			if blanks%2 != 0 {
				right++
			}
			fmt.Printf("%*v%v%*v", left, strings.Repeat(" ", left), l.Content, right, strings.Repeat(" ", right))
		default:
			// By default align to the left side.
			fmt.Printf("%-*v", width, l.Content)
		}
	}
}

// LineAlignLeft creates a line that aligns its content to the left side.
func LineAlignLeft(content string) Line {
	return Line{
		Content: content,
		Align:   LEFT,
	}
}

// LineAlignRight creates a line that aligns its content to the right side.
func LineAlignRight(content string) Line {
	return Line{
		Content: content,
		Align:   RIGHT,
	}
}

// LineAlignCenter creates a line that aligns its content in the center.
func LineAlignCenter(content string) Line {
	return Line{
		Content: content,
		Align:   CENTER,
	}
}
