// TODO:
// 1. Move title window near to status window.
// 2. Title and status window use the same set of parameters.
// 3. Update title and status windows every second if playing.
// 4. Update title and status windows after command processing.
// 5. Available config from any window (passed as a parameter).

package main

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strconv"
	gosync "sync"
	"time"

	ncurses "github.com/gbin/goncurses"
	"github.com/vchimishuk/asp/config"
	"github.com/vchimishuk/chubby"
	ctime "github.com/vchimishuk/chubby/time"
)

var stdScr *ncurses.Window
var titleWnd *PanelWindow
var statusWnd *StatusWindow
var cmdWnd *CommandWindow
var browserWnd *BrowserWindow
var ncursesMu gosync.Mutex

func main() {
	var err error

	// Display error on exit if any.
	defer func() {
		if err != nil {
			// TODO: Display error.
			fmt.Printf("asp: %s\n", err)
		}
	}()

	u, err := user.Current()
	if err != nil {
		return
	}
	keys, err := config.ParseKeysFile(filepath.Join(u.HomeDir,
		".config/asp/keys.conf"))
	if err != nil {
		return
	}
	colors, err := config.ParseColorsFile(filepath.Join(u.HomeDir,
		".config/asp/colors.conf"))
	if err != nil {
		return
	}

	client := &chubby.Chubby{}
	// TODO: host:port
	if err = client.Connect("localhost", 5115); err != nil {
		return
	}
	defer client.Close()

	if err = initUI(colors, client); err != nil {
		return
	}
	defer releaseUI()

	// Listen for server notifications.
	// stopNoticeLoop := sync.NewCond(1)
	go handleEvents(client)
	// defer stopNoticeLoop.Signal()

	for {
		ncursesMu.Lock()
		titleWnd.SetText(" " + browserWnd.Path())
		ncursesMu.Unlock()

		ch := stdScr.GetChar()
		key := ncurses.Key(ch)
		cmd := keyToCmd(keys, key)

		ncursesMu.Lock()
		err := browserWnd.Command(cmd)
		ncursesMu.Unlock()
		if err != nil {
			// TODO: Display it in status
		}
		switch cmd {
		case config.CommandPause:
			// TODO: Error handling.
			client.Pause()
		case config.CommandStop:
			// TODO: Error handling.
			client.Stop()
		case config.CommandQuit:
			// TODO: Close client.
			return
		}
	}
}

func initUI(colors config.Colors, client *chubby.Chubby) error {
	var err error
	stdScr, err = ncurses.Init()
	if err != nil {
		return err
	}
	if err := stdScr.Keypad(true); err != nil {
		return err
	}
	if err := ncurses.StartColor(); err != nil {
		return err
	}
	if err := initColors(colors); err != nil {
		return err
	}

	ncurses.Echo(false)
	ncurses.CBreak(true)
	ncurses.Cursor(0)
	// nonl()
	// raw()

	h, w := stdScr.MaxYX()
	// Top panel window.
	titleWnd, err = NewPanelWindow(w, 0, 0)
	if err != nil {
		return err
	}
	titleWnd.SetBackground(config.ColorPair(config.ColorTitle))

	// Bottom panel window -- window above command one.
	statusWnd, err = NewStatusWindow(w, h-2, 0)
	if err != nil {
		return err
	}

	// The lowes command window.
	cmdWnd, err = NewCommandWindow(w, h-1, 0)
	if err != nil {
		return err
	}

	// Browser window to browse VFS.
	browserWnd, err = NewBrowserWindow(client, cmdWnd.Input,
		h-2, w, 1, 0)
	if err != nil {
		return err
	}
	browserWnd.SetPath("/") // TODO: Set las dir.

	ncurses.UpdatePanels()
	ncurses.Update()

	return nil
}

func releaseUI() {
	ncurses.End()
}

func initColors(colors config.Colors) error {
	for id, c := range colors {
		err := ncurses.InitPair(int16(id), int16(c.Fg), int16(c.Bg))
		if err != nil {
			return err
		}
	}

	return nil
}

func keyToCmd(keys config.Keys, key ncurses.Key) config.Command {
	sect, ok := keys[config.SectionBrowser]
	if !ok {
		return config.CommandNoop
	}
	for cmd, ks := range sect {
		for _, k := range ks {
			if k == key {
				return cmd
			}
		}
	}
	for cmd, ks := range keys[config.SectionGlobal] {
		for _, k := range ks {
			if k == key {
				return cmd
			}
		}
	}

	return config.CommandNoop
}

func handleEvents(client *chubby.Chubby) {
	var state chubby.State = chubby.StateStopped
	var track *chubby.Track
	var started int64
	var ticker *time.Ticker

	events, err := client.Events(true)
	if err != nil {
		// TODO: Handle error -- retry in loop.
		panic(err)
	}

	st, err := client.Status()
	if err != nil {
		// TODO:
		panic(err)
	}
	state = st.State
	track = st.Track
	started = time.Now().Unix() - int64(st.TrackPos)

	for {
		data := make(map[string]string)
		if track != nil {
			data["p"] = track.Path
			data["a"] = track.Artist
			data["b"] = track.Album
			data["t"] = track.Title
			data["n"] = strconv.Itoa(track.Number)
			data["l"] = track.Length.String()
			data["o"] = ctime.Time(time.Now().Unix() -
				started).String()
			// "r": strconv.Itoa(plist.Length),
			// "q": strconv.Itoa(se.PlistPos),
		}

		ncursesMu.Lock()
		browserWnd.SetSelected(data["p"])
		// TODO: Set selected for playlist window.
		statusWnd.Update(state, data)
		ncursesMu.Unlock()

		if state == chubby.StatePlaying && ticker == nil {
			ticker = time.NewTicker(time.Millisecond * 900)
		} else if state != chubby.StatePlaying && ticker != nil {
			ticker.Stop()
			ticker = nil
		}

		var tickerCh <-chan time.Time
		if ticker != nil {
			tickerCh = ticker.C
		}

		select {
		case e, ok := <-events:
			if !ok {
				break
			}
			if se, ok := e.(*chubby.StatusEvent); ok {
				state = se.State
				// plist = se.Playlist
				track = se.Track
				started = time.Now().Unix() - int64(se.TrackPos)
			}
		case <-tickerCh:
		}
	}

	if ticker != nil {
		ticker.Stop()
	}
}
