package edittor_window

func addStrToLine(str string, total_content *[]string, line_index int, col_index int) {
	line := (*total_content)[line_index]
	(*total_content)[line_index] = line[:col_index] + str + line[col_index:]
}

func removeStrFromLine(strlen int, total_content *[]string, line_index int, col_index int) {
	(*total_content)[line_index] =
		(*total_content)[line_index][:col_index] + (*total_content)[line_index][strlen+col_index:]
}

func breakLine(total_content *[]string, line_index int, col_index int) {
	line := (*total_content)[line_index]
	first_half_line := line[:col_index]
	second_half_line := line[col_index:]
	first_half := (*total_content)[:line_index+1]
	second_half := append([]string{""}, (*total_content)[line_index+1:]...)
	(*total_content) = append(first_half, second_half...)
	(*total_content)[line_index] = first_half_line
	(*total_content)[line_index+1] = second_half_line
}

func collapseLine(total_content *[]string, line_index int) {
	(*total_content)[line_index] = (*total_content)[line_index] + (*total_content)[line_index+1]
	(*total_content) = append((*total_content)[:line_index+1], (*total_content)[line_index+2:]...)
}

func removeLine(total_content *[]string, line_index int) {
	*total_content = append((*total_content)[:line_index], (*total_content)[line_index+1:]...)
}

func addLine(str string, total_content *[]string, line_index int) {
	first_half := (*total_content)[:line_index]
	second_half := append([]string{str}, (*total_content)[line_index:]...)
	*total_content = append(first_half, second_half...)
}
