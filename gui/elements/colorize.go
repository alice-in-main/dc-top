package elements

type FColor string
type BColor string

const (
	reset                = "\033[0m"
	Red           FColor = "\033[31m"
	Green         FColor = "\033[32m"
	Yellow        FColor = "\033[33m"
	Blue          FColor = "\033[34m"
	Purple        FColor = "\033[35m"
	Cyan          FColor = "\033[36m"
	Gray          FColor = "\033[37m"
	White         FColor = "\033[97m"
	B_Black       BColor = "\033[40m"
	B_Red         BColor = "\033[41m"
	B_Green       BColor = "\033[42m"
	B_Orange      BColor = "\033[43m"
	B_Blue        BColor = "\033[44m"
	B_Magenta     BColor = "\033[45m"
	B_Cyan        BColor = "\033[46m"
	B_LightGray   BColor = "\033[47m"
	B_DarkGray    BColor = "\033[100m"
	B_LightRed    BColor = "\033[101m"
	B_LightGreen  BColor = "\033[102m"
	B_Yellow      BColor = "\033[103m"
	B_LightBlue   BColor = "\033[104m"
	B_LightPurple BColor = "\033[105m"
	B_Teal        BColor = "\033[106m"
	B_White       BColor = "\033[107m"
)

func Foreground(str string, color FColor) string {
	return string(color) + str + string(reset)
}

func Background(str string, color BColor) string {
	return string(color) + str + string(reset)
}

func Colorize(str string, f_color FColor, b_color BColor) string {
	return string(b_color) + string(f_color) + str + string(reset)
}
