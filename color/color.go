package color

import ncurses "github.com/gbin/goncurses"

type Color int

const (
	Black   Color = Color(ncurses.C_BLACK)
	Blue    Color = Color(ncurses.C_BLUE)
	Cyan    Color = Color(ncurses.C_CYAN)
	Green   Color = Color(ncurses.C_GREEN)
	Magenta Color = Color(ncurses.C_MAGENTA)
	Red     Color = Color(ncurses.C_RED)
	White   Color = Color(ncurses.C_WHITE)
	Yellow  Color = Color(ncurses.C_YELLOW)
)

type Pair struct {
	Fg Color
	Bg Color
}
