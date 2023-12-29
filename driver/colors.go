package main

import (
	"fmt"
)

type Color string

const (
	ColorBlack      Color = "\033[0;30m"
	ColorRed        Color = "\033;31m"
	ColorGreen      Color = "\033[0;32m"
	ColorYellow     Color = "\033[0;33m"
	ColorBlue       Color = "\033[0;34m"
	ColorPurple     Color = "\033[0;35m"
	ColorCyan       Color = "\033[0;36m"
	ColorWhite      Color = "\033[0;37m"
	ColorBlackBold  Color = "\033[1;30m"
	ColorRedBold    Color = "\033[1;31m"
	ColorGreenBold  Color = "\033[1;32m"
	ColorYellowBold Color = "\033[1;33m"
	ColorBlueBold   Color = "\033[1;34m"
	ColorPurpleBold Color = "\033[1;35m"
	ColorCyanBold   Color = "\033[1;36m"
	ColorWhiteBold  Color = "\033[1;37m"
	ColorBlackHI    Color = "\033[0;90m"
	ColorRedHI      Color = "\033[0;91m"
	ColorGreenHI    Color = "\033[0;92m"
	ColorYellowHI   Color = "\033[0;93m"
	ColorBlueHI     Color = "\033[0;94m"
	ColorPurpleHI   Color = "\033[0;95m"
	ColorCyanHI     Color = "\033[0;96m"
	ColorWhiteHI    Color = "\033[0;97m"
	ColorBgBlack    Color = "\033[40m"
	ColorBgRed      Color = "\033[41m"
	ColorBgGreen    Color = "\033[42m"
	ColorBgYellow   Color = "\033[43m"
	ColorBgBlue     Color = "\033[44m"
	ColorBgPurple   Color = "\033[45m"
	ColorBgCyan     Color = "\033[46m"
	ColorBgWhite    Color = "\033[47m"
	ColorBgBlackHI  Color = "\033[0;100m"
	ColorBgRedHI    Color = "\033[0;101m"
	ColorBgGreenHI  Color = "\033[0;102m"
	ColorBgYellowHI Color = "\033[0;103m"
	ColorBgBlueHI   Color = "\033[0;104m"
	ColorBgPurpleHI Color = "\033[0;105m"
	ColorBgCyanHI   Color = "\033[0;106m"
	ColorBgWhiteHI  Color = "\033[0;107m"
	ColorReset      Color = "\033[0m"
)

func PrtCol(msg string, color Color) {
	fmt.Printf("%s%s%s", color, msg, ColorReset)
}
