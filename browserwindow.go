package main

import (
	"path"
	"path/filepath"
	"strconv"
	"strings"

	ncurses "github.com/gbin/goncurses"
	"github.com/vchimishuk/asp/config"
	"github.com/vchimishuk/asp/format"
	"github.com/vchimishuk/chubby"
)

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

func NewBrowserWindow(client *chubby.Chubby,
	h, w, y, x int) (*BrowserWindow, error) {

	root, err := ncurses.NewWindow(h, w, y, x)
	if err != nil {
		return nil, err
	}
	// title, err := NewPanelWindow(root, 1, w, 0, 0)
	// if err != nil {
	// 	return nil, err
	// }
	list, err := NewListWindow(root)
	if err != nil {
		return nil, err
	}

	return &BrowserWindow{
		path:      "",
		client:    client,
		panel:     ncurses.NewPanel(root),
		list:      list,
		items:     nil,
		dirFmtr:   format.NewFormatter(config.FormatBrowserDir),
		trackFmtr: format.NewFormatter(config.FormatBrowserTrack),
	}, nil
}

func (w *BrowserWindow) Command(cmd config.Cmd) error {
	switch cmd {
	case config.CmdApply:
		i := w.list.Cursor()
		if i != nil {
			w.apply(i.(Item))
		}
	case config.CmdPlay:
		i := w.list.Cursor()
		if i != nil {
			w.play(i.(Item))
		}
	case config.CmdBack:
		w.back()
	case config.CmdSelected:
		w.selected()
	default:
		w.list.Command(cmd)
	}

	return nil // TODO:
}

func (w *BrowserWindow) SetSelected(path string) {
	w.list.SetSelected(path)
}

func (w *BrowserWindow) Path() string {
	return w.path
}

func (w *BrowserWindow) SetPath(p string) error {
	entries, err := w.client.List(p)
	if err != nil {
		return err
	}
	w.items = entries

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

	return nil
}

func (w *BrowserWindow) Search(text string) {
	w.list.Search(text)
}

func (w *BrowserWindow) SearchNext() {
	w.list.SearchNext()
}

func (w *BrowserWindow) SearchPrev() {
	w.list.SearchPrev()
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

func (w *BrowserWindow) selected() {
	p := w.list.Selected()
	if p == "" {
		return
	}

	w.SetPath(path.Dir(p))
	for i, _ := range w.items {
		it := w.items[i]
		if it.IsDir() {
			continue
		}
		if it.Track().Path == p {
			w.list.SetCursor(i + 1)
		}
	}
}

func isParent(path, parent string) bool {
	return strings.HasPrefix(path, parent)
}
