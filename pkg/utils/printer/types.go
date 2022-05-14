package printer

import (
	"fmt"
	"strings"
)

type Alignment string

const (
	LEFT   Alignment = "left"
	RIGHT  Alignment = "right"
	CENTER Alignment = "center"
)

type TableRow []Line

func (r TableRow) Valid(t *Table) bool {
	return len(r) == t.columnsCount
}

func (r TableRow) Print(t *Table) {
	if r.Valid(t) {
		printInBox('|', func() {
			for index, entry := range r {
				if index != 0 {
					fmt.Print("|")
				}
				entry.Print(t.columnWidths[index])
			}
		})
	}
}

func (r TableRow) AddEntry(entry Line) TableRow {
	return append(r, entry)
}

func (r TableRow) ColumnsCount() int {
	return len(r)
}

type Line struct {
	Content string
	Align   Alignment
}

func (l *Line) Len() int {
	return len(l.Content)
}

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
			// By default align left.
			fmt.Printf("%-*v", width, l.Content)
		}
	}
}

func LineAlignLeft(content string) Line {
	return Line{
		Content: content,
		Align:   LEFT,
	}
}

func LineAlignRight(content string) Line {
	return Line{
		Content: content,
		Align:   RIGHT,
	}
}

func LineAlignCenter(content string) Line {
	return Line{
		Content: content,
		Align:   CENTER,
	}
}
