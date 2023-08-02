package config

import (
	"io"

	ncurses "github.com/gbin/goncurses"
	"github.com/vchimishuk/asp/color"
)

// type Color struct {
// 	ID int
// 	Fg color.Color
// 	Bg color.Color
// }

//type ColorItem string
type Color int

// const (
// 	ColorItemCursor         ColorItem = "cursor-normal"
// 	ColorItemCursorSelected ColorItem = "cursor-selected"
// 	ColorItemListNormal     ColorItem = "list-normal"
// 	ColorItemListSelected   ColorItem = "list-selected"
// 	ColorItemNormal         ColorItem = "normal"
// 	ColorItemStatus         ColorItem = "status"
// 	ColorItemTitle          ColorItem = "title"
// )

const (
	ColorCursor Color = iota + 1
	ColorCursorSelected
	ColorList
	ColorListSelected
	ColorNormal
	ColorStatus
	ColorTitle
)

type Colors map[Color]color.Pair

// var defaultColors = Colors{
// 	ColorItemCursor:         color.Pair{color.White, color.Black},
// 	ColorItemCursorSelected: color.Pair{color.White, color.Black},
// 	ColorItemListNormal:     color.Pair{color.White, color.Black},
// 	ColorItemListSelected:   color.Pair{color.White, color.Black},
// 	ColorItemNormal:         color.Pair{color.White, color.Black},
// 	ColorItemStatus:         color.Pair{color.White, color.Black},
// 	ColorItemTitle:          color.Pair{color.White, color.Black},
// }
var defaultColors = map[string]struct {
	ID   Color
	Pair color.Pair
}{
	"cursor": {ColorCursor,
		color.Pair{color.Black, color.Cyan}},
	"cursor-selected": {ColorCursorSelected,
		color.Pair{color.Red, color.Cyan}},
	"list": {ColorList,
		color.Pair{color.White, color.Black}},
	"list-selected": {ColorListSelected,
		color.Pair{color.Red, color.Black}},
	"normal": {ColorNormal,
		color.Pair{color.White, color.Black}},
	"status": {ColorStatus,
		color.Pair{color.White, color.Black}},
	"title": {ColorTitle,
		color.Pair{color.Black, color.Blue}},
}

func ParseColorsFile(name string) (Colors, error) {
	// TODO:
	return ParseColors(nil)
}

func ParseColors(r io.Reader) (Colors, error) {
	// TODO:
	c := make(Colors)

	for _, s := range defaultColors {
		c[s.ID] = s.Pair
	}

	return c, nil
}

func ColorPair(c Color) ncurses.Char {
	return ncurses.ColorPair(int16(c))
}
