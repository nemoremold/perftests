package printer

import (
	"fmt"
	"strings"
)

// Sheet is an output format where there is a title, followed by a header section,
// followed by a table sections with a number of tables, followed by a footer section.
type Sheet struct {
	// formatted is `true` when the sheet has been formatted (and therefore is able
	// to be printed).
	formatted bool
	// suggested is `true` when the sheet has proposed the minimal width it needs to
	// have. A suggested width must exist before formatting the sheet.
	suggested bool

	// wantedWidth is the width of the sheet if it is greater than `suggestedWidth`.
	wantedWidth int
	// width is the actual width of the sheet.
	width int
	// suggestedWidth is calculated with the widths of all elements in the sheet.
	// It is the width of the sheet if it is greater than `wantedWidth`.
	suggestedWidth int

	// title is a single line string.
	title Line
	// header is a section with several lines of strings.
	header []Line
	// tables is a section with several tables.
	tables []Table
	// footer is a section with several lines of strings.
	footer []Line
}

// NewSheet instantiates a new sheet instance.
func NewSheet(want int, title Line) *Sheet {
	return &Sheet{
		formatted:      false,
		suggested:      false,
		wantedWidth:    want,
		width:          0,
		suggestedWidth: 0,
		title:          title,
		header:         nil,
		tables:         nil,
		footer:         nil,
	}
}

// WantedWidth returns the current wanted width of the sheet.
func (s *Sheet) WantedWidth() int {
	return s.wantedWidth
}

// Width returns the current width of the sheet.
func (s *Sheet) Width() int {
	return s.width
}

// SuggestedWidth returns the current suggested width of the sheet.
func (s *Sheet) SuggestedWidth() int {
	return s.suggestedWidth
}

// SetWantedWidth sets the `wantedWidth` of the sheet, marking the sheet
// as un-formatted.
func (s *Sheet) SetWantedWidth(want int) *Sheet {
	s.wantedWidth = want
	s.formatted = false
	return s
}

// SetTitle sets the title section of the sheet, marking the sheet
// as un-formatted and un-suggested (sheet element changed).
func (s *Sheet) SetTitle(title Line) *Sheet {
	s.title = title
	s.formatted = false
	s.suggested = false
	return s
}

// SetTitle sets the header section of the sheet, marking the sheet
// as un-formatted and un-suggested (sheet element changed).
func (s *Sheet) SetHeader(header []Line) *Sheet {
	s.header = header
	s.formatted = false
	s.suggested = false
	return s
}

// SetTitle sets the table section of the sheet, marking the sheet
// as un-formatted and un-suggested (sheet element changed).
func (s *Sheet) SetTables(tables []Table) *Sheet {
	s.tables = tables
	s.formatted = false
	s.suggested = false
	return s
}

// SetTitle sets the footer section of the sheet, marking the sheet
// as un-formatted and un-suggested (sheet element changed).
func (s *Sheet) SetFooter(footer []Line) *Sheet {
	s.footer = footer
	s.formatted = false
	s.suggested = false
	return s
}

// SuggestWidth calculates all elements' suggested widths and proposes a
// minimal width for the sheet.
func (s *Sheet) SuggestWidth() *Sheet {
	if !s.suggested {
		for _, table := range s.tables {
			if s.suggestedWidth < table.SuggestWidth().SuggestedWidth() {
				s.suggestedWidth = table.SuggestWidth().SuggestedWidth()
			}
		}

		lines := make([]Line, 0)
		lines = append(lines, s.title)
		lines = append(lines, s.header...)
		lines = append(lines, s.footer...)
		for _, line := range lines {
			if s.suggestedWidth < line.Len() {
				s.suggestedWidth = line.Len()
			}
		}
		s.suggested = true
	}
	return s
}

// DetermineWidth compares the suggested width and the wanted width, determining
// the width of the sheet. If the width of the sheet has changed, the sheet will
// enforce the new width on all of its components.
func (s *Sheet) DetermineWidth() *Sheet {
	if !s.formatted {
		if !s.suggested {
			s.SuggestWidth()
		}

		previous := s.width
		if s.wantedWidth > s.suggestedWidth {
			s.width = s.wantedWidth
		} else {
			s.width = s.suggestedWidth
		}

		if previous != s.width {
			for index, table := range s.tables {
				s.tables[index] = *table.SetWantedWidth(s.width)
			}
		}
		s.formatted = true
	}
	return s
}

// Print prints the sheet.
func (s *Sheet) Print() {
	s.DetermineWidth()

	// Title section.
	printDivider(s.width)
	s.printLines([]Line{s.title})
	printDivider(s.width)

	// Header section.
	if s.header != nil {
		s.printLines(s.header)
		printDivider(s.width)
	}

	if s.footer != nil || s.tables != nil {
		printInBox('|', func() {
			fmt.Print(strings.Repeat("|", s.width))
		})
	}

	// Table sections.
	for index, table := range s.tables {
		table.Print()
		if s.footer != nil || index < len(s.tables)-1 {
			printInBox('|', func() {
				fmt.Print(strings.Repeat("|", s.width))
			})
		}
	}

	// Footer section.
	if s.footer != nil {
		printDivider(s.width)
		s.printLines(s.footer)
		printDivider(s.width)
	}
}

// printLines prints a line of the sheet with surrounded `|`.
func (s *Sheet) printLines(lines []Line) {
	for _, line := range lines {
		printInBox('|', func() {
			line.Print(s.width)
		})
	}
}
