package main

import (
	"path/filepath"
	"strconv"
	"strings"

	ncurses "github.com/gbin/goncurses"
	"github.com/vchimishuk/asp/config"
	"github.com/vchimishuk/asp/format"
	"github.com/vchimishuk/chubby"
)

// TODO: Move browserwindow.go in separate package?
type Item interface {
	// TODO: Create new type -- Path, which incapsulates
	//       full path and dir/file flag.
	//       Or... create this type inside shubby?
	IsDir() bool
	Path() string
	Format(width int) string
	IsSelected(val string) bool
}

// TODO: Rename to parentItem.
type ParentItem struct {
	path string
}

func (i *ParentItem) IsDir() bool {
	return true
}

func (i *ParentItem) Path() string {
	return i.path
}

func (i *ParentItem) Format(width int) string {
	if width >= 3 {
		return "../" + strings.Repeat(" ", width-3)
	} else {
		return ""
	}
}

func (i *ParentItem) IsSelected(val string) bool {
	return false
}

func newParentItem(path string) *ParentItem {
	return &ParentItem{path}
}

// type parentDirFormatter struct {
// }

// func (f *parentDirFormatter) Format(data map[string]string, width int) string {
// 	if width >= 3 {
// 		return "../" + strings.Repeat(" ", width-3)
// 	} else {
// 		return ""
// 	}
// }

type TrackItem struct {
	data map[string]string
	fmtr format.Formatter
	dir  bool
	path string
}

func newTrackItem(data map[string]string, fmtr format.Formatter,
	dir bool, path string) *TrackItem {

	return &TrackItem{data, fmtr, dir, path}
}

func (i *TrackItem) IsDir() bool {
	return i.dir
}

func (i *TrackItem) Path() string {
	return i.path
}

func (i *TrackItem) Format(width int) string {
	return i.fmtr.Format(i.data, width)
}

func (i *TrackItem) IsSelected(val string) bool {
	if i.dir {
		return strings.HasPrefix(val, i.path+"/")
	} else {
		return i.path == val
	}
}

type BrowserWindow struct {
	// Current folder path browser navigated in.
	path   string
	client *chubby.Chubby
	panel  *ncurses.Panel
	// title     *PanelWindow
	list      *ListWindow
	items     []chubby.Entry
	dirFmtr   format.Formatter
	trackFmtr format.Formatter
}

func NewBrowserWindow(client *chubby.Chubby, input func(string) string,
	h, w, y, x int) (*BrowserWindow, error) {

	root, err := ncurses.NewWindow(h, w, y, x)
	if err != nil {
		return nil, err
	}
	// title, err := NewPanelWindow(root, 1, w, 0, 0)
	// if err != nil {
	// 	return nil, err
	// }
	list, err := NewListWindow(root, input)
	if err != nil {
		return nil, err
	}

	return &BrowserWindow{
		path:   "",
		client: client,
		panel:  ncurses.NewPanel(root),
		// title:     title, // TODO: Add browser.title & playlist.title fmtrs config.
		list:      list,
		items:     nil,
		dirFmtr:   format.NewFormatter("{%n}/"),                 // TODO: Config.
		trackFmtr: format.NewFormatter("{-*%:%a - %t}{20%:%l}"), // TODO: Config.
	}, nil
}

func (w *BrowserWindow) Section() config.Section {
	return config.SectionBrowser
}

func (w *BrowserWindow) Command(cmd config.Command) error {
	switch cmd {
	case config.CommandApply:
		i := w.list.Cursor()
		if i != nil {
			w.apply(i.(Item))
		}
	case config.CommandApplyDir:
		i := w.list.Cursor()
		if i != nil {
			w.play(i.(Item))
		}
	case config.CommandBack:
		w.back()
	default:
		w.list.Command(cmd)
	}

	return nil // TODO:
}

// TODO: Rename to Activate()
func (w *BrowserWindow) Active() {
	w.panel.Top()
}

func (w *BrowserWindow) SetSelected(path string) {
	w.list.SetSelected(path)
}

func (w *BrowserWindow) Path() string {
	return w.path
}

func (w *BrowserWindow) SetPath(p string) {
	entries, err := w.client.List(p)
	if err != nil {
		// TODO: Print error in status and exit.
		panic(err)
	}

	items := make([]ListItem, 0, len(entries)+1)
	// items = append(items, newTrackItem(nil, &parentDirFormatter{}, true,
	// 	filepath.Dir(p)))
	items = append(items, newParentItem(filepath.Dir(p)))
	parent := -1

	for i, e := range entries {
		var data map[string]string
		var fmtr format.Formatter
		var path string

		if e.IsDir() {
			data = map[string]string{
				"p": e.Dir().Path,
				"n": e.Dir().Name,
			}
			fmtr = w.dirFmtr
			path = e.Dir().Path
		} else {
			data = map[string]string{
				"p": e.Track().Path,
				"a": e.Track().Artist,
				"b": e.Track().Album,
				"t": e.Track().Title,
				"n": strconv.Itoa(e.Track().Number),
				"l": e.Track().Length.String(),
			}
			fmtr = w.trackFmtr
			path = e.Track().Path
		}
		items = append(items, newTrackItem(data, fmtr, e.IsDir(), path))

		if isParent(w.path, path) {
			parent = i + 1 // Thre first one is "..".
		}
	}

	w.list.Clear()
	w.list.Add(items...)
	// w.title.SetText(p)

	if parent != -1 {
		w.list.SetCursor(parent)
	}
	w.path = p
}

func (w *BrowserWindow) apply(item Item) { // TODO: error?
	if item.IsDir() {
		w.SetPath(item.Path())
	} else {
		err := w.play(item)
		if err != nil {
			// TODO:
			panic(err)
		}
	}
}

func (w *BrowserWindow) play(item Item) error {
	return w.client.Play(item.Path())
}

func (w *BrowserWindow) back() { // TODO: error?
	w.SetPath(filepath.Dir(w.path))
}

func isParent(path, parent string) bool {
	return strings.HasPrefix(path, parent)
}
