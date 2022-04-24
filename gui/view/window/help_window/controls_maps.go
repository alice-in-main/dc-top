package help_window

type Control struct {
	key     string
	meaning string
}

func MainControls() []Control {
	return []Control{
		{"'h'", "Display more controls"},
		{"'l'", "Watch container logs"},
		{"'e'", "Open shell inside selected container"},
		{"'/'", "Filter containers"},
		{"'c'", "Clear filter"},
		{"'v'", "Edit docker-compose yaml"},
		{"'i'", "Inspect selected container"},
		{"Ctrl+P", "Pause selected container"},
		{"Ctrl+R", "Restart selected container"},
		{"Delete", "Remove selected container"},
		{"Ctrl+S", "Stop selected container"},
		{"Ctrl+U", "Update docker compose"},
		{"Ctrl+W", "Restart docker compose"},
		{"Ctrl+D", "Remove (down) docker compose"},
		{"'!'", "Reverse sort order"},
		{"'g'", "Go to the top of the container list/inspect info"},
		{"'G'", "Go to the buttom of the container list"},
		{"Up/Down", "Browse containers/Scroll inspect info"},
		{"F[1-5]", "Sort by column"},
	}
}

func LogControls() []Control {
	return []Control{
		{"'h'", "Display controls"},
		{"'l'/'q'", "Exit current logs"},
		{"'/'", "Search inside logs"},
		{"'c'", "Clear search"},
		{"'n'", "Jump to next search result"},
		{"'N'", "Jump to previous search result"},
		{"Up/Down", "Browse logs"},
		{"'f'", "Resume following logs"},
	}
}

func EdittorControls() []Control {
	return []Control{
		{"Ctrl+H", "Display controls"},
		{"'h'", "Exit controls"},
		{"Ctrl+Q", "Exit edittor without saving"},
		{"Ctrl+S", "Exit edittor and save"},
		{"Ctrl+Z", "Undo"},
		{"Ctrl+Alt+Z", "Redo"},
		{"Ctrl+F", "Search"},
		{"Ctrl+A", "Clear search"},
		{"Ctrl+D", "Delete line"},
		{"'n' / 'N'", "Next search result / Previous search result"},
	}
}
