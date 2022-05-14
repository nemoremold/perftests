package printer

type Table struct {
	formatted bool
	suggested bool

	wantedWidth    int
	width          int
	suggestedWidth int

	columnsCount          int
	columnWidths          []int
	suggestedColumnWidths []int

	title   Line
	indexes TableRow
	datum   []TableRow
}

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
		indexes:        nil,
		datum:          nil,
	}
}

func (t *Table) WantedWidth() int {
	return t.wantedWidth
}

func (t *Table) Width() int {
	return t.width
}

func (t *Table) SuggestedWidth() int {
	return t.suggestedWidth
}

func (t *Table) ColumnsCount() int {
	return t.columnsCount
}

func (t *Table) SetWantedWidth(want int) *Table {
	t.wantedWidth = want
	t.formatted = false
	return t
}

func (t *Table) SetTitle(title Line) *Table {
	t.title = title
	t.formatted = false
	t.suggested = false
	return t
}

func (t *Table) SetIndexes(indexes TableRow) *Table {
	if indexes.Valid(t) {
		t.indexes = indexes
		t.formatted = false
		t.suggested = false
	}
	return t
}

func (t *Table) SetDatum(datum []TableRow) *Table {
	if t.AreValidRows(datum) {
		t.datum = datum
		t.formatted = false
		t.suggested = false
	}
	return t
}

func (t *Table) AreValidRows(rows []TableRow) bool {
	for _, row := range rows {
		if !row.Valid(t) {
			return false
		}
	}
	return true
}

func (t *Table) SuggestWidth() *Table {
	if !t.suggested {
		suggestedWidth := 0
		suggestedColumnWidths := make([]int, t.columnsCount)
		for index, indexName := range t.indexes {
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

func (t *Table) Print() {
	t.DetermineWidth()

	// Title section.
	printDivider(t.width)
	printInBox('|', func() {
		t.title.Print(t.width)
	})
	printDivider(t.width)

	// Index section.
	if t.indexes != nil {
		t.indexes.Print(t)
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
