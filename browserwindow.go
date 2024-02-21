package main

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vchimishuk/asp/config"
	"github.com/vchimishuk/asp/format"
	"github.com/vchimishuk/chubby"
)

type item struct {
	entry chubby.Entry
	data  map[string]string
	fmtr  format.Formatter
}

func newItem(entry chubby.Entry, data map[string]string,
	fmtr format.Formatter) *item {

	return &item{entry, data, fmtr}
}

func (i *item) Format(width int) string {
	return i.fmtr.Format(i.data, width)
}

func (i *item) IsActive(val string) bool {
	if i.entry.IsDir() {
		d := i.entry.Dir()
		// TODO: Slash suffix should be done using formatter.
		return d.Name != ".." && strings.HasPrefix(val, d.Path+"/")
	} else {
		return i.entry.Track().Path == val
	}
}

type BrowserWindow struct {
	path      string
	list      *ListWindow
	items     []chubby.Entry
	dirFmtr   format.Formatter
	trackFmtr format.Formatter
}

func NewBrowserWindow(h, w, y, x int) (*BrowserWindow, error) {
	list, err := NewListWindow(h, w, y, x)
	return &BrowserWindow{
		path:      "",
		list:      list,
		items:     nil,
		dirFmtr:   format.NewFormatter(config.FormatBrowserDir),
		trackFmtr: format.NewFormatter(config.FormatBrowserTrack),
	}, err
}

func (w *BrowserWindow) Cursor() chubby.Entry {
	return w.list.Cursor().(*item).entry
}

func (w *BrowserWindow) SetActive(path string) {
	w.list.SetActive(path)
}

func (w *BrowserWindow) ShowActive() {
	w.list.ShowActive()
}

func (w *BrowserWindow) SetDir(p string, entries []chubby.Entry) error {
	items := make([]ListItem, 0, len(entries)+1)
	pd := &chubby.Dir{
		Path: filepath.Dir(p),
		Name: "..",
	}
	items = append(items, newItem(pd, map[string]string{
		"p": pd.Path,
		"n": "..",
	}, w.dirFmtr))
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
		items = append(items, newItem(e, data, fmtr))

		if isParent(w.path, path) {
			parent = i + 1 // Thre first one is "..".
		}
	}

	w.list.Clear()
	w.list.Add(items...)

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

func (w *BrowserWindow) Up() {
	w.list.Up()
}

func (w *BrowserWindow) Down() {
	w.list.Down()
}

func (w *BrowserWindow) PageUp() {
	w.list.PageUp()
}

func (w *BrowserWindow) PageDown() {
	w.list.PageDown()
}

func (w *BrowserWindow) Home() {
	w.list.Home()
}

func (w *BrowserWindow) End() {
	w.list.End()
}

func isParent(path, parent string) bool {
	return strings.HasPrefix(path, parent)
}

func (w *BrowserWindow) Delete() {
	w.list.Delete()
}
