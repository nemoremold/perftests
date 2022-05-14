package printer

type Sheet struct {
	formatted bool
	suggested bool

	wantedWidth    int
	width          int
	suggestedWidth int

	title  Line
	header []Line
	tables []Table
	footer []Line
}

func NewSheetPtr(want int, title Line) *Sheet {
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

func (s *Sheet) WantedWidth() int {
	return s.wantedWidth
}

func (s *Sheet) Width() int {
	return s.width
}

func (s *Sheet) SuggestedWidth() int {
	return s.suggestedWidth
}

func (s *Sheet) SetWantedWidth(want int) *Sheet {
	s.wantedWidth = want
	s.formatted = false
	return s
}

func (s *Sheet) SetTitle(title Line) *Sheet {
	s.title = title
	s.formatted = false
	s.suggested = false
	return s
}

func (s *Sheet) SetHeader(header []Line) *Sheet {
	s.header = header
	s.formatted = false
	s.suggested = false
	return s
}

func (s *Sheet) SetTables(tables []Table) *Sheet {
	s.tables = tables
	s.formatted = false
	s.suggested = false
	return s
}

func (s *Sheet) SetFooter(footer []Line) *Sheet {
	s.footer = footer
	s.formatted = false
	s.suggested = false
	return s
}

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
			for _, table := range s.tables {
				table.SetWantedWidth(s.width)
			}
		}
		s.formatted = true
	}
	return s
}

func (s *Sheet) Print() {
	s.DetermineWidth()

	// Title section.
	printDivider(s.width, '-')
	s.printLines([]Line{s.title})
	printDivider(s.width, '-')

	// Header section.
	if s.header != nil {
		s.printLines(s.header)
		printDivider(s.width, '-')
	}

	if s.footer != nil || s.tables != nil {
		printDivider(s.width, '|')
	}

	// Table sections.
	for index, table := range s.tables {
		table.Print()
		if s.footer != nil || index < len(s.tables)-1 {
			printDivider(s.width, '|')
		}
	}

	// Footer section.
	if s.footer != nil {
		printDivider(s.width, '-')
		s.printLines(s.footer)
		printDivider(s.width, '-')
	}
}

func (s *Sheet) printLines(lines []Line) {
	for _, line := range lines {
		printInBox('|', func() {
			line.Print(s.width)
		})
	}
}
