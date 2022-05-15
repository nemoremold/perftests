package printer

// Table is an output format where there is a title, followed by a header row,
// followed by data rows.
type Table struct {
	// formatted is `true` when the table has been formatted (and therefore is able
	// to be printed).
	formatted bool
	// suggested is `true` when the table has proposed the minimal width it needs to
	// have. A suggested width must exist before formatting the table.
	suggested bool

	// wantedWidth is the width of the table if it is greater than `suggestedWidth`.
	wantedWidth int
	// width is the actual width of the table.
	width int
	// suggestedWidth is calculated with the widths of all elements in the table.
	// It is the width of the table if it is greater than `wantedWidth`.
	suggestedWidth int

	// columnsCount is the number of columns in the table.
	columnsCount int
	// columnWidths are the actual widths of every column.
	columnWidths []int
	// columnWidths are the suggested widths of every column. The minimal width of
	// a column is the greatest element width in that column plus 1 (room for a blank space).
	suggestedColumnWidths []int

	// title is a single line string.
	title Line
	// headers is the header row of the table.
	headers TableRow
	// datum are the data rows of the table.
	datum []TableRow
}

// NewTable instantiates a new table instance.
func NewTable(want, columnsCount int, title Line) *Table {
	return &Table{
		formatted:      false,
		suggested:      false,
		wantedWidth:    want,
		width:          0,
		suggestedWidth: 0,
		columnsCount:   columnsCount,
		columnWidths:   nil,
		title:          title,
		headers:        nil,
		datum:          nil,
	}
}

// WantedWidth returns the current wanted width of the table.
func (t *Table) WantedWidth() int {
	return t.wantedWidth
}

// Width returns the current width of the table.
func (t *Table) Width() int {
	return t.width
}

// SuggestedWidth returns the current suggested width of the table.
func (t *Table) SuggestedWidth() int {
	return t.suggestedWidth
}

// ColumnsCount returns the number of the columns in the table.
func (t *Table) ColumnsCount() int {
	return t.columnsCount
}

// SetWantedWidth sets the `wantedWidth` of the table, marking the table
// as un-formatted.
func (t *Table) SetWantedWidth(want int) *Table {
	t.wantedWidth = want
	t.formatted = false
	return t
}

// SetTitle sets the title of the table, marking the table as
// un-formatted and un-suggested (table element changed).
func (t *Table) SetTitle(title Line) *Table {
	t.title = title
	t.formatted = false
	t.suggested = false
	return t
}

// SetHeaders sets the header row of the table, marking the table as
// un-formatted and un-suggested (table element changed).
func (t *Table) SetHeaders(indexes TableRow) *Table {
	if indexes.Valid(t) {
		t.headers = indexes
		t.formatted = false
		t.suggested = false
	}
	return t
}

// SetDatum sets the data rows of the table, marking the table as
// un-formatted and un-suggested (table element changed).
func (t *Table) SetDatum(datum []TableRow) *Table {
	if t.AreValidRows(datum) {
		t.datum = datum
		t.formatted = false
		t.suggested = false
	}
	return t
}

// AreValidRows checks whether the rows fit into the table.
func (t *Table) AreValidRows(rows []TableRow) bool {
	for _, row := range rows {
		if !row.Valid(t) {
			return false
		}
	}
	return true
}

// SuggestWidth calculates all elements' suggested widths and proposes a
// minimal width for the table. The minimal width of a column is the greatest
// element width in that column plus 1 (room for a blank space).
func (t *Table) SuggestWidth() *Table {
	if !t.suggested {
		suggestedWidth := 0
		suggestedColumnWidths := make([]int, t.columnsCount)
		for index, indexName := range t.headers {
			suggestedColumnWidths[index] = indexName.Len()
		}
		for column := 0; column < t.columnsCount; column++ {
			for _, row := range t.datum {
				if suggestedColumnWidths[column] < row[column].Len() {
					suggestedColumnWidths[column] = row[column].Len()
				}
			}
			suggestedColumnWidths[column]++
			suggestedWidth += suggestedColumnWidths[column]
			if column != 0 {
				suggestedWidth++
			}
		}
		t.suggestedWidth = suggestedWidth
		t.suggestedColumnWidths = suggestedColumnWidths
		t.suggested = true
	}
	return t
}

// DetermineWidth compares the suggested width and the wanted width, determining
// the width of the table. If the width of the sheet has changed, the table will
// adjusting the width of its columns accordingly.
func (t *Table) DetermineWidth() *Table {
	if !t.formatted {
		if !t.suggested {
			t.SuggestWidth()
		}

		previous := t.width
		if t.wantedWidth > t.suggestedWidth {
			t.width = t.wantedWidth
		} else {
			t.width = t.suggestedWidth
		}

		if previous != t.width {
			if t.suggestedWidth <= t.width {
				margin := t.width - t.suggestedWidth
				sharedIncrease := margin / t.columnsCount
				bonusIncrease := margin % t.columnsCount
				t.columnWidths = make([]int, len(t.suggestedColumnWidths))
				for index := range t.suggestedColumnWidths {
					t.columnWidths[index] = t.suggestedColumnWidths[index] + sharedIncrease
					if index+1 <= bonusIncrease {
						t.columnWidths[index]++
					}
				}
			}
		}
		t.formatted = true
	}
	return t
}

// Print prints the table.
func (t *Table) Print() {
	t.DetermineWidth()

	// Title section.
	printDivider(t.width)
	printInBox('|', func() {
		t.title.Print(t.width)
	})
	printDivider(t.width)

	// Index section.
	if t.headers != nil {
		t.headers.Print(t)
		printDivider(t.width)
	}

	// Datum section.
	if t.datum != nil {
		for _, row := range t.datum {
			row.Print(t)
		}
		printDivider(t.width)
	}
}
