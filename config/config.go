package config

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	ncurses "github.com/gbin/goncurses"
	"github.com/vchimishuk/config"
)

// TODO: Add to github.com/vchimishuk/config custom value validator?
var spec = &config.Spec{
	Blocks: []*config.BlockSpec{
		&config.BlockSpec{
			Name:   "colors",
			Strict: true,
			Properties: []*config.PropertySpec{
				&config.PropertySpec{
					Type: config.TypeString,
					Name: string("cursor"),
				},
				&config.PropertySpec{
					Type: config.TypeString,
					Name: string("cursor-selected"),
				},
				&config.PropertySpec{
					Type: config.TypeString,
					Name: string("list"),
				},
				&config.PropertySpec{
					Type: config.TypeString,
					Name: string("list-selected"),
				},
				&config.PropertySpec{
					Type: config.TypeString,
					Name: string("normal"),
				},
				&config.PropertySpec{
					Type: config.TypeString,
					Name: string("status"),
				},
				&config.PropertySpec{
					Type: config.TypeString,
					Name: string("title"),
				},
			},
		},
		&config.BlockSpec{
			Name:   "keys",
			Strict: true,
			Properties: []*config.PropertySpec{
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdApply),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdBack),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdEnd),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdHome),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdNext),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdPageDown),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdPageUp),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdPause),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdPlay),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdPrev),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdQuit),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdSearch),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdSearchNext),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdSearchPrev),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdSelected),
				},
				&config.PropertySpec{
					Type: config.TypeStringList,
					Name: string(CmdStop),
				},
			},
		},
	},
}

var colorNames = map[string]int16{
	"black":   ncurses.C_BLACK,
	"blue":    ncurses.C_BLUE,
	"cyan":    ncurses.C_CYAN,
	"green":   ncurses.C_GREEN,
	"magenta": ncurses.C_MAGENTA,
	"red":     ncurses.C_RED,
	"white":   ncurses.C_WHITE,
	"yellow":  ncurses.C_YELLOW,
}

var (
	ColorCursor         ncurses.Char
	ColorCursorSelected ncurses.Char
	ColorList           ncurses.Char
	ColorListSelected   ncurses.Char
	ColorNormal         ncurses.Char
	ColorStatus         ncurses.Char
	ColorTitle          ncurses.Char
)

type Cmd string

const (
	CmdApply      Cmd = "apply"
	CmdBack       Cmd = "back"
	CmdEnd        Cmd = "end"
	CmdHome       Cmd = "home"
	CmdNext       Cmd = "next"
	CmdNoop       Cmd = "noop"
	CmdPageDown   Cmd = "page-down"
	CmdPageUp     Cmd = "page-up"
	CmdPause      Cmd = "pause"
	CmdPlay       Cmd = "play"
	CmdPrev       Cmd = "prev"
	CmdQuit       Cmd = "quit"
	CmdSearch     Cmd = "search"
	CmdSearchNext Cmd = "search-next"
	CmdSearchPrev Cmd = "search-prev"
	CmdSelected   Cmd = "selected"
	CmdStop       Cmd = "stop"
)

var defKeymap = map[Cmd][]ncurses.Key{
	CmdApply: []ncurses.Key{
		ncurses.KEY_RETURN,
		ncurses.Key('l'),
	},
	CmdBack: []ncurses.Key{
		ncurses.KEY_BACKSPACE,
		ncurses.Key('h'),
		ctrlKey('h'),
	},
	CmdEnd: []ncurses.Key{
		ncurses.KEY_END,
		ctrlKey('e'),
	},
	CmdHome: []ncurses.Key{
		ncurses.KEY_HOME,
		ctrlKey('a'),
	},
	CmdNext: []ncurses.Key{
		ncurses.KEY_DOWN,
		ncurses.Key('j'),
		ctrlKey('n'),
	},
	CmdPageDown: []ncurses.Key{
		ncurses.KEY_PAGEDOWN,
		ctrlKey('v'),
		ctrlKey('d'),
	},
	CmdPageUp: []ncurses.Key{
		ncurses.KEY_PAGEUP,
		ctrlKey('b'),
	},
	CmdPause: []ncurses.Key{
		ncurses.Key(' '),
		ncurses.Key('p'),
	},
	CmdPlay: []ncurses.Key{
		ncurses.Key('x'),
	},
	CmdPrev: []ncurses.Key{
		ncurses.KEY_UP,
		ncurses.Key('k'),
		ctrlKey('p'),
	},
	CmdQuit: []ncurses.Key{
		ncurses.Key('q'),
	},
	CmdSearch: []ncurses.Key{
		ncurses.Key('/'),
	},
	CmdSearchNext: []ncurses.Key{
		ncurses.Key('n'),
	},
	CmdSearchPrev: []ncurses.Key{
		ncurses.Key('N'),
	},
	CmdSelected: []ncurses.Key{
		ncurses.Key('G'),
	},
	CmdStop: []ncurses.Key{
		ncurses.Key('s'),
	},
}

var keymap map[ncurses.Key]Cmd = make(map[ncurses.Key]Cmd)

func Load() error {
	var cfg *config.Config
	cd, err := configDir()
	if err != nil {
		return err
	}
	file := filepath.Join(cd, "asp.conf")
	d, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = &config.Config{}
		} else {
			return err
		}
	} else {
		cfg, err = config.Parse(spec, string(d))
		if err != nil {
			return err
		}

	}

	err = initColors(cfg)
	if err != nil {
		return err
	}
	err = initKeymap(cfg)
	if err != nil {
		return err
	}

	return nil
}

func Command(key ncurses.Key) Cmd {
	c, ok := keymap[key]
	if ok {
		return c
	}

	return CmdNoop
}

func initKeymap(cfg *config.Config) error {
	block := &config.Block{}
	if cfg.Has("keys") {
		block = cfg.Block("keys")
	}

	for cmd, keys := range defKeymap {
		if block.Has(string(cmd)) {
			for _, s := range block.StringList(string(cmd)) {
				k, err := parseKey(s)
				if err != nil {
					return err
				}
				keymap[k] = cmd
			}
		} else {
			for _, k := range keys {
				keymap[k] = cmd
			}
		}
	}

	return nil
}

func initColors(cfg *config.Config) error {
	colors := []struct {
		ID   int
		Var  *ncurses.Char
		Name string
		Def  string
	}{
		{1, &ColorCursor, "cursor", "black:cyan"},
		{2, &ColorCursorSelected, "cursor-selected", "red:cyan"},
		{3, &ColorList, "list", "white:black"},
		{4, &ColorListSelected, "list-selected", "red:black"},
		{5, &ColorNormal, "normal", "white:black"},
		{6, &ColorStatus, "status", "black:blue"},
		{7, &ColorTitle, "title", "black:blue"},
	}

	block := &config.Block{}
	if cfg.Has("colors") {
		block = cfg.Block("colors")
	}

	for _, c := range colors {
		// TODO: Config value validation.
		cl, err := initPair(c.ID, block.StringOr(c.Name, c.Def))
		if err != nil {
			return err
		}
		*c.Var = cl
	}

	return nil
}

func initPair(id int, s string) (ncurses.Char, error) {
	pts := strings.SplitN(s, ":", 2)
	if len(pts) != 2 {
		return 0, fmt.Errorf("invalid color pair: %s", s)
	}
	fg, okf := colorNames[pts[0]]
	bg, okb := colorNames[pts[1]]
	if !okf || !okb {
		return 0, fmt.Errorf("invalid color pair: %s", s)
	}

	err := ncurses.InitPair(int16(id), fg, bg)
	if err != nil {
		return 0, err
	}
	c := ncurses.ColorPair(int16(id))

	return c, nil
}

func parseKey(s string) (ncurses.Key, error) {
	if len(s) == 1 {
		return ncurses.Key(rune(s[0])), nil
	} else if len(s) == 2 && s[0] == '^' {
		return ctrlKey(rune(s[1])), nil
	} else if len(s) > 1 && s[0] == '#' {
		i, err := strconv.Atoi(s[1:])
		return ncurses.Key(i), err
	} else {
		return 0, fmt.Errorf("invalid key: %s", s)
	}
}

func ctrlKey(r rune) ncurses.Key {
	return ncurses.Key(r) & 0x1F

}

func configDir() (string, error) {
	ch := os.Getenv("XDG_CONFIG_HOME")
	if ch != "" {
		return ch, nil
	}

	u, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(u.HomeDir, ".config/asp"), nil
}
