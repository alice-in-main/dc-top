package help_window

type Control struct {
	key     string
	meaning string
}

func MainControls() []Control {
	return []Control{
		{"'h'", "Display more controls"},
		{"'l'", "Watch container logs"},
		{"'e'", "Open shell inside the container"},
		{"'/'", "Filter containers"},
		{"'c'", "Clear filter"},
		{"'v'", "Edit docker-compose yaml"},
		{"'i'", "Inspect selected container"},
		{"'p'", "Pause selected container"},
		{"'r'", "Restart selected container"},
		{"Delete", "Remove selected container"},
		{"'s'", "Stop selected container"},
		{"'!'", "Reverse sort order"},
		{"'g'", "Go to the top of the container list"},
		{"'G'", "Go to the buttom of the container list"},
		{"F[1-5]", "Sort by column"},
	}
}

func LogControls() []Control {
	return []Control{
		{"'h'", "Display controls"},
		{"'l' or 'q'", "Exit current logs"},
		{"'/'", "Search inside logs"},
		{"'c'", "Clear search"},
		{"'n'", "Jump to next search result"},
		{"'N'", "Jump to previous search result"},
		{"Up/Down", "Browse logs"},
		{"'f'", "Resume following logs"},
	}
}
