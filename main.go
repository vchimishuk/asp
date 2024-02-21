package main

import (
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
	msgWndShowCond *sync.Cond
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
	err := doMain()
	if err != nil {
		printErr(err)
		os.Exit(1)
	}
}

func doMain() error {
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

	var eventsDone <-chan any
	var err error
	chub = &chubby.Chubby{}
	eventsDone, err = reconnect(chub)
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
		// TODO: err: short write
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
		NcursesMu.Lock()
		// TODO: Move into updateWindows()
		// TODO:  Title window should be updated using some global
		//        state just like browser window is.
		titleWnd.Update(map[string]string{
			"p": browserPath,
		})
		NcursesMu.Unlock()

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
				err = chdir(path.Dir(activePath))
				NcursesMu.Lock()
				browserWnd.ShowActive()
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
				err = chub.Stop()
			case config.CmdUp:
				NcursesMu.Lock()
				browserWnd.Up()
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
			eventsDone, err = reconnect(chub)
			if err != nil {
				showMessage("server connection error")
			} else {
				hideMessage(true)
			}
		}
	}

	chub.Close()

	err = config.SavePath(browserPath)
	if err != nil {
		return fmt.Errorf("failed to save current path: %w", err)
	}

	wait(eventsDone, time.Second)

	return nil
}

func reconnect(chub *chubby.Chubby) (<-chan any, error) {
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

	err := chub.Connect(host, port)
	if err != nil {
		return nil, err
	}

	done := make(chan any, 1)
	go handleEvents(chub, done)

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
	titleWnd.Delete()
	browserWnd.Delete()
	statusWnd.Delete()
	cmdWnd.Delete()
	msgWnd.Delete()
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

	// TODO: Use input loop to do it.
	go delayedMessageHider()

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
	track := chubStatus.Track
	if track != nil {
		data["p"] = track.Path
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

	// // TODO: Title window shoul accept the same params as status window.
	// titleWnd.Update(map[string]string{
	// 	"p": activePath,
	// })
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
	msgWndShowCond.Signal()
}

func hideMessage(force bool) {
	NcursesMu.Lock()
	defer NcursesMu.Unlock()

	if msgWndHideTime.Unix() != 0 &&
		(force || msgWndHideTime.Before(time.Now())) {

		doHideMessage()
	}
}

func doHideMessage() {
	if msgWnd != nil {
		msgWnd.Clear()
		msgWndHideTime = time.UnixMilli(0)
	}
}

func delayedMessageHider() {
	msgWndShowCond = sync.NewCond(&NcursesMu)

	for {
		NcursesMu.Lock()
		if msgWndHideTime.Before(time.Now()) {
			doHideMessage()
			msgWndShowCond.Wait()
			NcursesMu.Unlock()
		} else {
			NcursesMu.Unlock()
			time.Sleep(time.Second)
		}
	}
}

func handleEvents(chub *chubby.Chubby, done chan<- any) {
	var ticker *time.Ticker

	events, err := chub.Events(true)
	if err != nil {
		return
	}

	s, err := chub.Status()
	if err != nil {
		return
	}

	NcursesMu.Lock()
	chubStatus = s
	chubStarted = time.Now().Unix() - int64(chubStatus.TrackPos)
	NcursesMu.Unlock()

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

func printErr(err error) {
	fmt.Fprintf(os.Stderr, "asp: %s\n", err.Error())
}
