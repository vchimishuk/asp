package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strconv"
	"sync"
	"syscall"
	"time"

	ncurses "github.com/gbin/goncurses"
	"github.com/vchimishuk/asp/config"
	"github.com/vchimishuk/chubby"
	ctime "github.com/vchimishuk/chubby/time"
	"github.com/vchimishuk/opt"
)

const Version = "0.1.0"

// TODO: Add args into usage.
var Options = []*opt.Desc{
	{"h", "host", opt.ArgString, "HOST",
		"server host name"},
	{"", "help", opt.ArgNone, "",
		"display this help"},
	{"p", "port", opt.ArgString, "PORT",
		"server port"},
	{"v", "version", opt.ArgNone,
		"", "output version information and exit"},
}

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

var (
	chub           *chubby.Chubby
	eventsDone     <-chan any
	activePath     string
	browserPath    string
	browserEntries []chubby.Entry
)

var (
	chubStatus  *chubby.Status
	chubStarted int64
)

func main() {
	opts, args, err := opt.Parse(os.Args[1:], Options)
	if err != nil {
		printErr(err)
		os.Exit(1)
	}
	if len(args) != 0 {
		printErr(errors.New("no arguments expected"))
		os.Exit(1)
	}
	if opts.Has("help") {
		printUsage(opts)
		os.Exit(0)
	}
	if opts.Has("version") {
		printVersion()
		os.Exit(0)
	}
	err = doMain(opts, args)
	if err != nil {
		printErr(err)
		os.Exit(1)
	}
}

func doMain(opts opt.Options, args []string) error {
	if err := initNcurses(); err != nil {
		return fmt.Errorf("failed to initalize ncurses: %w", err)
	}
	defer destroyNcurses()

	if err := config.Load(); err != nil {
		return fmt.Errorf("failed to load configuration file: %w", err)
	}

	if err := initUI(); err != nil {
		return fmt.Errorf("failed to initalize UI: %w", err)
	}

	host, port, err := hostPort(opts)
	if err != nil {
		return err
	}

	var eventsDone <-chan any
	chub = &chubby.Chubby{}
	eventsDone, err = reconnect(chub, host, port)
	if err != nil {
		return fmt.Errorf("server connection error: %w", err)
	}

	NcursesMu.Lock()
	browserPath, err = config.LoadPath()
	if browserPath == "" || err != nil {
		browserPath = "/"
	}

	browserEntries, err = chub.List(browserPath)
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	updateWindows()
	NcursesMu.Unlock()

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGWINCH)
	go func() {
		for {
			<-sigs

			// TODO: Asp fails on resize if command window is active.
			NcursesMu.Lock()
			destroyNcurses()
			if err := initNcurses(); err != nil {
				printErr(fmt.Errorf("failed to re-initalize ncurses: %w",
					err))
				os.Exit(1)

			}
			if err := initUI(); err != nil {
				printErr(fmt.Errorf("failed to re-initalize UI: %w", err))
				os.Exit(1)
			}
			updateWindows()
			NcursesMu.Unlock()
		}
	}()

inputLoop:
	for {
		var err error
		ch := rootWnd.GetChar()
		if ch != 0 {
			key := ncurses.Key(ch)
			cmd := config.Command(key)

			switch cmd {
			case config.CmdApply:
				entry := browserWnd.Cursor()
				if entry.IsDir() {
					err = chdir(entry.Dir().Path)
				} else {
					err = chub.Play(entry.Track().Path)
				}
			case config.CmdBack:
				err = chdir(path.Dir(browserPath))
			case config.CmdEnd:
				NcursesMu.Lock()
				browserWnd.End()
				NcursesMu.Unlock()
			case config.CmdDown:
				NcursesMu.Lock()
				browserWnd.Down()
				NcursesMu.Unlock()
			case config.CmdHome:
				NcursesMu.Lock()
				browserWnd.Home()
				NcursesMu.Unlock()
			case config.CmdPause:
				NcursesMu.Lock()
				err = chub.Pause()
				NcursesMu.Unlock()
			case config.CmdPlay:
				entry := browserWnd.Cursor()
				var p string
				if entry.IsDir() {
					p = entry.Dir().Path
				} else {
					p = entry.Track().Path
				}
				err = chub.Play(p)
			case config.CmdPageDown:
				NcursesMu.Lock()
				browserWnd.PageDown()
				NcursesMu.Unlock()
			case config.CmdPageUp:
				NcursesMu.Lock()
				browserWnd.PageUp()
				NcursesMu.Unlock()
			case config.CmdSearch:
				// Hide message window first in case it is active.
				hideMessage(true)
				text := cmdWnd.Input("Search:")
				NcursesMu.Lock()
				browserWnd.Search(text)
				NcursesMu.Unlock()
			case config.CmdShowActive:
				if activePath != "" {
					err = chdir(path.Dir(activePath))
					NcursesMu.Lock()
					browserWnd.ShowActive()
					NcursesMu.Unlock()
				}
			case config.CmdSearchNext:
				NcursesMu.Lock()
				browserWnd.SearchNext()
				NcursesMu.Unlock()
			case config.CmdSearchPrev:
				NcursesMu.Lock()
				browserWnd.SearchPrev()
				NcursesMu.Unlock()
			case config.CmdSeekBackward:
				NcursesMu.Lock()
				err = chub.Seek(ctime.New(5),
					chubby.SeekModeBackward)
				NcursesMu.Unlock()
			case config.CmdSeekForward:
				NcursesMu.Lock()
				err = chub.Seek(ctime.New(5),
					chubby.SeekModeForward)
				NcursesMu.Unlock()
			case config.CmdStop:
				err = chub.Stop()
			case config.CmdUp:
				NcursesMu.Lock()
				browserWnd.Up()
				NcursesMu.Unlock()
			case config.CmdVolumeDown:
				NcursesMu.Lock()
				err = chub.Volume(-2, chubby.VolumeModeRel)
				NcursesMu.Unlock()
			case config.CmdVolumeUp:
				NcursesMu.Lock()
				err = chub.Volume(2, chubby.VolumeModeRel)
				NcursesMu.Unlock()
			case config.CmdQuit:
				break inputLoop
			}
		}
		if err != nil {
			if chubby.IsServerError(err) {
				showMessage("server error received")
			} else {
				// Network related error. Close connection,
				// connection attempt will be performed lower.
				chub.Close()
			}
		}
		if !chub.Connected() {
			eventsDone, err = reconnect(chub, host, port)
			if err != nil {
				showMessage("server connection error")
			} else {
				hideMessage(true)
			}
		}

		hideMessage(false)
	}

	chub.Close()

	err = config.SavePath(browserPath)
	if err != nil {
		return fmt.Errorf("failed to save current path: %w", err)
	}

	wait(eventsDone, time.Second)

	return nil
}

func reconnect(chub *chubby.Chubby, host string, port int) (<-chan any, error) {
	err := chub.Connect(host, port)
	if err != nil {
		return nil, err
	}

	events, err := chub.Events(true)
	if err != nil {
		chub.Close()
		return nil, err
	}

	chubStatus, err = chub.Status()
	if err != nil {
		chub.Close()
		return nil, err
	}

	chubStarted = time.Now().Unix() - int64(chubStatus.TrackPos)
	done := make(chan any, 1)
	go handleEvents(events, done)

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
	ncurses.Update()

	return nil
}

func destroyNcurses() {
	if titleWnd != nil {
		titleWnd.Delete()
	}
	if browserWnd != nil {
		browserWnd.Delete()
	}
	if statusWnd != nil {
		statusWnd.Delete()
	}
	if cmdWnd != nil {
		cmdWnd.Delete()
	}
	if msgWnd != nil {
		msgWnd.Delete()
	}
	ncurses.End()
}

func initUI() error {
	var err error
	h, w := rootWnd.MaxYX()
	// Top panel window.
	titleWnd, err = NewTitleWindow(w, 0, 0)
	if err != nil {
		return err
	}
	rootWnd.Refresh()
	// Browser window to browse VFS.
	browserWnd, err = NewBrowserWindow(h-3, w, 1, 0)
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

	ncurses.Update()

	return nil
}

func updateWindows() {
	browserWnd.SetDir(browserPath, browserEntries)
	updateStatus()
	// TODO: Update message window.
}

func updateStatus() {
	if chubStatus == nil {
		statusWnd.Update(chubby.StateStopped, nil)
		return
	}

	data := make(map[string]string)
	data["v"] = strconv.Itoa(chubStatus.Volume)

	track := chubStatus.Track
	if track != nil {
		data["a"] = track.Artist
		data["b"] = track.Album
		data["t"] = track.Title
		data["n"] = strconv.Itoa(track.Number)
		data["l"] = track.Length.String()
		data["o"] = ctime.Time(time.Now().Unix() - chubStarted).String()
		// "r": strconv.Itoa(plist.Length),
		// "q": strconv.Itoa(se.PlistPos),
	}

	if track == nil {
		activePath = ""
		browserWnd.SetActive(activePath)
	} else if activePath != track.Path {
		activePath = track.Path
		browserWnd.SetActive(activePath)
	}
	data["p"] = browserPath

	titleWnd.Update(data)
	statusWnd.Update(chubStatus.State, data)
	// Call to restore cursor on command window in case it is active.
	cmdWnd.Refresh()
}

func chdir(p string) error {
	es, err := chub.List(p)
	if err != nil {
		return err
	}

	NcursesMu.Lock()
	browserPath = p
	browserEntries = es
	updateWindows()
	NcursesMu.Unlock()

	return nil
}

func showMessage(format string, args ...any) {
	NcursesMu.Lock()
	defer NcursesMu.Unlock()
	msgWnd.Update(format, args...)

	delay := time.Second * 3
	msgWndHideTime = time.Now().Add(delay)
}

func hideMessage(force bool) {
	if msgWndHideTime.Unix() != 0 &&
		(force || msgWndHideTime.Before(time.Now())) {
		NcursesMu.Lock()
		defer NcursesMu.Unlock()

		msgWnd.Clear()
		msgWndHideTime = time.UnixMilli(0)
	}
}

func handleEvents(events <-chan chubby.Event, done chan<- any) {
	var ticker *time.Ticker

loop:
	for {
		NcursesMu.Lock()
		updateStatus()
		NcursesMu.Unlock()

		state := chubStatus.State
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
				NcursesMu.Lock()
				chubStatus = &chubby.Status{
					State:       se.State,
					Volume:      se.Volume,
					PlaylistPos: se.PlaylistPos,
					TrackPos:    se.TrackPos,
					Playlist:    se.Playlist,
					Track:       se.Track,
				}

				chubStarted = time.Now().Unix() -
					int64(se.TrackPos)
				NcursesMu.Unlock()
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

func hostPort(opts opt.Options) (string, int, error) {
	var host string = config.ChubHost
	var port string = strconv.Itoa(config.ChubPort)
	var err error

	if h := os.Getenv("ASP_HOST"); h != "" {
		host = h
	}
	if h, ok := opts.String("host"); ok {
		host = h
	}

	if p := os.Getenv("ASP_PORT"); p != "" {
		port = p
	}
	if p, ok := opts.String("port"); ok {
		port = p
	}

	iport, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port %s", port)
	}

	return host, iport, err
}

func printErr(err error) {
	fmt.Fprintf(os.Stderr, "asp: %s\n", err.Error())
}

func printUsage(opts opt.Options) {
	fmt.Println("usage: asp [OPTION]...")
	fmt.Println()
	fmt.Print(opt.Usage(Options))
}

func printVersion() {
	fmt.Printf("%s\n", Version)
}
