package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	ncurses "github.com/gbin/goncurses"
	"github.com/vchimishuk/asp/config"
	"github.com/vchimishuk/chubby"
	ctime "github.com/vchimishuk/chubby/time"
)

var NcursesMu sync.Mutex

var (
	rootWnd        *ncurses.Window
	titleWnd       *TitleWindow
	statusWnd      *StatusWindow
	browserWnd     *BrowserWindow
	cmdWnd         *CommandWindow
	msgWnd         *MessageWindow
	msgWndHideTime time.Time
)

func main() {
	if err := initNcurses(); err != nil {
		printError("failed to initalize ncurses: %s", err)
		os.Exit(1)
	}
	defer destroyNcurses()

	if err := config.Load(); err != nil {
		printError("failed to load configuration file: %s", err)
		os.Exit(1)
	}

	var eventsDone <-chan any
	var err error
	client := &chubby.Chubby{}
	eventsDone, err = reconnect(client)
	if err != nil {
		printError("server connection error: %s")
		os.Exit(1)
	}

	if err := initUI(client); err != nil {
		printError("failed to initalize UI: %s", err)
		os.Exit(1)
	}

	p, err := config.LoadPath()
	if err != nil {
		p = "/"
	}
	err = browserWnd.SetPath(p)
	if err != nil {
		printError("server command failed: %s", err)
		os.Exit(1)
	}

inputLoop:
	for {
		NcursesMu.Lock()
		titleWnd.Update(map[string]string{
			"p": browserWnd.Path(),
		})
		NcursesMu.Unlock()

		var err error
		ch := rootWnd.GetChar()
		if ch != 0 {
			key := ncurses.Key(ch)
			cmd := config.Command(key)

			switch cmd {
			case config.CmdPause:
				err = client.Pause()
			case config.CmdSearch:
				// Hide message window first in case it is active.
				hideMessage(true)
				text := cmdWnd.Input("Search:")
				NcursesMu.Lock()
				browserWnd.Search(text)
				NcursesMu.Unlock()
			case config.CmdSearchNext:
				NcursesMu.Lock()
				browserWnd.SearchNext()
				NcursesMu.Unlock()
			case config.CmdSearchPrev:
				NcursesMu.Lock()
				browserWnd.SearchPrev()
				NcursesMu.Unlock()
			case config.CmdStop:
				err = client.Stop()
			case config.CmdQuit:
				break inputLoop
			default:
				NcursesMu.Lock()
				err = browserWnd.Command(cmd)
				NcursesMu.Unlock()
			}
		}
		if err != nil {
			if chubby.IsServerError(err) {
				showMessage("server error received")
			} else {
				// Network related error. Close connection,
				// connection attempt will be performed lower.
				client.Close()
			}
		}
		if !client.Connected() {
			eventsDone, err = reconnect(client)
			if err != nil {
				showMessage("server connection error")
			} else {
				hideMessage(true)
			}
		}
	}

	client.Close()

	err = config.SavePath(browserWnd.Path())
	if err != nil {
		printError("failed to save current path: %s", err)
		os.Exit(1)
	}

	wait(eventsDone, time.Second)
}

func reconnect(client *chubby.Chubby) (<-chan any, error) {
	host := os.Getenv("ASP_CHUB_HOST")
	if host == "" {
		host = config.ChubHost
	}
	ports := os.Getenv("ASP_CHUB_PORT")
	var port int
	if ports != "" {
		port, _ = strconv.Atoi(ports)
	}
	if port == 0 {
		port = config.ChubPort
	}

	err := client.Connect(host, port)
	if err != nil {
		return nil, err
	}

	done := make(chan any)
	go handleEvents(client, done)

	return done, nil
}

func initNcurses() error {
	var err error
	rootWnd, err = ncurses.Init()
	if err != nil {
		return err
	}
	if err := rootWnd.Keypad(true); err != nil {
		return err
	}
	if err := ncurses.StartColor(); err != nil {
		return err
	}
	rootWnd.Timeout(1000)

	ncurses.Echo(false)
	ncurses.CBreak(true)
	ncurses.Cursor(0)
	// TODO: nonl()
	//       tell curses not to do NL->CR/NL on output
	// TODO: raw()
	//       Ctrl-C generates keycode 0x03 instead of SIGINT

	return nil
}

func destroyNcurses() {
	ncurses.End()
}

func initUI(client *chubby.Chubby) error {
	var err error
	h, w := rootWnd.MaxYX()
	// Top panel window.
	titleWnd, err = NewTitleWindow(w, 0, 0)
	if err != nil {
		return err
	}
	// Browser window to browse VFS.
	browserWnd, err = NewBrowserWindow(client, h-2, w, 1, 0)
	if err != nil {
		return err
	}
	// Current paying status window.
	statusWnd, err = NewStatusWindow(w, h-2, 0)
	if err != nil {
		return err
	}
	// Command (e.g. search prompt) window.
	cmdWnd, err = NewCommandWindow(w, h-1, 0)
	if err != nil {
		return err
	}
	// Message window to display errors and other messages.
	// Command and Messgage windows share the same spot. Only one
	// window can be visible at time.
	msgWnd, err = NewMessageWindow(w, h-1, 0)
	if err != nil {
		return err
	}
	ncurses.UpdatePanels()
	ncurses.Update()

	return nil
}

func showMessage(format string, args ...any) {
	NcursesMu.Lock()
	defer NcursesMu.Unlock()
	msgWnd.Update(format, args...)

	delay := time.Second * 3
	msgWndHideTime = time.Now().Add(delay)

	// TODO: Prevent live goroutines growth.
	go func() {
		time.Sleep(delay)
		hideMessage(false)
	}()
}

func hideMessage(force bool) {
	NcursesMu.Lock()
	defer NcursesMu.Unlock()

	if msgWndHideTime.Unix() == 0 {
		return
	}
	if force || msgWndHideTime.Before(time.Now()) {
		msgWnd.Clear()
		ncurses.Update()
		msgWndHideTime = time.UnixMilli(0)
	}
}

func handleEvents(client *chubby.Chubby, done chan<- any) {
	var state chubby.State = chubby.StateStopped
	var track *chubby.Track
	var started int64
	var ticker *time.Ticker

	events, err := client.Events(true)
	if err != nil {
		return
	}

	st, err := client.Status()
	if err != nil {
		return
	}
	state = st.State
	track = st.Track
	started = time.Now().Unix() - int64(st.TrackPos)

loop:
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

		NcursesMu.Lock()
		browserWnd.SetSelected(data["p"])
		// TODO: Set selected for playlist window.
		statusWnd.Update(state, data)
		NcursesMu.Unlock()
		// Call to restore cursor on command window in case it is active.
		cmdWnd.Refresh()

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
				break loop
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

	done <- struct{}{}
}

func wait(ch <-chan any, delay time.Duration) {
	t := time.NewTicker(delay)
	select {
	case <-ch:
	case <-t.C:
	}
}

func printError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "asp: %s\n", fmt.Sprintf(format, args...))
}
