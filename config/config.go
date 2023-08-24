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

var spec = &config.Spec{
	Blocks: []*config.BlockSpec{
		&config.BlockSpec{
			Name:   "colors",
			Strict: true,
			Properties: []*config.PropertySpec{
				&config.PropertySpec{
					Type:   config.TypeString,
					Name:   string("cursor"),
					Parser: parseColor,
				},
				&config.PropertySpec{
					Type:   config.TypeString,
					Name:   string("cursor-selected"),
					Parser: parseColor,
				},
				&config.PropertySpec{
					Type:   config.TypeString,
					Name:   string("list"),
					Parser: parseColor,
				},
				&config.PropertySpec{
					Type:   config.TypeString,
					Name:   string("list-selected"),
					Parser: parseColor,
				},
				&config.PropertySpec{
					Type:   config.TypeString,
					Name:   string("normal"),
					Parser: parseColor,
				},
				&config.PropertySpec{
					Type:   config.TypeString,
					Name:   string("status"),
					Parser: parseColor,
				},
				&config.PropertySpec{
					Type:   config.TypeString,
					Name:   string("title"),
					Parser: parseColor,
				},
			},
		},
		&config.BlockSpec{
			Name:   "keys",
			Strict: true,
			Properties: []*config.PropertySpec{
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdApply),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdBack),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdEnd),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdHome),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdNext),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdPageDown),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdPageUp),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdPause),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdPlay),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdPrev),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdQuit),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdSearch),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdSearchNext),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdSearchPrev),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdSelected),
					Parser: parseKey,
				},
				&config.PropertySpec{
					Type:   config.TypeStringList,
					Name:   string(CmdStop),
					Parser: parseKey,
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
		ks := block.AnyOr(string(cmd), keys).([]ncurses.Key)
		for _, k := range ks {
			keymap[k] = cmd
		}
	}

	return nil
}

func initColors(cfg *config.Config) error {
	colors := []struct {
		ID   int16
		Var  *ncurses.Char
		Name string
		Def  []int16
	}{
		{1, &ColorCursor, "cursor",
			[]int16{colorNames["black"],
				colorNames["cyan"]}},
		{2, &ColorCursorSelected, "cursor-selected",
			[]int16{colorNames["red"],
				colorNames["cyan"]}},
		{3, &ColorList, "list",
			[]int16{colorNames["white"],
				colorNames["black"]}},
		{4, &ColorListSelected, "list-selected",
			[]int16{colorNames["red"],
				colorNames["black"]}},
		{5, &ColorNormal, "normal",
			[]int16{colorNames["white"],
				colorNames["black"]}},
		{6, &ColorStatus, "status",
			[]int16{colorNames["black"],
				colorNames["blue"]}},
		{7, &ColorTitle, "title",
			[]int16{colorNames["black"],
				colorNames["blue"]}},
	}

	block := &config.Block{}
	if cfg.Has("colors") {
		block = cfg.Block("colors")
	}

	for _, c := range colors {
		v := block.AnyOr(c.Name, c.Def)
		fg := v.([]int16)[0]
		bg := v.([]int16)[1]
		err := ncurses.InitPair(c.ID, fg, bg)
		if err != nil {
			return err
		}
		*c.Var = ncurses.ColorPair(c.ID)
	}

	return nil
}

func parseColor(v any) (any, error) {
	pts := strings.SplitN(v.(string), ":", 2)
	if len(pts) != 2 {
		return 0, errors.New("invalid color pair")
	}
	fg, ok := colorNames[pts[0]]
	if !ok {
		return 0, fmt.Errorf("invalid color: %s", pts[0])
	}
	bg, ok := colorNames[pts[1]]
	if !ok {
		return 0, fmt.Errorf("invalid color: %s", pts[1])
	}

	return []int16{fg, bg}, nil
}

func parseKey(v any) (any, error) {
	res := []ncurses.Key{}

	for _, s := range v.([]string) {
		if len(s) == 1 {
			res = append(res, ncurses.Key(rune(s[0])))
		} else if len(s) == 2 && s[0] == '^' {
			res = append(res, ctrlKey(rune(s[1])))
		} else if len(s) > 1 && s[0] == '#' {
			i, err := strconv.Atoi(s[1:])
			if err != nil {
				return nil, err
			}
			res = append(res, ncurses.Key(i))
		} else {
			return nil, fmt.Errorf("invalid key: %s", s)
		}
	}

	return res, nil
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
