// Copyright 2017 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// This file is part of asp.
//
// Chub is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Chub is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Chub. If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	ncurses "github.com/gbin/goncurses"
	"github.com/vchimishuk/asp/config/scanner"
)

type Section string

const (
	SectionGlobal   Section = "global"
	SectionPlaylist Section = "playlist"
	SectionBrowser  Section = "vfs"
)

// TODO: Rename to Cmd.
type Command string

const (
	CommandBack       Command = "back"
	CommandEnd        Command = "end"
	CommandEnter      Command = "enter"
	CommandHome       Command = "home"
	CommandNext       Command = "next"
	CommandNoop       Command = "noop"
	CommandPageDown   Command = "page-down"
	CommandPageUp     Command = "page-up"
	CommandPause      Command = "pause"
	CommandPlay       Command = "play"
	CommandPrev       Command = "prev"
	CommandQuit       Command = "quit"
	CommandSearch     Command = "search"
	CommandSearchNext Command = "search-next"
	CommandSearchPrev Command = "search-prev"
	CommandSelected   Command = "selected"
	CommandStop       Command = "stop"
)

type Keys map[Section]map[Command][]ncurses.Key

var defaultKeys = Keys{
	SectionGlobal: {
		CommandEnd:  {ncurses.KEY_END, ctrlKey('e')},
		CommandHome: {ncurses.KEY_HOME, ctrlKey('a')},
		CommandNext: {ncurses.KEY_DOWN, ncurses.Key('j'),
			ctrlKey('n')},
		CommandPageDown: {ncurses.KEY_PAGEDOWN, ctrlKey('v'),
			ctrlKey('d')},
		CommandPageUp: {ncurses.KEY_PAGEUP, ctrlKey('b')},
		CommandPause:  {ncurses.Key(' '), ncurses.Key('p')},
		CommandPrev: {ncurses.KEY_UP, ncurses.Key('k'),
			ctrlKey('p')},
		CommandQuit:       {ncurses.Key('q')},
		CommandSearch:     {ncurses.Key('/')},
		CommandSearchNext: {ncurses.Key('n')},
		CommandSearchPrev: {ncurses.Key('N')},
		CommandStop:       {ncurses.Key('s')},
	},
	SectionPlaylist: {},
	SectionBrowser: {
		CommandBack: {ncurses.KEY_BACKSPACE, ncurses.Key('h'),
			ctrlKey('h')},
		CommandEnter:    {ncurses.KEY_RETURN},
		CommandPlay:     {ncurses.Key('x')},
		CommandSelected: {ncurses.Key('G')},
	},
}

func ParseKeysFile(name string) (Keys, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ParseKeys(f)
}

func ParseKeys(r io.Reader) (Keys, error) {
	scnr := scanner.New(r)
	keys := make(Keys)

	for scnr.Scan() {
		pts := strings.SplitN(scnr.Key(), ".", 2)
		s := SectionGlobal
		var c Command
		if len(pts) == 2 {
			switch pts[0] {
			case string(SectionPlaylist):
				s = SectionPlaylist
			case string(SectionBrowser):
				s = SectionBrowser
			default:
				err := fmt.Errorf("invalid command at line %d",
					scnr.Line())
				return nil, err
			}
			c = parseCommand(pts[1])
		} else {
			c = parseCommand(pts[0])
		}
		if c == "" {
			err := fmt.Errorf("invalid command at line %d",
				scnr.Line())
			return nil, err
		}

		for _, k := range strings.Split(scnr.Val(), ",") {
			k = strings.Trim(k, " ")
			i, err := strconv.Atoi(k)
			if err != nil {
				err := fmt.Errorf("invalid key %s at line %d",
					k, scnr.Line())
				return nil, err
			}
			if _, ok := keys[s]; !ok {
				keys[s] = make(map[Command][]ncurses.Key)
			}
			keys[s][c] = append(keys[s][c], ncurses.Key(i))
		}
	}
	if err := scnr.Err(); err != nil {
		return nil, err
	}

	for s, m := range defaultKeys {
		for c, k := range m {
			if _, ok := keys[s]; !ok {
				keys[s] = make(map[Command][]ncurses.Key)
			}
			if _, ok := keys[s][c]; !ok {
				keys[s][c] = k
			}
		}
	}

	return keys, nil
}

func parseCommand(s string) Command {
	switch s {
	case string(CommandEnter):
		return CommandEnter
	case string(CommandNext):
		return CommandNext
	case string(CommandPrev):
		return CommandPrev
	case string(CommandQuit):
		return CommandQuit
	default:
		return ""
	}
}

func ctrlKey(r rune) ncurses.Key {
	return ncurses.Key(r) & 0x1F

}
