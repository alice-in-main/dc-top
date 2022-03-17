package elements

import "github.com/gdamore/tcell/v2"

var default_seperator_style = tcell.StyleDefault.Foreground(tcell.ColorGray)

func getAbsoluteWidths(total_width int, relative_widths []float64) (absolute_widths []int) {
	absolute_widths = make([]int, len(relative_widths))
	sum := 0
	for i, relative_width := range relative_widths {
		absolute_widths[i] = int(relative_width * float64(total_width))
		sum += absolute_widths[i]
	}
	absolute_widths[0] += total_width - sum // fix leftover width
	return absolute_widths
}

func TableWithHeader(total_width int, relative_widths []float64, cells [][]StringStyler, header_cells []StringStyler) (rows []StringStyler) {
	rows = Table(total_width, relative_widths, [][]StringStyler{header_cells})
	rows = append(rows, generateUnderline(total_width, relative_widths))
	rows = append(rows, Table(total_width, relative_widths, cells)...)
	return rows
}

func Table(total_width int, relative_widths []float64, cells [][]StringStyler) (rows []StringStyler) {
	const column_seperator = '\u2502'
	rows = make([]StringStyler, len(cells))
	absolute_widths := getAbsoluteWidths(total_width, relative_widths)
	for i, unprocessed_row := range cells {
		var row StringStyler = unprocessed_row[0]
		var accumulated_length = absolute_widths[0]
		for j, cell := range unprocessed_row[1:] {
			row = row.Concat(accumulated_length-1, RuneDrawer([]rune{column_seperator}, default_seperator_style)).Concat(accumulated_length, cell)
			accumulated_length += absolute_widths[j+1]
		}
		rows[i] = row
	}
	return rows
}

func TableWithoutSeperator(total_width int, relative_widths []float64, cells [][]StringStyler) (rows []StringStyler) {
	rows = make([]StringStyler, len(cells))
	absolute_widths := getAbsoluteWidths(total_width, relative_widths)
	for i, unprocessed_row := range cells {
		var row StringStyler = EmptyDrawer()
		var accumulated_length = 0
		for j, cell := range unprocessed_row {
			next_length := accumulated_length + absolute_widths[j]
			row = row.Concat(accumulated_length, cell)
			rows[i] = row
			accumulated_length = next_length
		}
	}
	return rows
}

func generateUnderline(total_width int, relative_widths []float64) StringStyler {
	const underline_rune = '\u2500'
	const t_seperator = '\u253C'
	absolute_widths := getAbsoluteWidths(total_width, relative_widths)
	var row StringStyler = RuneRepeater(underline_rune, default_seperator_style)
	var accumulated_length = absolute_widths[0]
	for j := range absolute_widths[1:] {
		row = row.Concat(accumulated_length-1, RuneDrawer([]rune{t_seperator}, default_seperator_style)).Concat(accumulated_length, RuneRepeater(underline_rune, default_seperator_style))
		accumulated_length += absolute_widths[j+1]
	}
	return row
}
